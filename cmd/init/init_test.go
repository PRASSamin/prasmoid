package init

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestInitPlasmoid(t *testing.T) {
	// Create a temporary directory for the project
	projectParentDir, err := os.MkdirTemp("", "init-test-project-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(projectParentDir); err != nil {
			t.Errorf("Failed to remove project dir: %v", err)
		}
	}()

	// Create a temporary home directory for the symlink
	tmpHome, err := os.MkdirTemp("", "init-test-home-")
	if err != nil {
		t.Fatalf("Failed to create temp home dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpHome); err != nil {
			t.Errorf("Failed to remove temp home dir: %v", err)
		}
	}()

	// Set the HOME environment variable for the test
	originalHome := os.Getenv("HOME")
	if err := os.Setenv("HOME", tmpHome); err != nil {
		t.Fatalf("Failed to set HOME environment variable: %v", err)
	}
	defer func() {
		if err := os.Setenv("HOME", originalHome); err != nil {
			t.Errorf("Failed to restore HOME environment variable: %v", err)
		}
	}()

	// Create the plasmoids directory in the temp home
	plasmoidsDir := filepath.Join(tmpHome, ".local/share/plasma/plasmoids")
	if err := os.MkdirAll(plasmoidsDir, 0755); err != nil {
		t.Fatalf("Failed to create plasmoids dir: %v", err)
	}

	// Create a ProjectConfig
	Config = ProjectConfig{
		Name:        "TestPlasmoid",
		Path:        filepath.Join(projectParentDir, "TestPlasmoid"),
		ID:          "org.kde.testplasmoid",
		Description: "A test plasmoid",
		AuthorName:  "Test Author",
		AuthorEmail: "test@example.com",
		License:     "MIT",
		InitGit:     false,
	}

	// Change to the project parent directory
	originalWd, _ := os.Getwd()
	if err := os.Chdir(projectParentDir); err != nil {
		t.Fatalf("Failed to change directory to %s: %v", projectParentDir, err)
	}
	defer func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("Failed to restore original directory: %v", err)
		}
	}()

	// Run initPlasmoid
	if err := InitPlasmoid(); err != nil {
		t.Fatalf("initPlasmoid failed: %v", err)
	}

	// Verify files were created
	expectedFiles := []string{
		"metadata.json",
		"contents/ui/main.qml",
		"contents/config/main.xml",
		"contents/icons/prasmoid.svg",
		".gitignore",
		"prasmoid.config.js",
		".prasmoid/commands",
	}

	projectPath := Config.Path
	for _, file := range expectedFiles {
		path := filepath.Join(projectPath, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("Expected file or dir '%s' to be created, but it was not", path)
		}
	}

	// Verify metadata.json content
	metadataPath := filepath.Join(projectPath, "metadata.json")
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		t.Fatalf("Failed to read metadata.json: %v", err)
	}

	var metadata Metadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		t.Fatalf("Failed to unmarshal metadata.json: %v", err)
	}

	if metadata.KPlugin.Id != Config.ID {
		t.Errorf("Expected metadata ID '%s', but got '%s'", Config.ID, metadata.KPlugin.Id)
	}
	if metadata.KPlugin.Name != Config.Name {
		t.Errorf("Expected metadata Name '%s', but got '%s'", Config.Name, metadata.KPlugin.Name)
	}

	// Verify symlink was created
	symlinkPath := filepath.Join(plasmoidsDir, Config.ID)
	if _, err := os.Lstat(symlinkPath); os.IsNotExist(err) {
		t.Errorf("Expected symlink '%s' to be created, but it was not", symlinkPath)
	}
}
