package cmd_tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/PRASSamin/prasmoid/cmd"
	"github.com/PRASSamin/prasmoid/utils"
)

// setupTestEnvironment creates a temporary project and a temporary home directory.
func setupTestEnvironment(t *testing.T) (projectDir, homeDir string, cleanup func()) {
	projectDir, projectCleanup := setupTestProject(t)

	tmpHome, err := os.MkdirTemp("", "test-home-")
	if err != nil {
		t.Fatalf("Failed to create temp home dir: %v", err)
	}

	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpHome)

	plasmoidsDir := filepath.Join(tmpHome, ".local/share/plasma/plasmoids")
	if err := os.MkdirAll(plasmoidsDir, 0755); err != nil {
		t.Fatalf("Failed to create plasmoids dir: %v", err)
	}

	cleanup = func() {
		os.Setenv("HOME", originalHome)
		os.RemoveAll(tmpHome)
		projectCleanup()
	}

	return projectDir, tmpHome, cleanup
}

func TestLinkAndUnlink(t *testing.T) {
	_, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	dest, err := utils.GetDevDest()
	if err != nil {
		t.Fatalf("GetDevDest() failed: %v", err)
	}

	// Test Link
	t.Run("link plasmoid", func(t *testing.T) {
		if err := cmd.LinkPlasmoid(dest); err != nil {
			t.Fatalf("LinkPlasmoid failed: %v", err)
		}

		if _, err := os.Lstat(dest); os.IsNotExist(err) {
			t.Errorf("Expected symlink '%s' to be created, but it was not", dest)
		}
	})

	// Test Unlink
	t.Run("unlink plasmoid", func(t *testing.T) {
		cmd.UnlinkCmd.Run(nil, []string{})

		if _, err := os.Lstat(dest); err == nil {
			t.Errorf("Expected symlink '%s' to be removed, but it still exists", dest)
		}
	})
}

func TestInstallAndUninstall(t *testing.T) {
	_, _, cleanup := setupTestEnvironment(t)
	defer cleanup()

	dest, err := utils.GetDevDest()
	if err != nil {
		t.Fatalf("GetDevDest() failed: %v", err)
	}

	// Test Install
	t.Run("install plasmoid", func(t *testing.T) {
		if err := cmd.InstallPlasmoid(); err != nil {
			t.Fatalf("InstallPlasmoid failed: %v", err)
		}

		// Verify it's a directory, not a symlink
		info, err := os.Stat(dest)
		if os.IsNotExist(err) {
			t.Fatalf("Expected install directory '%s' to be created, but it was not", dest)
		}
		if !info.IsDir() {
			t.Errorf("Expected '%s' to be a directory, but it is not", dest)
		}

		// Verify a file inside
		metadataPath := filepath.Join(dest, "metadata.json")
		if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
			t.Errorf("Expected '%s' to exist in installed directory, but it does not", metadataPath)
		}
	})

	// Test Uninstall
	t.Run("uninstall plasmoid", func(t *testing.T) {
		if err := cmd.UninstallPlasmoid(); err != nil {
			t.Fatalf("UninstallPlasmoid failed: %v", err)
		}

		if _, err := os.Stat(dest); err == nil {
			t.Errorf("Expected install directory '%s' to be removed, but it still exists", dest)
		}
	})
}