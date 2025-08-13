package runtime_tests

import (
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"

	rtime "github.com/PRASSamin/prasmoid/internal/runtime"
	"golang.org/x/sys/unix"
)

func TestOSModule(t *testing.T) {
	vm := rtime.NewRuntime()
	_, err := vm.RunString(`const os = require('os');`)
	if err != nil {
		t.Fatalf("Failed to declare os: %v", err)
	}

	t.Run("arch", func(t *testing.T) {
		script := `os.arch();`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		if val.String() == "" {
			t.Errorf("Expected a non-empty string, got empty")
		}
		if val.String() != runtime.GOARCH {
			t.Errorf("Expected %s, got %s", runtime.GOARCH, val.String())
		}
	})

	t.Run("platform", func(t *testing.T) {
		script := `os.platform();`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		if val.String() == "" {
			t.Errorf("Expected a non-empty string, got empty")
		}
		if val.String() != runtime.GOOS {
			t.Errorf("Expected %s, got %s", runtime.GOOS, val.String())
		}
	})

	t.Run("release", func(t *testing.T) {
		script := `os.release();`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		if val.String() == "" {
			t.Errorf("Expected a non-empty string, got empty")
		}
		uname := syscall.Utsname{}
		err = syscall.Uname(&uname)
		if err != nil {
			t.Errorf("Failed to get kernel version: %v", err)
		}
		release := ""
		for _, c := range uname.Release {
			if c == 0 {
				break
			}
			release += string(byte(c))
		}
		if val.String() != release {
			t.Errorf("Expected %s, got %s", release, val.String())
		}
	})

	t.Run("type", func(t *testing.T) {
		script := `os.type();`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		if val.String() == "" {
			t.Errorf("Expected a non-empty string, got empty")
		}
		if val.String() != "Linux" {
			t.Errorf("Expected Linux, got %s", val.String())
		}
	})

	t.Run("homedir", func(t *testing.T) {
		script := `os.homedir();`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		if val.String() == "" {
			t.Errorf("Expected a non-empty string, got empty")
		}
		user, err := os.UserHomeDir()
		if err != nil {
			t.Errorf("Failed to get user home directory: %v", err)
		}
		if val.String() != user {
			t.Errorf("Expected %s, got %s", user, val.String())
		}
	})

	t.Run("hostname", func(t *testing.T) {
		script := `os.hostname();`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		if val.String() == "" {
			t.Errorf("Expected a non-empty string, got empty")
		}
		name, err := os.Hostname()
		if err != nil {
			t.Errorf("Failed to get hostname: %v", err)
		}
		if val.String() != name {
			t.Errorf("Expected %s, got %s", name, val.String())
		}
	})

	t.Run("tmpdir", func(t *testing.T) {
		script := `os.tmpdir();`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		if val.String() == "" {
			t.Errorf("Expected a non-empty string, got empty")
		}
		tmpdir := os.TempDir()
		if val.String() != tmpdir {
			t.Errorf("Expected %s, got %s", tmpdir, val.String())
		}
	})

	// Uptime is Linux-specific, so only test if on Linux
	// if runtime.GOOS == "linux" {
	t.Run("uptime", func(t *testing.T) {
		script := `os.uptime();`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		if val.ToFloat() <= 0 {
			t.Errorf("Expected uptime > 0, got %f", val.ToFloat())
		}
		raw, _ := os.ReadFile("/proc/uptime")
		uptime, _ := strconv.ParseFloat(strings.Fields(string(raw))[0], 64)
		if val.ToFloat() != uptime {
			t.Errorf("Expected %f, got %f", uptime, val.ToFloat())
		}
	})
	// }

	t.Run("freemem", func(t *testing.T) {
		script := `os.freemem();`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		if val.ToInteger() <= 0 {
			t.Errorf("Expected freemem > 0, got %d", val.ToInteger())
		}
		var mem syscall.Sysinfo_t
		err = syscall.Sysinfo(&mem)
		if err != nil {
			t.Errorf("Failed to get memory info: %v", err)
		}
		if val.ToInteger() != int64(mem.Freeram*uint64(mem.Unit)) {
			t.Errorf("Expected %d, got %d", mem.Freeram*uint64(mem.Unit), val.ToInteger())
		}
	})

	t.Run("totalmem", func(t *testing.T) {
		script := `os.totalmem();`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		if val.ToInteger() <= 0 {
			t.Errorf("Expected totalmem > 0, got %d", val.ToInteger())
		}
		var mem syscall.Sysinfo_t
		err = syscall.Sysinfo(&mem)
		if err != nil {
			t.Errorf("Failed to get memory info: %v", err)
		}
		if val.ToInteger() != int64(mem.Totalram*uint64(mem.Unit)) {
			t.Errorf("Expected %d, got %d", mem.Totalram*uint64(mem.Unit), val.ToInteger())
		}
	})

	t.Run("loadavg", func(t *testing.T) {
		script := `os.loadavg();`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		arr := val.Export().([]float64)
		if len(arr) != 3 {
			t.Errorf("Expected loadavg to return array of length 3, got %d", len(arr))
		}
	})

	t.Run("endianness", func(t *testing.T) {
		script := `os.endianness();`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		if val.String() != "LE" && val.String() != "BE" {
			t.Errorf("Expected endianness to be LE or BE, got %s", val.String())
		}
	})

	t.Run("availableParallelism", func(t *testing.T) {
		script := `os.availableParallelism();`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		if val.ToInteger() <= 0 {
			t.Errorf("Expected availableParallelism > 0, got %d", val.ToInteger())
		}
		if val.ToInteger() != int64(runtime.NumCPU()) {
			t.Errorf("Expected %d, got %d", runtime.NumCPU(), val.ToInteger())
		}
	})

	t.Run("machine", func(t *testing.T) {
		script := `os.machine();`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		if val.String() == "" {
			t.Errorf("Expected a non-empty string, got empty")
		}
		var uname unix.Utsname
		if err := unix.Uname(&uname); err != nil {
			t.Errorf("Failed to get machine info: %v", err)
		}
		var machine []byte
		for _, c := range uname.Machine {
			if c == 0 {
				break
			}
			machine = append(machine, byte(c))
		}
		if val.String() != string(machine) {
			t.Errorf("Expected %s, got %s", string(machine), val.String())
		}
	})

	t.Run("userInfo", func(t *testing.T) {
		script := `os.userInfo();`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		obj := val.ToObject(vm)
		if obj.Get("uid").String() == "" ||
			obj.Get("gid").String() == "" ||
			obj.Get("username").String() == "" ||
			obj.Get("homedir").String() == "" {
			t.Errorf("UserInfo result mismatch: %+v", obj.Export())
		}
	})
}
