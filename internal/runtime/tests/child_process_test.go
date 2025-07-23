package runtime_tests

import (
	"testing"

	rtime "github.com/PRASSamin/prasmoid/internal/runtime"
)

func TestChildProcessModule(t *testing.T) {
	vm := rtime.NewRuntime()
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
}
