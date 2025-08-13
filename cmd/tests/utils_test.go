package cmd_tests

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/PRASSamin/prasmoid/cmd"
	"github.com/PRASSamin/prasmoid/utils"
)

// setupTestProject creates a temporary directory with a dummy metadata.json file.
// It returns the path to the temporary directory and a cleanup function.
func setupTestProject(t *testing.T) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "plasmoid-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create a dummy metadata.json
	metadata := map[string]interface{}{
		"KPlugin": map[string]interface{}{
			"Id":      "org.kde.testplasmoid",
			"Version": "1.0.0",
			"Name":    "Test Plasmoid",
		},
	}
	metadataPath := filepath.Join(tmpDir, "metadata.json")
	data, _ := json.MarshalIndent(metadata, "", "  ")
	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		t.Fatalf("Failed to write metadata.json: %v", err)
	}

	// Create a dummy contents directory
	if err := os.Mkdir(filepath.Join(tmpDir, "contents"), 0755); err != nil {
		t.Fatalf("Failed to create contents dir: %v", err)
	}

	for relPath, content := range cmd.FileTemplates {
		fullPath := filepath.Join(tmpDir, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Errorf("failed to create directory %s: %v", filepath.Dir(fullPath), err)
		}

		if _, err := os.Stat(fullPath); err == nil {
			t.Errorf("Skipping existing file: %s", relPath)
		}

		tmpl, err := template.New(relPath).Parse(content)
		if err != nil {
			t.Errorf("failed to parse template for %s: %v", relPath, err)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, cmd.Config); err != nil {
			t.Errorf("failed to execute template for %s: %v", relPath, err)
		}

		if err := os.WriteFile(fullPath, buf.Bytes(), 0644); err != nil {
			t.Errorf("failed to write file %s: %v", fullPath, err)
		}
	}

	// Change to the temp directory for the duration of the test
	originalWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory to %s: %v", tmpDir, err)
	}

	cleanup := func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("Failed to restore original directory: %v", err)
		}
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("Failed to remove temporary directory: %v", err)
		}
	}

	return tmpDir, cleanup
}

func TestGetDataFromMetadata(t *testing.T) {
	_, cleanup := setupTestProject(t)
	defer cleanup()

	t.Run("get existing key", func(t *testing.T) {
		id, err := utils.GetDataFromMetadata("Id")
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}
		if id != "org.kde.testplasmoid" {
			t.Errorf("Expected Id 'org.kde.testplasmoid', but got '%s'", id)
		}
	})

	t.Run("get non-existing key", func(t *testing.T) {
		_, err := utils.GetDataFromMetadata("NonExistent")
		if err == nil {
			t.Error("Expected an error for non-existing key, but got none")
		}
	})
}

func TestUpdateMetadata(t *testing.T) {
	tmpDir, cleanup := setupTestProject(t)
	defer cleanup()

	metadataPath := filepath.Join(tmpDir, "metadata.json")

	t.Run("update existing key", func(t *testing.T) {
		err := utils.UpdateMetadata("Version", "1.1.0")
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}

		data, _ := os.ReadFile(metadataPath)
		var meta map[string]map[string]interface{}
		if err := json.Unmarshal(data, &meta); err != nil {
			t.Fatalf("Failed to unmarshal metadata.json: %v", err)
		}

		if meta["KPlugin"]["Version"] != "1.1.0" {
			t.Errorf("Expected version '1.1.0', but got '%s'", meta["KPlugin"]["Version"])
		}
	})

	t.Run("add new key", func(t *testing.T) {
		err := utils.UpdateMetadata("NewKey", "NewValue")
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}

		data, _ := os.ReadFile(metadataPath)
		var meta map[string]map[string]interface{}
		if err := json.Unmarshal(data, &meta); err != nil {
			t.Fatalf("Failed to unmarshal metadata.json: %v", err)
		}

		if meta["KPlugin"]["NewKey"] != "NewValue" {
			t.Errorf("Expected NewKey 'NewValue', but got '%s'", meta["KPlugin"]["NewKey"])
		}
	})
}

func TestIsValidPlasmoid(t *testing.T) {
	tmpDir, cleanup := setupTestProject(t)
	defer cleanup()

	t.Run("valid plasmoid", func(t *testing.T) {
		if !utils.IsValidPlasmoid() {
			t.Error("Expected IsValidPlasmoid to be true, but it was false")
		}
	})

	t.Run("missing metadata.json", func(t *testing.T) {
		if err := os.Remove(filepath.Join(tmpDir, "metadata.json")); err != nil {
			t.Errorf("Failed to remove metadata.json: %v", err)
		}
		if utils.IsValidPlasmoid() {
			t.Error("Expected IsValidPlasmoid to be false, but it was true")
		}
	})

	t.Run("missing contents dir", func(t *testing.T) {
		if err := os.RemoveAll(filepath.Join(tmpDir, "contents")); err != nil {
			t.Errorf("Failed to remove contents dir: %v", err)
		}
		if utils.IsValidPlasmoid() {
			t.Error("Expected IsValidPlasmoid to be false, but it was true")
		}
	})
}
