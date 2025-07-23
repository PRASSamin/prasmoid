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
	_proc := module.Get("exports").(*goja.Object)


	// process.exit(code)
	_proc.Set("exit", func(call goja.FunctionCall) goja.Value {
		code := 0
		if len(call.Arguments) > 0 {
			code = int(call.Arguments[0].ToInteger())
		}
		os.Exit(code)
		return goja.Undefined()
	})

	// process.cwd()
	_proc.Set("cwd", func(call goja.FunctionCall) goja.Value {
		dir, err := os.Getwd()
		if err != nil {
			return vm.ToValue(fmt.Sprintf("cwd error: %v", err))
		}
		return vm.ToValue(dir)
	})

	// process.chdir(path)
	_proc.Set("chdir", func(call goja.FunctionCall) goja.Value {
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
	_proc.Set("uptime", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(time.Since(startTime).Seconds())
	})

	// process.memoryUsage()
	_proc.Set("memoryUsage", func(call goja.FunctionCall) goja.Value {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		obj := vm.NewObject()
		obj.Set("rss", m.Sys)
		obj.Set("heapTotal", m.HeapSys)
		obj.Set("heapUsed", m.HeapAlloc)
		obj.Set("external", m.StackSys)
		return obj
	})

	// process.kill(pid)
	_proc.Set("kill", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 1 {
			return vm.ToValue("kill: pid required")
		}
		pid := int(call.Arguments[0].ToInteger())
		err := syscall.Kill(pid, syscall.SIGTERM)
		if err != nil {
			return vm.ToValue(fmt.Sprintf("kill error: %v", err))
		}
		return goja.Undefined()
	})

	// process.getuid()
	_proc.Set("getuid", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(os.Getuid())
	})

	// process.getgid()
	_proc.Set("getgid", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(os.Getgid())
	})

	// process.geteuid()
	_proc.Set("geteuid", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(os.Geteuid())
	})

	// process.getegid()
	_proc.Set("getegid", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(os.Getegid())
	})

	// process.env
	type Process struct {
		env map[string]string
	}
	
	envs := LoadEnvWithPrefix()
	p := &Process{env: envs}

	_proc.Set("env", p.env)


	// process.nextTick(fn)
	_proc.Set("nextTick", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.ToValue("nextTick: missing function")
		}
		fn, ok := goja.AssertFunction(call.Arguments[0])
		if !ok {
			return vm.ToValue("nextTick: argument must be function")
		}
		go func() {
			_, err := fn(goja.Undefined())
			if err != nil {
				fmt.Println("nextTick error:", err)
			}
		}()
		return goja.Undefined()
	})

	// === NOT IMPLEMENTED FUNCTIONS ===

	notImplemented := func(name string) func(goja.FunctionCall) goja.Value {
		return func(call goja.FunctionCall) goja.Value {
			return vm.ToValue(fmt.Sprintf("process.%s is not implemented in this runtime", name))
		}
	}

	notImplList := []string{
		"binding", "dlopen", "getActiveResourcesInfo", "reallyExit", "loadEnvFile",
		"cpuUsage", "resourceUsage", "constrainedMemory", "availableMemory", "execve",
		"ref", "unref", "hrtime", "openStdin", "getgroups", "assert",
		"setUncaughtExceptionCaptureCallback", "hasUncaughtExceptionCaptureCallback",
		"emitWarning", "setSourceMapsEnabled", "getBuiltinModule", "abort",
		"initgroups", "setgroups", "setegid", "seteuid", "setgid", "setuid",
	}

	for _, name := range notImplList {
		_proc.Set(name, notImplemented(name))
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
