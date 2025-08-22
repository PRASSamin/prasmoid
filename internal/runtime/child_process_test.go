package runtime

import (
	"fmt"
	"testing"

)

func TestChildProcessModule(t *testing.T) {
	vm := NewRuntime()
	_, err := vm.RunString(`const child_process = require('child_process');`)
	if err != nil {
		t.Fatalf("Failed to declare child_process: %v", err)
	}

	t.Run("execSync - valid command", func(t *testing.T) {
		script := `child_process.execSync('echo hello');`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		if val.String() != "hello\n" {
			t.Errorf("Expected 'hello\n', but got '%s'", val.String())
		}
	})

	t.Run("execSync - invalid command", func(t *testing.T) {
		script := `child_process.execSync('nonexistent_command');`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		// Expecting an error message from the Go side
		if _, ok := val.Export().(string); !ok || !val.ToBoolean() {
			t.Errorf("Expected an error string, but got %v", val.Export())
		}
	})

	t.Run("execSync - no command", func(t *testing.T) {
		script := `child_process.execSync('');`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		if val.String() != "No command given" {
			t.Errorf("Expected 'No command given', but got '%s'", val.String())
		}
	})

	t.Run("exec - not implemented", func(t *testing.T) {
		script := `child_process.exec('ls');`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		if val.String() != "child_process.exec is not implemented in this runtime, use execSync instead" {
			t.Errorf("Expected not implemented message, but got '%s'", val.String())
		}
	})

	t.Run("not implemented functions", func(t *testing.T) {
		notImplementedFuncs := []string{
			"execFileSync",
			"execFile",
			"spawn",
			"spawnSync",
		}

		for _, name := range notImplementedFuncs {
			t.Run(name, func(t *testing.T) {
				script := fmt.Sprintf(`child_process.%s();`, name)
				val, err := vm.RunString(script)
				if err != nil {
					t.Fatalf("vm.RunString() for %s failed: %v", name, err)
				}
				expected := fmt.Sprintf("child_process.%s is not implemented in this runtime", name)
				if val.String() != expected {
					t.Errorf("Expected '%s', but got '%s'", expected, val.String())
				}
			})
		}
	})
}