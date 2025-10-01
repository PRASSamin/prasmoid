package upgrade

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/PRASSamin/prasmoid/utils"
	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

// TestUpgradeCmd tests the upgradeCmd.Run function
func TestUpgradeCmd(t *testing.T) {
	// Save original functions and restore after test
	originalCheckRoot := utils.CheckRoot
	originalOsExecutable := osExecutable
	originalExecCommand := execCommand
	originalOsRemove := osRemove
	originalRootGetCacheFilePath := rootGetCacheFilePath
	originalUtilsIsPackageInstalled := utilsIsPackageInstalled

	t.Cleanup(func() {
		utilsCheckRoot = originalCheckRoot
		osExecutable = originalOsExecutable
		execCommand = originalExecCommand
		osRemove = originalOsRemove
		rootGetCacheFilePath = originalRootGetCacheFilePath
		utilsIsPackageInstalled = originalUtilsIsPackageInstalled
	})

	// Mock checkRoot to always succeed by default for upgradeCmd tests
	utilsCheckRoot = func() error { return nil }

	// Helper to capture stdout
	captureOutput := func() (*bytes.Buffer, func()) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		color.Output = w // Redirect color output as well
		buf := new(bytes.Buffer)
		return buf, func() {
			_ = w.Close()
			_, _ = io.Copy(buf, r)
			os.Stdout = oldStdout
			color.Output = oldStdout
		}
	}

	t.Run("curl not installed", func(t *testing.T) {
		// Arrange
		utilsIsPackageInstalled = func(pkg string) bool { return false }
		defer func() { utilsIsPackageInstalled = originalUtilsIsPackageInstalled }()

		buf, restoreOutput := captureOutput()

		// Act
		upgradeCmd.Run(nil, []string{})

		// Assert
		restoreOutput()
		assert.Contains(t, buf.String(), "upgrade command is disabled due to missing dependencies.")
	})

	t.Run("checkRoot fails", func(t *testing.T) {
		// Arrange
		utilsIsPackageInstalled = func(pkg string) bool { return true }
		utilsCheckRoot = func() error { return errors.New("root check failed") }
		defer func() { utilsCheckRoot = func() error { return nil } }()

		buf, restoreOutput := captureOutput()

		// Act
		upgradeCmd.Run(nil, []string{})

		// Assert
		restoreOutput()
		assert.Contains(t, buf.String(), "root check failed")
	})

	t.Run("os.Executable fails", func(t *testing.T) {
		// Arrange
		utilsIsPackageInstalled = func(pkg string) bool { return true }
		osExecutable = func() (string, error) { return "", errors.New("exec error") }

		buf, restoreOutput := captureOutput()

		// Act
		upgradeCmd.Run(nil, []string{})

		// Assert
		restoreOutput()
		assert.Contains(t, buf.String(), "Failed to get current executable path: exec error")
	})

	t.Run("command.Run fails", func(t *testing.T) {
		// Arrange
		utilsIsPackageInstalled = func(pkg string) bool { return true }
		osExecutable = func() (string, error) { return "/usr/local/bin/prasmoid", nil }
		execCommand = func(name string, arg ...string) *exec.Cmd {
			// Mock the command to fail
			return exec.Command("bash", "-c", "exit 1")
		}

		buf, restoreOutput := captureOutput()

		// Act
		upgradeCmd.Run(nil, []string{})

		// Assert
		restoreOutput()
		assert.Contains(t, buf.String(), "Update failed: exit status 1")
	})

	t.Run("os.Remove fails", func(t *testing.T) {
		// Arrange
		utilsIsPackageInstalled = func(pkg string) bool { return true }
		osExecutable = func() (string, error) { return "/usr/local/bin/prasmoid", nil }
		execCommand = func(name string, arg ...string) *exec.Cmd {
			// Mock the command to succeed
			return exec.Command("bash", "-c", "exit 0")
		}
		osRemove = func(name string) error { return errors.New("remove error") }
		rootGetCacheFilePath = func() string { return "/tmp/cache" }

		buf, restoreOutput := captureOutput()

		// Act
		upgradeCmd.Run(nil, []string{})

		// Assert
		restoreOutput()
		assert.Contains(t, buf.String(), "Warning: Failed to remove update cache file: remove error")
	})
}

