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
	_ = _process.Set("exit", func(call goja.FunctionCall) goja.Value {
		code := 0
		if len(call.Arguments) > 0 {
			code = int(call.Arguments[0].ToInteger())
		}
		os.Exit(code)
		return goja.Undefined()
	})

	// process.cwd()
	_ = _process.Set("cwd", func(call goja.FunctionCall) goja.Value {
		dir, err := os.Getwd()
		if err != nil {
			return vm.ToValue(fmt.Sprintf("cwd error: %v", err))
		}
		return vm.ToValue(dir)
	})

	// process.chdir(path)
	_ = _process.Set("chdir", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.ToValue("chdir: path is required")
		}
		path := call.Arguments[0].String()
		err := os.Chdir(path)
		if err != nil {
			return vm.ToValue(fmt.Sprintf("chdir error: %v", err))
		}
		return goja.Undefined()
	})

	// process.uptime()
	_ = _process.Set("uptime", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(time.Since(startTime).Seconds())
	})

	// process.memoryUsage()
	_ = _process.Set("memoryUsage", func(call goja.FunctionCall) goja.Value {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		obj := vm.NewObject()
		SetObjProperty(obj, "rss", m.Sys)
		SetObjProperty(obj, "heapTotal", m.HeapSys)
		SetObjProperty(obj, "heapUsed", m.HeapAlloc)
		SetObjProperty(obj, "external", m.StackSys)
		return obj
	})

	// process.kill(pid)
	_ = _process.Set("kill", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return vm.ToValue("kill: pid required")
		}
		pid := call.Arguments[0].ToInteger()
		if pid <= 0 {
			// Prevent calling syscall.Kill with dangerous PIDs like -1 or 0
			return vm.ToValue(fmt.Sprintf("kill error: invalid pid %d", pid))
		}
		err := syscall.Kill(int(pid), syscall.SIGTERM) // Or whatever signal is passed
		if err != nil {
			return vm.ToValue(fmt.Sprintf("kill error: %v", err))
		}
		return goja.Undefined()
	})

	// process.getuid()
	_ = _process.Set("getuid", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(os.Getuid())
	})

	// process.getgid()
	_ = _process.Set("getgid", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(os.Getgid())
	})

	// process.geteuid()
	_ = _process.Set("geteuid", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(os.Geteuid())
	})

	// process.getegid()
	_ = _process.Set("getegid", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(os.Getegid())
	})

	// process.env
	type Process struct {
		env map[string]string
	}

	// Get current working directory for loading env files
	wd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current working directory for env files: %v\n", err)
		wd = "" // Fallback to empty string if error
	}
	envs := LoadEnvWithPrefix(wd)
	p := &Process{env: envs}

	_ = _process.Set("env", p.env)

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
		_ = _process.Set(name, notImplemented(name))
	}
}

var startTime = time.Now()

func LoadEnvWithPrefix(baseDir string) map[string]string {
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
		path := filepath.Join(baseDir, file)
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
