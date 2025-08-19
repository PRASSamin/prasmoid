package install

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/PRASSamin/prasmoid/utils"
)

func TestInstallAndUninstall(t *testing.T) {
	_, _, cleanup := SetupTestEnvironment(t)
	defer cleanup()

	dest, err := utils.GetDevDest()
	if err != nil {
		t.Fatalf("GetDevDest() failed: %v", err)
	}

	// Test Install
	t.Run("install plasmoid", func(t *testing.T) {
		if err := InstallPlasmoid(); err != nil {
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

}
