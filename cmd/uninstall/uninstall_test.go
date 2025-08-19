package uninstall

import (
	"os"
	"testing"

	"github.com/PRASSamin/prasmoid/cmd/install"
	"github.com/PRASSamin/prasmoid/utils"
)


func TestInstallAndUninstall(t *testing.T) {
	_, _, cleanup := install.SetupTestEnvironment(t)
	defer cleanup()

	dest, err := utils.GetDevDest()
	if err != nil {
		t.Fatalf("GetDevDest() failed: %v", err)
	}

	// Test Uninstall
	t.Run("uninstall plasmoid", func(t *testing.T) {
		if err := UninstallPlasmoid(); err != nil {
			t.Fatalf("UninstallPlasmoid failed: %v", err)
		}

		if _, err := os.Stat(dest); err == nil {
			t.Errorf("Expected install directory '%s' to be removed, but it still exists", dest)
		}
	})
}
