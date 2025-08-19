package link

import (
	"os"
	"testing"

	"github.com/PRASSamin/prasmoid/tests"
	"github.com/PRASSamin/prasmoid/utils"
)

func TestLinkAndUnlink(t *testing.T) {
	_, _, cleanup := tests.SetupTestEnvironment(t)
	defer cleanup()

	dest, err := utils.GetDevDest()
	if err != nil {
		t.Fatalf("GetDevDest() failed: %v", err)
	}

	// Test Link
	t.Run("link plasmoid", func(t *testing.T) {
		if err := LinkPlasmoid(dest); err != nil {
			t.Fatalf("LinkPlasmoid failed: %v", err)
		}

		if _, err := os.Lstat(dest); os.IsNotExist(err) {
			t.Errorf("Expected symlink '%s' to be created, but it was not", dest)
		}
	})
}