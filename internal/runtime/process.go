package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/dop251/goja"
)

func Process(vm *goja.Runtime, module *goja.Object) {
	_process := module.Get("exports").(*goja.Object)

	// process.exit(code)
	if err := _process.Set("exit", func(call goja.FunctionCall) goja.Value {
		code := 0
		if len(call.Arguments) > 0 {
			code = int(call.Arguments[0].ToInteger())
		}
		os.Exit(code)
		return goja.Undefined()
	}); err != nil {
		fmt.Printf("Error setting process.exit: %v\n", err)
	}

	// process.cwd()
	if err := _process.Set("cwd", func(call goja.FunctionCall) goja.Value {
		dir, err := os.Getwd()
		if err != nil {
			return vm.ToValue(fmt.Sprintf("cwd error: %v", err))
		}
		return vm.ToValue(dir)
	}); err != nil {
		fmt.Printf("Error setting process.cwd: %v\n", err)
	}

	// process.chdir(path)
	if err := _process.Set("chdir", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.ToValue("chdir: path is required")
		}
		path := call.Arguments[0].String()
		err := os.Chdir(path)
		if err != nil {
			return vm.ToValue(fmt.Sprintf("chdir error: %v", err))
		}
		return goja.Undefined()
	}); err != nil {
		fmt.Printf("Error setting process.chdir: %v\n", err)
	}

	// process.uptime()
	if err := _process.Set("uptime", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(time.Since(startTime).Seconds())
	}); err != nil {
		fmt.Printf("Error setting process.uptime: %v\n", err)
	}

	// process.memoryUsage()
	if err := _process.Set("memoryUsage", func(call goja.FunctionCall) goja.Value {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		obj := vm.NewObject()
		SetObjProperty(obj, "rss", m.Sys)
		SetObjProperty(obj, "heapTotal", m.HeapSys)
		SetObjProperty(obj, "heapUsed", m.HeapAlloc)
		SetObjProperty(obj, "external", m.StackSys)
		return obj
	}); err != nil {
		fmt.Printf("Error setting process.memoryUsage: %v\n", err)
	}

	// process.kill(pid)
	if err := _process.Set("kill", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return vm.ToValue("kill: pid required")
		}
		pid := int(call.Arguments[0].ToInteger())
		err := syscall.Kill(pid, syscall.SIGTERM)
		if err != nil {
			return vm.ToValue(fmt.Sprintf("kill error: %v", err))
		}
		return goja.Undefined()
	}); err != nil {
		fmt.Printf("Error setting process.kill: %v\n", err)
	}

	// process.getuid()
	if err := _process.Set("getuid", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(os.Getuid())
	}); err != nil {
		fmt.Printf("Error setting process.getuid: %v\n", err)
	}

	// process.getgid()
	if err := _process.Set("getgid", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(os.Getgid())
	}); err != nil {
		fmt.Printf("Error setting process.getgid: %v\n", err)
	}

	// process.geteuid()
	if err := _process.Set("geteuid", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(os.Geteuid())
	}); err != nil {
		fmt.Printf("Error setting process.geteuid: %v\n", err)
	}

	// process.getegid()
	if err := _process.Set("getegid", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(os.Getegid())
	}); err != nil {
		fmt.Printf("Error setting process.getegid: %v\n", err)
	}

	// process.env
	type Process struct {
		env map[string]string
	}

	envs := LoadEnvWithPrefix()
	p := &Process{env: envs}

	if err := _process.Set("env", p.env); err != nil {
		fmt.Printf("Error setting process.env: %v\n", err)
	}

	// === NOT IMPLEMENTED FUNCTIONS ===

	notImplemented := func(name string) func(goja.FunctionCall) goja.Value {
		return func(call goja.FunctionCall) goja.Value {
			return vm.ToValue(fmt.Sprintf("process.%s is not implemented in this runtime", name))
		}
	}

	notImplList := []string{
		"nextTick", "binding", "dlopen", "getActiveResourcesInfo", "reallyExit", "loadEnvFile",
		"cpuUsage", "resourceUsage", "constrainedMemory", "availableMemory", "execve",
		"ref", "unref", "hrtime", "openStdin", "getgroups", "assert",
		"setUncaughtExceptionCaptureCallback", "hasUncaughtExceptionCaptureCallback",
		"emitWarning", "setSourceMapsEnabled", "getBuiltinModule", "abort",
		"initgroups", "setgroups", "setegid", "seteuid", "setgid", "setuid",
	}

	for _, name := range notImplList {
		if err := _process.Set(name, notImplemented(name)); err != nil {
			fmt.Printf("Error setting process.%s: %v\n", name, err)
		}
	}
}

var startTime = time.Now()

func LoadEnvWithPrefix() map[string]string {
	envs := make(map[string]string)
	envFiles := []string{
		".env.development.local",
		".env.local",
		".env.development",
		".env.production.local",
		".env.production",
		".env.test.local",
		".env.test",
		".env",
	}

	// Load from actual system env
	for _, e := range os.Environ() {
		envKeyValue := strings.SplitN(e, "=", 2)
		envs[envKeyValue[0]] = envKeyValue[1]
	}

	for _, file := range envFiles {
		path := filepath.Join(".", file)
		if _, err := os.Stat(path); err == nil {
			loadEnvFile(envs, path)
		}
	}

	return envs
}

func loadEnvFile(envMap map[string]string, filePath string) {
	content, _ := os.ReadFile(filePath)
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "PRASMOID_") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := parts[0]
				val := strings.Trim(parts[1], `"`)
				envMap[key] = val
			}
		}
	}
}
