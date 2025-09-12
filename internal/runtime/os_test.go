package runtime

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

func TestOSModule(t *testing.T) {
	vm := NewRuntime()
	_, err := vm.RunString(`const os = require('os');`)
	require.NoError(t, err)

	t.Run("arch", func(t *testing.T) {
		script := `os.arch();`
		val, err := vm.RunString(script)
		require.NoError(t, err)
		require.NotEmpty(t, val.String())
		require.Equal(t, runtime.GOARCH, val.String())
	})

	t.Run("platform", func(t *testing.T) {
		script := `os.platform();`
		val, err := vm.RunString(script)
		require.NoError(t, err)
		require.NotEmpty(t, val.String())
		require.Equal(t, runtime.GOOS, val.String())
	})

	t.Run("release", func(t *testing.T) {
		script := `os.release();`
		val, err := vm.RunString(script)
		require.NoError(t, err)
		require.NotEmpty(t, val.String())

		uname := syscall.Utsname{}
		err = syscall.Uname(&uname)
		require.NoError(t, err)
		release := ""
		for _, c := range uname.Release {
			if c == 0 {
				break
			}
			release += string(byte(c))
		}
		require.Equal(t, release, val.String())
	})

	t.Run("type", func(t *testing.T) {
		script := `os.type();`
		val, err := vm.RunString(script)
		require.NoError(t, err)
		require.NotEmpty(t, val.String())
		require.Equal(t, "Linux", val.String())
	})

	t.Run("homedir", func(t *testing.T) {
		script := `os.homedir();`
		val, err := vm.RunString(script)
		require.NoError(t, err)
		require.NotEmpty(t, val.String())

		home, err := os.UserHomeDir()
		require.NoError(t, err)
		require.Equal(t, home, val.String())
	})

	t.Run("hostname", func(t *testing.T) {
		script := `os.hostname();`
		val, err := vm.RunString(script)
		require.NoError(t, err)
		require.NotEmpty(t, val.String())

		name, err := os.Hostname()
		require.NoError(t, err)
		require.Equal(t, name, val.String())
	})

	t.Run("tmpdir", func(t *testing.T) {
		script := `os.tmpdir();`
		val, err := vm.RunString(script)
		require.NoError(t, err)
		require.NotEmpty(t, val.String())
		require.Equal(t, os.TempDir(), val.String())
	})

	t.Run("uptime", func(t *testing.T) {
		script := `os.uptime();`
		val, err := vm.RunString(script)
		require.NoError(t, err)

		if runtime.GOOS == "linux" {
			require.Greater(t, val.ToFloat(), float64(0))
			raw, _ := os.ReadFile("/proc/uptime")
			uptime, _ := strconv.ParseFloat(strings.Fields(string(raw))[0], 64)
			require.Equal(t, uptime, val.ToFloat())
		} else {
			require.Equal(t, "Error: uptime not supported on "+runtime.GOOS, val.String())
		}
	})

	t.Run("freemem", func(t *testing.T) {
		script := `os.freemem();`
		val, err := vm.RunString(script)
		require.NoError(t, err)

		if runtime.GOOS == "linux" {
			require.Greater(t, val.ToInteger(), int64(0))
			var mem syscall.Sysinfo_t
			require.NoError(t, syscall.Sysinfo(&mem))
			require.Equal(t, int64(mem.Freeram*uint64(mem.Unit)), val.ToInteger())
		} else {
			require.Equal(t, "Error: freemem not supported on "+runtime.GOOS, val.String())
		}
	})

	t.Run("totalmem", func(t *testing.T) {
		script := `os.totalmem();`
		val, err := vm.RunString(script)
		require.NoError(t, err)

		if runtime.GOOS == "linux" {
			require.Greater(t, val.ToInteger(), int64(0))
			var mem syscall.Sysinfo_t
			require.NoError(t, syscall.Sysinfo(&mem))
			require.Equal(t, int64(mem.Totalram*uint64(mem.Unit)), val.ToInteger())
		} else {
			require.Equal(t, "Error: totalmem not supported on "+runtime.GOOS, val.String())
		}
	})

	t.Run("loadavg", func(t *testing.T) {
		script := `os.loadavg();`
		val, err := vm.RunString(script)
		require.NoError(t, err)

		if runtime.GOOS == "linux" {
			arr := val.Export().([]float64)
			require.Len(t, arr, 3)
		} else {
			require.Equal(t, "Error: loadavg not supported on "+runtime.GOOS, val.String())
		}
	})

	t.Run("endianness", func(t *testing.T) {
		script := `os.endianness();`
		val, err := vm.RunString(script)
		require.NoError(t, err)
		require.Contains(t, []string{"LE", "BE"}, val.String())
	})

	t.Run("availableParallelism", func(t *testing.T) {
		script := `os.availableParallelism();`
		val, err := vm.RunString(script)
		require.NoError(t, err)
		require.Greater(t, val.ToInteger(), int64(0))
		require.Equal(t, int64(runtime.NumCPU()), val.ToInteger())
	})

	t.Run("machine", func(t *testing.T) {
		script := `os.machine();`
		val, err := vm.RunString(script)
		require.NoError(t, err)
		require.NotEmpty(t, val.String())

		var uname unix.Utsname
		require.NoError(t, unix.Uname(&uname))
		var machine []byte
		for _, c := range uname.Machine {
			if c == 0 {
				break
			}
			machine = append(machine, byte(c))
		}
		require.Equal(t, string(machine), val.String())
	})

	t.Run("userInfo", func(t *testing.T) {
		script := `os.userInfo();`
		val, err := vm.RunString(script)
		require.NoError(t, err)
		obj := val.ToObject(vm)
		require.NotEmpty(t, obj.Get("uid").String())
		require.NotEmpty(t, obj.Get("gid").String())
		require.NotEmpty(t, obj.Get("username").String())
		require.NotEmpty(t, obj.Get("homedir").String())

		t.Run("empty SHELL env", func(t *testing.T) {
			oldShell := os.Getenv("SHELL")
			_ = os.Unsetenv("SHELL")
			defer func() { _ = os.Setenv("SHELL", oldShell) }()

			script := `os.userInfo();`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			obj := val.ToObject(vm)
			require.Equal(t, "unknown", obj.Get("shell").String())
		})
	})

	t.Run("not implemented functions", func(t *testing.T) {
		notImplementedFuncs := []string{
			"cpus", "networkInterfaces", "setPriority", "getPriority", "version",
		}

		for _, name := range notImplementedFuncs {
			t.Run(name, func(t *testing.T) {
				script := fmt.Sprintf(`os.%s();`, name)
				val, err := vm.RunString(script)
				require.NoError(t, err)
				expected := fmt.Sprintf("os.%s is not implemented in this runtime", name)
				require.Equal(t, expected, val.String())
			})
		}
	})
}
