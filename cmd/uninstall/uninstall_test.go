package uninstall

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func TestUninstallCmd(t *testing.T) {
	// Save original functions and restore after test
	originalIsValidPlasmoid := utilsIsValidPlasmoid
	originalUninstallPlasmoid := UninstallPlasmoid
	t.Cleanup(func() {
		utilsIsValidPlasmoid = originalIsValidPlasmoid
		UninstallPlasmoid = originalUninstallPlasmoid
	})

	t.Run("invalid plasmoid", func(t *testing.T) {
		// Arrange
		utilsIsValidPlasmoid = func() bool { return false }

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		color.Output = w

		// Act
		UninstallCmd.Run(UninstallCmd, []string{})
		_ = w.Close()

		// Assert
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		assert.Contains(t, buf.String(), "Current directory is not a valid plasmoid.")
	})

	t.Run("uninstall success", func(t *testing.T) {
		// Arrange
		utilsIsValidPlasmoid = func() bool { return true }
		UninstallPlasmoid = func() error { return nil }

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		color.Output = w

		// Act
		UninstallCmd.Run(UninstallCmd, []string{})
		_ = w.Close()

		// Assert
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		assert.Contains(t, buf.String(), "Plasmoid uninstalled successfully")
	})

	t.Run("uninstall failure", func(t *testing.T) {
		// Arrange
		utilsIsValidPlasmoid = func() bool { return true }
		expectedError := errors.New("failed to remove")
		UninstallPlasmoid = func() error { return expectedError }

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		color.Output = w

		// Act
		UninstallCmd.Run(UninstallCmd, []string{})
		_ = w.Close()

		// Assert
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		assert.Contains(t, buf.String(), "Failed to uninstall plasmoid:")
		assert.Contains(t, buf.String(), expectedError.Error())
	})
}

func TestUninstallPlasmoid(t *testing.T) {
	// Save original functions and restore after test
	originalIsInstalled := utilsIsInstalled
	originalOsRemoveAll := osRemoveAll
	t.Cleanup(func() {
		utilsIsInstalled = originalIsInstalled
		osRemoveAll = originalOsRemoveAll
	})

	t.Run("IsInstalled returns error", func(t *testing.T) {
		// Arrange
		expectedError := errors.New("install check error")
		utilsIsInstalled = func() (bool, string, error) { return false, "", expectedError }

		// Act
		err := UninstallPlasmoid()

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
	})

	t.Run("not installed", func(t *testing.T) {
		// Arrange
		utilsIsInstalled = func() (bool, string, error) { return false, "", nil }
		var removeCalled bool
		osRemoveAll = func(path string) error { removeCalled = true; return nil }

		// Act
		err := UninstallPlasmoid()

		// Assert
		assert.NoError(t, err)
		assert.False(t, removeCalled, "os.RemoveAll should not be called if not installed")
	})

	t.Run("installed and remove succeeds", func(t *testing.T) {
		// Arrange
		installPath := "/tmp/test_plasmoid_install"
		utilsIsInstalled = func() (bool, string, error) { return true, installPath, nil }
		var removedPath string
		osRemoveAll = func(path string) error { removedPath = path; return nil }

		// Act
		err := UninstallPlasmoid()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, installPath, removedPath, "os.RemoveAll should be called with the correct path")
	})

	t.Run("installed and remove fails", func(t *testing.T) {
		// Arrange
		installPath := "/tmp/test_plasmoid_install"
		expectedError := errors.New("permission denied")
		utilsIsInstalled = func() (bool, string, error) { return true, installPath, nil }
		osRemoveAll = func(path string) error { return expectedError }

		// Act
		err := UninstallPlasmoid()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), fmt.Sprintf("failed to remove installation directory %s: %v", installPath, expectedError))
	})
}
