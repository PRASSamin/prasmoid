package tests

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"text/template"

	initCmd "github.com/PRASSamin/prasmoid/cmd/init"
)

// SetupTestProject creates a temporary directory with a dummy metadata.json file.
// It returns the path to the temporary directory and a cleanup function.
func SetupTestProject(t *testing.T) (string, func()) {
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

	for relPath, content := range initCmd.FileTemplates {
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
		if err := tmpl.Execute(&buf, initCmd.Config); err != nil {
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


// setupTestEnvironment creates a temporary project and a temporary home directory.
func SetupTestEnvironment(t *testing.T) (projectDir, homeDir string, cleanup func()) {
	projectDir, projectCleanup := SetupTestProject(t)

	tmpHome, err := os.MkdirTemp("", "test-home-")
	if err != nil {
		t.Fatalf("Failed to create temp home dir: %v", err)
	}

	originalHome := os.Getenv("HOME")
	if err := os.Setenv("HOME", tmpHome); err != nil {
		t.Fatalf("Failed to set HOME environment variable: %v", err)
	}

	plasmoidsDir := filepath.Join(tmpHome, ".local/share/plasma/plasmoids")
	if err := os.MkdirAll(plasmoidsDir, 0755); err != nil {
		t.Fatalf("Failed to create plasmoids dir: %v", err)
	}

	cleanup = func() {
		if err := os.Setenv("HOME", originalHome); err != nil {
			t.Errorf("Failed to restore HOME environment variable: %v", err)
		}
		if err := os.RemoveAll(tmpHome); err != nil {
			t.Errorf("Failed to remove temporary home directory: %v", err)
		}
		projectCleanup()
	}

	return projectDir, tmpHome, cleanup
}
