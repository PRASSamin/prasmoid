package runtime_tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/PRASSamin/prasmoid/internal/runtime"
)

func TestPrasmoidModule(t *testing.T) {
	// Create a temporary directory for the test
	tmpDir, err := os.MkdirTemp("", "prasmoid-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Change to the temp directory for the duration of the test
	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalWd)

	// Create a dummy metadata.json
	metadataContent := `
	{
	  "KPlugin": {
	    "Id": "org.kde.testplasmoid",
	    "Version": "1.0.0",
	    "Name": "Test Plasmoid"
	  }
	}
	`
	if err := os.WriteFile(filepath.Join(tmpDir, "metadata.json"), []byte(metadataContent), 0644); err != nil {
		t.Fatalf("Failed to write metadata.json: %v", err)
	}

	vm := runtime.NewRuntime()
	_, err = vm.RunString(`const prasmoid = require('prasmoid');`)
	if err != nil {
		t.Fatalf("Failed to declare prasmoid: %v", err)
	}

	t.Run("getMetadata - existing key", func(t *testing.T) {
		script := `prasmoid.getMetadata('Id');`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		if val.String() != "org.kde.testplasmoid" {
			t.Errorf("Expected Id 'org.kde.testplasmoid', but got '%s'", val.String())
		}
	})

	t.Run("getMetadata - non-existing key", func(t *testing.T) {
		script := `prasmoid.getMetadata('NonExistent');`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		// Expecting an error message from the Go side
		if _, ok := val.Export().(string); !ok || !val.ToBoolean() {
			t.Errorf("Expected an error string, but got %v", val.Export())
		}
	})
}
