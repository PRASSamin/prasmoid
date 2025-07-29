package cmd_tests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/PRASSamin/prasmoid/src/cmd"
)

func TestInitPlasmoid(t *testing.T) {
	// Create a temporary directory for the project
	projectParentDir, err := os.MkdirTemp("", "init-test-project-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(projectParentDir)

	// Create a temporary home directory for the symlink
	tmpHome, err := os.MkdirTemp("", "init-test-home-")
	if err != nil {
		t.Fatalf("Failed to create temp home dir: %v", err)
	}
	defer os.RemoveAll(tmpHome)

	// Set the HOME environment variable for the test
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)
	defer os.Setenv("HOME", originalHome)

	// Create the plasmoids directory in the temp home
	plasmoidsDir := filepath.Join(tmpHome, ".local/share/plasma/plasmoids")
	if err := os.MkdirAll(plasmoidsDir, 0755); err != nil {
		t.Fatalf("Failed to create plasmoids dir: %v", err)
	}

	// Create a ProjectConfig
	cmd.Config = cmd.ProjectConfig{
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
	os.Chdir(projectParentDir)
	defer os.Chdir(originalWd)

	// Run initPlasmoid
	if err := cmd.InitPlasmoid(); err != nil {
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

	projectPath := cmd.Config.Path
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

	var metadata cmd.Metadata
	if err := json.Unmarshal(data, &metadata); err != nil {
		t.Fatalf("Failed to unmarshal metadata.json: %v", err)
	}

	if metadata.KPlugin.Id != cmd.Config.ID {
		t.Errorf("Expected metadata ID '%s', but got '%s'", cmd.Config.ID, metadata.KPlugin.Id)
	}
	if metadata.KPlugin.Name != cmd.Config.Name {
		t.Errorf("Expected metadata Name '%s', but got '%s'", cmd.Config.Name, metadata.KPlugin.Name)
	}

	// Verify symlink was created
	symlinkPath := filepath.Join(plasmoidsDir, cmd.Config.ID)
	if _, err := os.Lstat(symlinkPath); os.IsNotExist(err) {
		t.Errorf("Expected symlink '%s' to be created, but it was not", symlinkPath)
	}
}
