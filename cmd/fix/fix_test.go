package fix

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func TestCliFixCmd(t *testing.T) {
	// Save original functions
	originalUtilsIsPackageInstalled := utilsIsPackageInstalled
	originalCheckRoot := utilsCheckRoot
	originalExecCommand := execCommand

	t.Cleanup(func() {
		utilsIsPackageInstalled = originalUtilsIsPackageInstalled
		utilsCheckRoot = originalCheckRoot
		execCommand = originalExecCommand
	})

	// Helper to capture stdout
	captureOutput := func() (*bytes.Buffer, func()) {
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		color.Output = w
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
		buf, restore := captureOutput()

		// Act
		cliFixCmd.Run(cliFixCmd, []string{})

		// Assert
		restore()
		output := buf.String()
		assert.Contains(t, output, "fix command is disabled due to missing curl dependency.")
	})

	t.Run("checkRoot fails", func(t *testing.T) {
		// Arrange
		utilsIsPackageInstalled = func(pkg string) bool { return true }
		utilsCheckRoot = func() error { return errors.New("not root") }
		buf, restore := captureOutput()

		// Act
		cliFixCmd.Run(cliFixCmd, []string{})

		// Assert
		restore()
		output := buf.String()
		assert.Contains(t, output, "not root")
	})

	t.Run("exec command fails", func(t *testing.T) {
		// Arrange
		utilsIsPackageInstalled = func(pkg string) bool { return true }
		utilsCheckRoot = func() error { return nil }
		execCommand = func(name string, arg ...string) *exec.Cmd {
			return exec.Command("bash", "-c", "exit 1")
		}
		buf, restore := captureOutput()

		// Act
		cliFixCmd.Run(cliFixCmd, []string{})

		// Assert
		restore()
		output := buf.String()
		assert.Contains(t, output, "Fix failed")
	})

	t.Run("exec command succeeds", func(t *testing.T) {
		// Arrange
		utilsIsPackageInstalled = func(pkg string) bool { return true }
		utilsCheckRoot = func() error { return nil }
		execCommand = func(name string, arg ...string) *exec.Cmd {
			return exec.Command("true")
		}
		buf, restore := captureOutput()

		// Act
		cliFixCmd.Run(cliFixCmd, []string{})

		// Assert
		restore()
		output := buf.String()
		assert.NotContains(t, output, "Fix failed")
	})
}