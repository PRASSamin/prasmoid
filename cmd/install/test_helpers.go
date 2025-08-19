package install

import (
	"os"
	"path/filepath"
	"testing"

	initCmd "github.com/PRASSamin/prasmoid/cmd/init"
)

// setupTestEnvironment creates a temporary project and a temporary home directory.
func SetupTestEnvironment(t *testing.T) (projectDir, homeDir string, cleanup func()) {
	projectDir, projectCleanup := initCmd.SetupTestProject(t)

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
