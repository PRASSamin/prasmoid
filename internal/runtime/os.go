package runtime

import (
	"fmt"
	"os"
	"os/user"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"github.com/dop251/goja"
	"golang.org/x/sys/unix"
)

func OS(vm *goja.Runtime, module *goja.Object) {
	_os := module.Get("exports").(*goja.Object)

	// arch()
	if err := _os.Set("arch", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(runtime.GOARCH)
	}); err != nil {
		fmt.Printf("Error setting os.arch: %v\n", err)
	}

	// platform()
	if err := _os.Set("platform", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(runtime.GOOS)
	}); err != nil {
		fmt.Printf("Error setting os.platform: %v\n", err)
	}

	// release() — kernel version
	if err := _os.Set("release", func(call goja.FunctionCall) goja.Value {
		var uname syscall.Utsname
		err := syscall.Uname(&uname)
		if err != nil {
			return vm.ToValue("unknown")
		}
		var release string
		for _, c := range uname.Release {
			if c == 0 {
				break
			}
			release += string(byte(c))
		}
		return vm.ToValue(release)
	}); err != nil {
		fmt.Printf("Error setting os.release: %v\n", err)
	}

	// type()
	if err := _os.Set("type", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue("Linux") // Or adjust based on runtime.GOOS
	}); err != nil {
		fmt.Printf("Error setting os.type: %v\n", err)
	}

	// homedir()
	if err := _os.Set("homedir", func(call goja.FunctionCall) goja.Value {
		home, _ := os.UserHomeDir()
		return vm.ToValue(home)
	}); err != nil {
		fmt.Printf("Error setting os.homedir: %v\n", err)
	}

	// hostname()
	if err := _os.Set("hostname", func(call goja.FunctionCall) goja.Value {
		name, err := os.Hostname()
		if err != nil {
			return vm.ToValue("unknown")
		}
		return vm.ToValue(name)
	}); err != nil {
		fmt.Printf("Error setting os.hostname: %v\n", err)
	}

	// tmpdir()
	if err := _os.Set("tmpdir", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(os.TempDir())
	}); err != nil {
		fmt.Printf("Error setting os.tmpdir: %v\n", err)
	}

	// uptime()
	if err := _os.Set("uptime", func(call goja.FunctionCall) goja.Value {
		if runtime.GOOS == "linux" {
			data, _ := os.ReadFile("/proc/uptime")
			uptime, _ := strconv.ParseFloat(strings.Fields(string(data))[0], 64)
			return vm.ToValue(uptime)
		} else {
			return vm.ToValue("Error: uptime not supported on " + runtime.GOOS)
		}
	}); err != nil {
		fmt.Printf("Error setting os.uptime: %v\n", err)
	}

	// freemem() and totalmem()
	if err := _os.Set("freemem", func(call goja.FunctionCall) goja.Value {
		var mem syscall.Sysinfo_t
		if err := syscall.Sysinfo(&mem); err != nil {
			return vm.ToValue("Error: freemem not supported on " + runtime.GOOS)
		}
		return vm.ToValue(mem.Freeram * uint64(mem.Unit))
	}); err != nil {
		fmt.Printf("Error setting os.freemem: %v\n", err)
	}

	if err := _os.Set("totalmem", func(call goja.FunctionCall) goja.Value {
		var mem syscall.Sysinfo_t
		if err := syscall.Sysinfo(&mem); err != nil {
			return vm.ToValue("Error: totalmem not supported on " + runtime.GOOS)
		}
		return vm.ToValue(mem.Totalram * uint64(mem.Unit))
	}); err != nil {
		fmt.Printf("Error setting os.totalmem: %v\n", err)
	}

	// loadavg()
	if err := _os.Set("loadavg", func(call goja.FunctionCall) goja.Value {
		var info syscall.Sysinfo_t
		if err := syscall.Sysinfo(&info); err != nil {
			return vm.ToValue("Error: loadavg not supported on " + runtime.GOOS)
		}
		return vm.ToValue([]float64{
			float64(info.Loads[0]) / 65536.0,
			float64(info.Loads[1]) / 65536.0,
			float64(info.Loads[2]) / 65536.0,
		})
	}); err != nil {
		fmt.Printf("Error setting os.loadavg: %v\n", err)
	}

	// endianness()
	if err := _os.Set("endianness", func(call goja.FunctionCall) goja.Value {
		var i int32 = 0x01020304
		u := (*[4]byte)(unsafe.Pointer(&i))
		if u[0] == 0x04 {
			return vm.ToValue("LE")
		}
		return vm.ToValue("BE")
	}); err != nil {
		fmt.Printf("Error setting os.endianness: %v\n", err)
	}

	// availableParallelism()
	if err := _os.Set("availableParallelism", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(runtime.NumCPU())
	}); err != nil {
		fmt.Printf("Error setting os.availableParallelism: %v\n", err)
	}

	// machine()
	if err := _os.Set("machine", func(call goja.FunctionCall) goja.Value {
		var uname unix.Utsname
		if err := unix.Uname(&uname); err != nil {
			return vm.ToValue("unknown")
		}
		var machine []byte
		for _, c := range uname.Machine {
			if c == 0 {
				break
			}
			machine = append(machine, byte(c))
		}
		return vm.ToValue(string(machine))
	}); err != nil {
		fmt.Printf("Error setting os.machine: %v\n", err)
	}

	// userInfo()
	if err := _os.Set("userInfo", func(call goja.FunctionCall) goja.Value {
		user, err := user.Current()
		shellPath := os.Getenv("SHELL")
		if shellPath == "" {
			shellPath = "unknown"
		}
		if err != nil {
			return vm.ToValue("unknown")
		}
		obj := vm.NewObject()
		SetObjProperty(obj, "uid", user.Uid)
		SetObjProperty(obj, "gid", user.Gid)
		SetObjProperty(obj, "name", user.Name)
		SetObjProperty(obj, "username", user.Username)
		SetObjProperty(obj, "homedir", user.HomeDir)
		SetObjProperty(obj, "shell", shellPath)
		return obj
	}); err != nil {
		fmt.Printf("Error setting os.userInfo: %v\n", err)
	}

	// === NOT IMPLEMENTED FUNCTIONS ===

	notImplemented := func(name string) func(goja.FunctionCall) goja.Value {
		return func(call goja.FunctionCall) goja.Value {
			return vm.ToValue(fmt.Sprintf("os.%s is not implemented in this runtime", name))
		}
	}

	notImplList := []string{
		"cpus", "networkInterfaces", "setPriority", "getPriority", "version",
	}

	for _, name := range notImplList {
		if err := _os.Set(name, notImplemented(name)); err != nil {
			fmt.Printf("Error setting os.%s: %v\n", name, err)
		}
	}

}
