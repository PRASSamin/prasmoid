/*
Copyright 2025 PRAS
*/
package format

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/PRASSamin/prasmoid/tests"
	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrettify(t *testing.T) {
	_, cleanup := tests.SetupTestProject(t)
	defer cleanup()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	prettify("contents")

	require.NoError(t, w.Close())
	var buf bytes.Buffer
	_,_ = io.Copy(&buf, r)
	os.Stdout = oldStdout
	color.Output = os.Stdout
	output := buf.String()
	assert.Contains(t, output, "Formatted 1 files")
}

func TestFormatCmdRun(t *testing.T) {
	t.Run("Not a valid plasmoid", func(t *testing.T) {
		// Create an empty directory that is not a plasmoid
		tmpDir, err := os.MkdirTemp("", "format-invalid-*")
		require.NoError(t, err)
		defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

		oldWd, _ := os.Getwd()
		require.NoError(t, os.Chdir(tmpDir))
		defer func() { require.NoError(t, os.Chdir(oldWd)) }()

		// Capture stderr to check for the error message
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		FormatCmd.Run(FormatCmd, []string{})

		require.NoError(t, w.Close())
		var buf bytes.Buffer
		_,_ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		color.Output = os.Stdout

		assert.Contains(t, buf.String(), "Current directory is not a valid plasmoid")
	})

	t.Run("Run format successfully", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		FormatCmd.Run(FormatCmd, []string{})

		require.NoError(t, w.Close())
		var buf bytes.Buffer
		_,_ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		color.Output = os.Stdout
		output := buf.String()
		assert.Contains(t, output, "Formatted 1 files")
	})

	t.Run("qmlformat not installed, user confirms installation", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		// Mock functions
		utilsIsPackageInstalled = func(name string) bool { return false }
		utilsDetectPackageManager = func() (string, error) { return "apt", nil }
		surveyAskOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
			*(response.(*bool)) = true
			return nil
		}
		utilsInstallQmlformat = func(pm string) error { return nil }

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		FormatCmd.Run(FormatCmd, []string{})

		require.NoError(t, w.Close())
		var buf bytes.Buffer
		_,_ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		color.Output = os.Stdout
		output := buf.String()
		assert.Contains(t, output, "qmlformat installed successfully")
	})

	t.Run("qmlformat not installed, user cancels installation", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		// Mock functions
		utilsIsPackageInstalled = func(name string) bool { return false }
		utilsDetectPackageManager = func() (string, error) { return "apt", nil }
		surveyAskOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
			*(response.(*bool)) = false
			return nil
		}

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		FormatCmd.Run(FormatCmd, []string{})

		require.NoError(t, w.Close())
		var buf bytes.Buffer
		_,_ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		color.Output = os.Stdout
		output := buf.String()
		assert.Contains(t, output, "Operation cancelled")
	})

	t.Run("qmlformat not installed, installation fails", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		// Mock functions
		utilsIsPackageInstalled = func(name string) bool { return false }
		utilsDetectPackageManager = func() (string, error) { return "apt", nil }
		surveyAskOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
			*(response.(*bool)) = true
			return nil
		}
		utilsInstallQmlformat = func(pm string) error { return errors.New("install error") }

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		FormatCmd.Run(FormatCmd, []string{})

		require.NoError(t, w.Close())
		var buf bytes.Buffer
		_,_ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		color.Output = os.Stdout
		output := buf.String()
		assert.Contains(t, output, "Failed to install qmlformat")
	})

	t.Run("survey returns error", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		// Mock functions
		utilsIsPackageInstalled = func(name string) bool { return false }
		utilsDetectPackageManager = func() (string, error) { return "apt", nil }
		surveyAskOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
			return errors.New("survey error")
		}

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		FormatCmd.Run(FormatCmd, []string{})

		require.NoError(t, w.Close())
		var buf bytes.Buffer
		_,_ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		color.Output = os.Stdout
		output := buf.String()
		assert.NotContains(t, output, "qmlformat installed successfully")
		assert.NotContains(t, output, "Operation cancelled")
	})
}

func TestPrettifyOnWatch(t *testing.T) {
	tmpDir, cleanup := tests.SetupTestProject(t)
	defer cleanup()

	contentsPath := filepath.Join(tmpDir, "contents")

	// The function blocks, so it needs to run in a goroutine for the test.
	go prettifyOnWatch(contentsPath)

	// Give the watcher time to start up.
	time.Sleep(200 * time.Millisecond)

	// Trigger a write event to a QML file.
	qmlFile := filepath.Join(contentsPath, "ui", "main.qml")
	require.NoError(t, os.WriteFile(qmlFile, []byte("new content"), 0644))

	// Wait for the debounce logic to trigger the format command.
	time.Sleep(400 * time.Millisecond)
}

func TestPrettifyError(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	prettify("/non-existent-path")

	require.NoError(t, w.Close())
	var buf bytes.Buffer
	_,_ = io.Copy(&buf, r)
	os.Stdout = oldStdout
	color.Output = os.Stdout
	output := buf.String()
	assert.Contains(t, output, "Error walking directory for prettify")
}

func TestFormatError(t *testing.T) {
	// Mock exec.Command to return a command that will fail
	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("non-existent-command")
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	format([]string{"file1.qml"})

	require.NoError(t, w.Close())
	var buf bytes.Buffer
	_,_ = io.Copy(&buf, r)
	os.Stdout = oldStdout
	color.Output = os.Stdout
	output := buf.String()
	assert.Contains(t, output, "Failed to format qml files")
}
