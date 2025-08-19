package regen

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/PRASSamin/prasmoid/consts"
	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func TestRegenConfigCmd(t *testing.T) {
	// Save original functions and restore after test
	originalAskForLocales := utilsAskForLocales
	originalCreateConfigFile := initCmdCreateConfigFile
	t.Cleanup(func() {
		utilsAskForLocales = originalAskForLocales
		initCmdCreateConfigFile = originalCreateConfigFile
	})

	// Mock implementations
	utilsAskForLocales = func(defaultLocales ...[]string) []string {
		return []string{"en", "fr"}
	}

	t.Run("success", func(t *testing.T) {
		// Arrange
		initCmdCreateConfigFile = func(locales []string) error {
			assert.Equal(t, []string{"en", "fr"}, locales)
			return nil
		}

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		color.Output = w

		// Act
		regenConfigCmd.Run(regenConfigCmd, []string{})
		_ = w.Close()

		// Assert
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		assert.Contains(t, buf.String(), "Config file regenerated successfully.")
	})

	t.Run("failure", func(t *testing.T) {
		// Arrange
		expectedError := errors.New("disk full")
		initCmdCreateConfigFile = func(locales []string) error {
			return expectedError
		}

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		color.Output = w

		// Act
		regenConfigCmd.Run(regenConfigCmd, []string{})
		_ = w.Close()

		// Assert
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		assert.Contains(t, buf.String(), "Failed to regenerate config file:")
		assert.Contains(t, buf.String(), expectedError.Error())
	})
}

func TestRegenTypesCmd(t *testing.T) {
	// Save original functions and restore after test
	originalCreateFileFromTemplate := initCmdCreateFileFromTemplate
	t.Cleanup(func() {
		initCmdCreateFileFromTemplate = originalCreateFileFromTemplate
	})

	t.Run("success", func(t *testing.T) {
		// Arrange
		initCmdCreateFileFromTemplate = func(filename string, content string) error {
			assert.Equal(t, "prasmoid.d.ts", filename)
			assert.Equal(t, consts.PRASMOID_DTS, content)
			return nil
		}

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		color.Output = w

		// Act
		regenTypesCmd.Run(regenTypesCmd, []string{})
		_ = w.Close()

		// Assert
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		assert.Contains(t, buf.String(), "prasmoid.d.ts regenerated successfully.")
	})

	t.Run("failure", func(t *testing.T) {
		// Arrange
		expectedError := errors.New("template error")
		initCmdCreateFileFromTemplate = func(filename string, content string) error {
			return expectedError
		}

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		color.Output = w

		// Act
		regenTypesCmd.Run(regenTypesCmd, []string{})
		_ = w.Close()

		// Assert
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		assert.Contains(t, buf.String(), "Failed to regenerate prasmoid.d.ts:")
		assert.Contains(t, buf.String(), expectedError.Error())
	})
}
