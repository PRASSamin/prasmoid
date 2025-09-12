package locales

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	root "github.com/PRASSamin/prasmoid/cmd"
	"github.com/PRASSamin/prasmoid/tests"
	"github.com/PRASSamin/prasmoid/types"
	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestI18nLocalesEditCmd(t *testing.T) {
	t.Run("invalid plasmoid directory", func(t *testing.T) {
		// Arrange
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		// Mock the validation to fail
		oldIsValidPlasmoid := utilsIsValidPlasmoid
		utilsIsValidPlasmoid = func() bool { return false }
		defer func() { utilsIsValidPlasmoid = oldIsValidPlasmoid }()

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		color.Output = w

		// Act
		I18nLocalesEditCmd.Run(I18nLocalesEditCmd, []string{})
		_ = w.Close()

		// Assert
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		assert.Contains(t, buf.String(), "Current directory is not a valid plasmoid.")
	})

	t.Run("user cancels locale selection", func(t *testing.T) {
		// Arrange
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		// Mock validation to succeed
		oldIsValidPlasmoid := utilsIsValidPlasmoid
		utilsIsValidPlasmoid = func() bool { return true }
		defer func() { utilsIsValidPlasmoid = oldIsValidPlasmoid }()

		// Mock AskForLocales to return nil (user cancelled)
		oldAskForLocales := utilsAskForLocales
		utilsAskForLocales = func(defaultLocales ...[]string) []string { return nil }
		defer func() { utilsAskForLocales = oldAskForLocales }()

		// Mock os.WriteFile to check if it's called
		var writeCalled bool
		oldWriteFile := osWriteFile
		osWriteFile = func(name string, data []byte, perm os.FileMode) error {
			writeCalled = true
			return nil
		}
		defer func() { osWriteFile = oldWriteFile }()

		// Act
		I18nLocalesEditCmd.Run(I18nLocalesEditCmd, []string{})

		// Assert
		assert.False(t, writeCalled, "os.WriteFile should not have been called")
	})

	t.Run("successfully edits locales", func(t *testing.T) {
		// Arrange
		projectDir, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		// Set initial config
		root.ConfigRC = types.Config{
			I18n: types.ConfigI18n{
				Locales: []string{"en"},
			},
		}

		// Mock validation to succeed
		oldIsValidPlasmoid := utilsIsValidPlasmoid
		utilsIsValidPlasmoid = func() bool { return true }
		defer func() { utilsIsValidPlasmoid = oldIsValidPlasmoid }()

		// Mock AskForLocales to return new locales
		newLocales := []string{"de", "es"}
		oldAskForLocales := utilsAskForLocales
		utilsAskForLocales = func(defaultLocales ...[]string) []string { return newLocales }
		defer func() { utilsAskForLocales = oldAskForLocales }()

		// Act
		I18nLocalesEditCmd.Run(I18nLocalesEditCmd, []string{})

		// Assert
		configFile := filepath.Join(projectDir, "prasmoid.config.js")
		content, err := os.ReadFile(configFile)
		require.NoError(t, err)
		assert.Contains(t, string(content), `"de"`)
		assert.Contains(t, string(content), `"es"`)
		assert.NotContains(t, string(content), `"en"`)
	})

	t.Run("file write fails", func(t *testing.T) {
		// Arrange
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		// Mock validation to succeed
		oldIsValidPlasmoid := utilsIsValidPlasmoid
		utilsIsValidPlasmoid = func() bool { return true }
		defer func() { utilsIsValidPlasmoid = oldIsValidPlasmoid }()

		// Mock AskForLocales to return new locales
		oldAskForLocales := utilsAskForLocales
		utilsAskForLocales = func(defaultLocales ...[]string) []string { return []string{"fr"} }
		defer func() { utilsAskForLocales = oldAskForLocales }()

		// Mock os.WriteFile to fail
		oldWriteFile := osWriteFile
		osWriteFile = func(name string, data []byte, perm os.FileMode) error {
			return errors.New("disk full")
		}
		defer func() { osWriteFile = oldWriteFile }()

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		color.Output = w

		// Act
		I18nLocalesEditCmd.Run(I18nLocalesEditCmd, []string{})
		_ = w.Close()

		// Assert
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		assert.Contains(t, buf.String(), "Error writing prasmoid.config.js: disk full")
	})
}
