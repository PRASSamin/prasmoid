package upgrade

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"os/user"
	"testing"

	"github.com/AlecAivazis/survey/v2"
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
	originalUtilsIsPackageInstalled := utilsIsPackageInstalled
	originalUtilsDetectPackageManager := utilsDetectPackageManager
	originalSurveyAskOne := surveyAskOne
	originalUtilsInstallPackage := utilsInstallPackage
	originalOsExecutable := osExecutable
	originalExecCommand := execCommand
	originalOsRemove := osRemove
	originalRootGetCacheFilePath := rootGetCacheFilePath

	t.Cleanup(func() {
		checkRoot = originalCheckRoot
		utilsIsPackageInstalled = originalUtilsIsPackageInstalled
		utilsDetectPackageManager = originalUtilsDetectPackageManager
		surveyAskOne = originalSurveyAskOne
		utilsInstallPackage = originalUtilsInstallPackage
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

	t.Run("curl not installed, user confirms, install succeeds, upgrade succeeds", func(t *testing.T) {
		// Arrange
		utilsIsPackageInstalled = func(pkg string) bool { return false }
		utilsDetectPackageManager = func() (string, error) { return "apt", nil }
		surveyAskOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
			*(response.(*bool)) = true // User confirms
			return nil
		}
		utilsInstallPackage = func(pm, pkg string, names map[string]string) error { return nil }
		osExecutable = func() (string, error) { return "/usr/local/bin/prasmoid", nil }
		execCommand = func(name string, arg ...string) *exec.Cmd {
			// Mock the command to succeed
			return exec.Command("bash", "-c", "exit 0")
		}
		osRemove = func(name string) error { return nil }
		rootGetCacheFilePath = func() string { return "/tmp/cache" }

		buf, restoreOutput := captureOutput()

		// Act
		upgradeCmd.Run(nil, []string{})

		// Assert
		restoreOutput()
		assert.NotContains(t, buf.String(), "Failed to install curl")
		assert.NotContains(t, buf.String(), "Update failed")
		assert.NotContains(t, buf.String(), "Warning: Failed to remove update cache file")
	})

	t.Run("curl not installed, user confirms, install fails", func(t *testing.T) {
		// Arrange
		utilsIsPackageInstalled = func(pkg string) bool { return false }
		utilsDetectPackageManager = func() (string, error) { return "apt", nil }
		surveyAskOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
			*(response.(*bool)) = true // User confirms
			return nil
		}
		utilsInstallPackage = func(pm, pkg string, names map[string]string) error { return errors.New("install failed") }

		buf, restoreOutput := captureOutput()

		// Act
		upgradeCmd.Run(nil, []string{})

		// Assert
		restoreOutput()
		assert.Contains(t, buf.String(), "Failed to install curl")
	})

	t.Run("curl not installed, user cancels", func(t *testing.T) {
		// Arrange
		utilsIsPackageInstalled = func(pkg string) bool { return false }
		utilsDetectPackageManager = func() (string, error) { return "apt", nil } // Added mock for DetectPackageManager
		surveyAskOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
			*(response.(*bool)) = false // User cancels
			return nil
		}

		buf, restoreOutput := captureOutput()

		// Act
		upgradeCmd.Run(nil, []string{})

		// Assert
		restoreOutput()
		assert.Contains(t, buf.String(), "Operation cancelled.")
	})

	t.Run("curl not installed, ask fails", func(t *testing.T) {
		// Arrange
		utilsIsPackageInstalled = func(pkg string) bool { return false }
		utilsDetectPackageManager = func() (string, error) { return "apt", nil } // Added mock for DetectPackageManager
		surveyAskOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
			return errors.New("ask failed") // survey.AskOne returns error
		}

		buf, restoreOutput := captureOutput()

		// Act
		upgradeCmd.Run(nil, []string{})

		// Assert
		restoreOutput()
		assert.Contains(t, buf.String(), "Failed to ask for curl installation: ask failed")
	})

	t.Run("os.Executable fails", func(t *testing.T) {
		// Arrange
		utilsIsPackageInstalled = func(pkg string) bool { return true } // Assume curl is installed
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
		utilsIsPackageInstalled = func(pkg string) bool { return true } // Assume curl is installed
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
		utilsIsPackageInstalled = func(pkg string) bool { return true } // Assume curl is installed
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
