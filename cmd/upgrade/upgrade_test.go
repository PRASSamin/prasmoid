package upgrade

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"os/user"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

// TestCheckRoot tests the checkRoot function
func TestCheckRoot(t *testing.T) {
	originalUserCurrent := userCurrent
	t.Cleanup(func() {
		userCurrent = originalUserCurrent
	})

	t.Run("user is root", func(t *testing.T) {
		userCurrent = func() (*user.User, error) {
			return &user.User{Uid: "0"}, nil
		}
		assert.NoError(t, checkRoot())
	})

	t.Run("user is not root", func(t *testing.T) {
		userCurrent = func() (*user.User, error) {
			return &user.User{Uid: "1000"}, nil
		}
		err := checkRoot()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "the requested operation requires superuser privileges")
	})

	t.Run("user.Current returns error", func(t *testing.T) {
		userCurrent = func() (*user.User, error) {
			return nil, errors.New("user error")
		}
		err := checkRoot()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get current user")
	})
}

// TestUpgradeCmd tests the upgradeCmd.Run function
func TestUpgradeCmd(t *testing.T) {
	// Save original functions and restore after test
	originalCheckRoot := checkRoot
	originalOsExecutable := osExecutable
	originalExecCommand := execCommand
	originalOsRemove := osRemove
	originalRootGetCacheFilePath := rootGetCacheFilePath

	t.Cleanup(func() {
		checkRoot = originalCheckRoot
		osExecutable = originalOsExecutable
		execCommand = originalExecCommand
		osRemove = originalOsRemove
		rootGetCacheFilePath = originalRootGetCacheFilePath
	})

	// Mock checkRoot to always succeed by default for upgradeCmd tests
	checkRoot = func() error { return nil }

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

	t.Run("checkRoot fails", func(t *testing.T) {
		// Arrange
		checkRoot = func() error { return errors.New("root check failed") }
		defer func() { checkRoot = func() error { return nil } }()

		buf, restoreOutput := captureOutput()

		// Act
		upgradeCmd.Run(nil, []string{})

		// Assert
		restoreOutput()
		assert.Contains(t, buf.String(), "root check failed")
	})

	t.Run("os.Executable fails", func(t *testing.T) {
		// Arrange
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
