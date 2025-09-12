package link

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
	"github.com/stretchr/testify/require"
)

func TestLinkCmd(t *testing.T) {
	t.Run("not a valid plasmoid", func(t *testing.T) {
		// Arrange
		_, _, cleanup := tests.SetupTestEnvironment(t)
		defer cleanup()
		_ = os.Remove("metadata.json") // Make it invalid

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		color.Output = w

		// Act
		LinkCmd.Run(LinkCmd, []string{})
		_ = w.Close()

		// Assert
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		assert.Contains(t, buf.String(), "Current directory is not a valid plasmoid.")
	})

	t.Run("GetDevDest fails", func(t *testing.T) {
		// Arrange
		_, _, cleanup := tests.SetupTestEnvironment(t)
		defer cleanup()

		// Mock
		oldGetDevDest := utilsGetDevDest
		utilsGetDevDest = func() (string, error) { return "", errors.New("get dev dest error") }
		defer func() { utilsGetDevDest = oldGetDevDest }()

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		color.Output = w

		// Act
		LinkCmd.Run(LinkCmd, []string{})
		_ = w.Close()

		// Assert
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		assert.Contains(t, buf.String(), "get dev dest error")
	})

	t.Run("where flag shows path", func(t *testing.T) {
		// Arrange
		_, _, cleanup := tests.SetupTestEnvironment(t)
		defer cleanup()
		_ = LinkCmd.Flags().Set("where", "true")
		defer func() { _ = LinkCmd.Flags().Set("where", "false") }()

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		color.Output = w

		// Act
		LinkCmd.Run(LinkCmd, []string{})
		_ = w.Close()

		// Assert
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		dest, _ := utils.GetDevDest()
		assert.Contains(t, buf.String(), "Plasmoid linked to:")
		assert.Contains(t, buf.String(), dest)
	})

	t.Run("LinkPlasmoid fails", func(t *testing.T) {
		// Arrange
		_, _, cleanup := tests.SetupTestEnvironment(t)
		defer cleanup()

		// Mock
		oldLinkPlasmoid := LinkPlasmoid
		LinkPlasmoid = func(dest string) error { return errors.New("link error") }
		defer func() { LinkPlasmoid = oldLinkPlasmoid }()

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		color.Output = w

		// Act
		LinkCmd.Run(LinkCmd, []string{})
		_ = w.Close()

		// Assert
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		assert.Contains(t, buf.String(), "Failed to link plasmoid:")
	})

	t.Run("link succeeds", func(t *testing.T) {
		// Arrange
		_, _, cleanup := tests.SetupTestEnvironment(t)
		defer cleanup()

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		color.Output = w

		// Act
		LinkCmd.Run(LinkCmd, []string{})
		_ = w.Close()

		// Assert
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		assert.Contains(t, buf.String(), "Plasmoid linked successfully.")
	})
}

func TestLinkPlasmoid(t *testing.T) {
	t.Run("successful link", func(t *testing.T) {
		// Arrange
		_, _, cleanup := tests.SetupTestEnvironment(t)
		defer cleanup()
		dest, _ := utils.GetDevDest()

		// Act
		err := LinkPlasmoid(dest)

		// Assert
		require.NoError(t, err)
		_, err = os.Lstat(dest)
		assert.NoError(t, err, "symlink should exist")
	})

	t.Run("destination already exists", func(t *testing.T) {
		// Arrange
		_, _, cleanup := tests.SetupTestEnvironment(t)
		defer cleanup()
		dest, _ := utils.GetDevDest()
		// Create a dummy file at the destination
		_ = os.WriteFile(dest, []byte("dummy"), 0644)

		// Act
		err := LinkPlasmoid(dest)

		// Assert
		require.NoError(t, err)
		linkInfo, _ := os.Lstat(dest)
		// Check if it's a symlink now
		assert.True(t, linkInfo.Mode()&os.ModeSymlink != 0)
	})

	t.Run("os.Getwd fails", func(t *testing.T) {
		// Arrange
		_, _, cleanup := tests.SetupTestEnvironment(t)
		defer cleanup()

		// Mock
		oldGetwd := osGetwd
		osGetwd = func() (string, error) { return "", errors.New("getwd error") }
		defer func() { osGetwd = oldGetwd }()

		// Act & Assert
		assert.Error(t, LinkPlasmoid("dest"))
	})

	t.Run("os.Symlink fails", func(t *testing.T) {
		// Arrange
		_, _, cleanup := tests.SetupTestEnvironment(t)
		defer cleanup()

		// Mock
		oldSymlink := osSymlink
		osSymlink = func(oldname, newname string) error { return errors.New("symlink error") }
		defer func() { osSymlink = oldSymlink }()

		// Act & Assert
		assert.Error(t, LinkPlasmoid("dest"))
	})
}
