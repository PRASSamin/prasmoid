package command

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/AlecAivazis/survey/v2"
	"github.com/PRASSamin/prasmoid/cmd"
	"github.com/PRASSamin/prasmoid/tests"
	"github.com/PRASSamin/prasmoid/types"
	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func TestRemoveCommand(t *testing.T) {
	// Common setup
	cmd.ConfigRC = types.Config{
		Commands: types.ConfigCommands{
			Dir: "test_commands",
		},
	}

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

	t.Run("successfully removes a command with force flag", func(t *testing.T) {
		// Arrange
		t.Cleanup(func() {
			filepathWalk = filepath.Walk
			osStat = os.Stat
			osRemove = os.Remove
		})

		var removedPath string
		filepathWalk = func(root string, walkFn filepath.WalkFunc) error {
			return walkFn("test_commands/test-cmd.js", &MockFileInfo{}, nil)
		}
		osStat = func(name string) (fs.FileInfo, error) { return &MockFileInfo{}, nil }
		osRemove = func(name string) error {
			removedPath = name
			return nil
		}
		buf, restore := captureOutput()

		_ = commandsRemoveCmd.Flags().Set("force", "true")
		_ = commandsRemoveCmd.Flags().Set("name", "test-cmd")
		commandsRemoveCmd.Run(commandsRemoveCmd, []string{})

		// Assert
		restore()
		output := buf.String()
		assert.Equal(t, filepath.Join("test_commands", "test-cmd.js"), removedPath)
		assert.Contains(t, output, "Successfully removed command: test-cmd")
	})

	t.Run("successfully removes a command with survey", func(t *testing.T) {
		// Arrange
		t.Cleanup(func() {
			filepathWalk = filepath.Walk
			surveyAskOne = survey.AskOne
			osStat = os.Stat
			osRemove = os.Remove
		})

		filepathWalk = func(root string, walkFn filepath.WalkFunc) error {
			return walkFn("test_commands/test-cmd.js", &MockFileInfo{}, nil)

		}
		surveyAskOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
			switch p.(type) {
			case *survey.Select:
				*(response.(*string)) = "test-cmd (test-cmd.js)"
			case *survey.Confirm:
				*(response.(*bool)) = true
			}
			return nil
		}
		osStat = func(name string) (fs.FileInfo, error) { return &MockFileInfo{}, nil }
		osRemove = func(name string) error { return nil }
		buf, restore := captureOutput()

		// Act
		RemoveCommand("", false) // trigger survey

		// Assert
		restore()
		output := buf.String()
		assert.Contains(t, output, "Successfully removed command: test-cmd")
	})

	t.Run("fails to remove command", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()
		// Arrange
		t.Cleanup(func() { osRemove = os.Remove })

		AddCommand("test-cmd")
		osRemove = func(name string) error { return errors.New("remove error") }
		buf, restore := captureOutput()

		// Act
		RemoveCommand("test-cmd", true)

		// Assert
		restore()
		output := buf.String()
		assert.Contains(t, output, "Error removing file")
	})

	t.Run("no commands found", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()
		// Arrange
		t.Cleanup(func() { osRemove = os.Remove })

		buf, restore := captureOutput()

		// Act
		RemoveCommand("test-cmd", true)

		// Assert
		restore()
		output := buf.String()
		assert.Contains(t, output, "No commands found in the commands directory.")
	})

	t.Run("command file does not exist", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()
		// Arrange
		t.Cleanup(func() { osRemove = os.Remove })
		AddCommand("test-cmd")
		buf, restore := captureOutput()

		// Act
		RemoveCommand("test-cmd2", true)

		// Assert
		restore()
		output := buf.String()
		assert.Contains(t, output, "Command file does not exist")
	})

	t.Run("filepath.Walk fails", func(t *testing.T) {
		// Arrange
		t.Cleanup(func() { filepathWalk = filepath.Walk })
		filepathWalk = func(root string, walkFn filepath.WalkFunc) error { return errors.New("walk error") }
		buf, restore := captureOutput()

		// Act
		RemoveCommand("test-cmd", true)

		// Assert
		restore()
		output := buf.String()
		assert.Contains(t, output, "Error walking commands directory")
	})
}
