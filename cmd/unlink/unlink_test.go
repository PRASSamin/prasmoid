package unlink

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/PRASSamin/prasmoid/tests"
	"github.com/PRASSamin/prasmoid/utils"
	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func TestLinkAndUnlink(t *testing.T) {
	_, _, cleanup := tests.SetupTestEnvironment(t)
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

	t.Run("failed to get destination", func(t *testing.T) {
		orgUtilsGetDevDest := utilsGetDevDest
		utilsGetDevDest = func() (string, error) { return "", errors.New("getdest error") }
		defer func() { utilsGetDevDest = orgUtilsGetDevDest }()

		orgStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		color.Output = w
		UnlinkCmd.Run(nil, []string{})
		_ = w.Close()

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = orgStdout
		assert.Contains(t, buf.String(), "getdest error")
	})
}
