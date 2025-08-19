package setup

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func TestSetupCmd(t *testing.T) {
	// Save original function and restore after test
	originalInstallDependencies := utilsInstallDependencies
	t.Cleanup(func() {
		utilsInstallDependencies = originalInstallDependencies
	})

	t.Run("success", func(t *testing.T) {
		// Arrange
		utilsInstallDependencies = func() error {
			return nil
		}

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		color.Output = w

		// Act
		SetupCmd.Run(SetupCmd, []string{})
		_ = w.Close()

		// Assert
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		assert.Empty(t, buf.String(), "Expected no output on success")
	})

	t.Run("failure", func(t *testing.T) {
		// Arrange
		expectedError := errors.New("installation failed")
		utilsInstallDependencies = func() error {
			return expectedError
		}

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		color.Output = w

		// Act
		SetupCmd.Run(SetupCmd, []string{})
		_ = w.Close()

		// Assert
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		assert.Contains(t, buf.String(), "Failed to install dependencies: installation failed")
	})
}
