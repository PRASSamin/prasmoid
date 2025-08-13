package runtime_tests

import (
	"os"
	"testing"

	rtime "github.com/PRASSamin/prasmoid/internal/runtime"
)

func TestProcessModule(t *testing.T) {
	vm := rtime.NewRuntime()
	_, err := vm.RunString(`const process = require('process');`)
	if err != nil {
		t.Fatalf("Failed to declare process: %v", err)
	}

	t.Run("cwd", func(t *testing.T) {
		script := `process.cwd();`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		wd, _ := os.Getwd()
		if val.String() != wd {
			t.Errorf("Expected cwd %s, got %s", wd, val.String())
		}
	})

	t.Run("chdir", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "chdir-test-")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer func() {
			if err := os.RemoveAll(tmpDir); err != nil {
				t.Errorf("Failed to remove temporary directory: %v", err)
			}
		}()

		script := `process.chdir('` + tmpDir + `');`
		_, err = vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		wd, _ := os.Getwd()
		if wd != tmpDir {
			t.Errorf("Expected working directory %s, got %s", tmpDir, wd)
		}
	})

	t.Run("uptime", func(t *testing.T) {
		script := `process.uptime();`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		if val.ToFloat() <= 0 {
			t.Errorf("Expected uptime > 0, got %f", val.ToFloat())
		}
	})

	t.Run("memoryUsage", func(t *testing.T) {
		script := `process.memoryUsage();`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		obj := val.ToObject(vm)
		if obj.Get("rss").ToInteger() <= 0 ||
			obj.Get("heapTotal").ToInteger() <= 0 ||
			obj.Get("heapUsed").ToInteger() <= 0 ||
			obj.Get("external").ToInteger() <= 0 {
			t.Errorf("Memory usage values should be positive: %+v", obj.Export())
		}
	})

	t.Run("getuid", func(t *testing.T) {
		script := `process.getuid();`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		if val.ToInteger() != int64(os.Getuid()) {
			t.Errorf("Expected uid %d, got %d", os.Getuid(), val.ToInteger())
		}
	})

	t.Run("getgid", func(t *testing.T) {
		script := `process.getgid();`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		if val.ToInteger() != int64(os.Getgid()) {
			t.Errorf("Expected gid %d, got %d", os.Getgid(), val.ToInteger())
		}
	})

	t.Run("geteuid", func(t *testing.T) {
		script := `process.geteuid();`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		if val.ToInteger() != int64(os.Geteuid()) {
			t.Errorf("Expected euid %d, got %d", os.Geteuid(), val.ToInteger())
		}
	})

	t.Run("getegid", func(t *testing.T) {
		script := `process.getegid();`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		if val.ToInteger() != int64(os.Getegid()) {
			t.Errorf("Expected egid %d, got %d", os.Getegid(), val.ToInteger())
		}
	})

	t.Run("env", func(t *testing.T) {
		script := `process.env.HOME;`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		if val.String() != os.Getenv("HOME") {
			t.Errorf("Expected HOME env var %s, got %s", os.Getenv("HOME"), val.String())
		}
	})
}
