package unlink

import (
	"os"
	"testing"

	"github.com/PRASSamin/prasmoid/cmd/install"
	"github.com/PRASSamin/prasmoid/utils"
)

func TestLinkAndUnlink(t *testing.T) {
	_, _, cleanup := install.SetupTestEnvironment(t)
	defer cleanup()

	dest, err := utils.GetDevDest()
	if err != nil {
		t.Fatalf("GetDevDest() failed: %v", err)
	}

	// Test Unlink
	t.Run("unlink plasmoid", func(t *testing.T) {
		UnlinkCmd.Run(nil, []string{})

		if _, err := os.Lstat(dest); err == nil {
			t.Errorf("Expected symlink '%s' to be removed, but it still exists", dest)
		}
	})
}