package command

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/PRASSamin/prasmoid/cmd"
	"github.com/PRASSamin/prasmoid/consts"
	"github.com/PRASSamin/prasmoid/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockFileInfo is a mock implementation of fs.FileInfo
type MockFileInfo struct{}

func (m *MockFileInfo) Name() string       { return "mock.js" }
func (m *MockFileInfo) Size() int64        { return 0 }
func (m *MockFileInfo) Mode() fs.FileMode  { return 0 }
func (m *MockFileInfo) ModTime() time.Time { return time.Now() }
func (m *MockFileInfo) IsDir() bool        { return false }
func (m *MockFileInfo) Sys() interface{}   { return nil }

func TestCommandNameValidator(t *testing.T) {
	// Setup fake commands dir
	tmpDir := t.TempDir()
	cmd.ConfigRC.Commands.Dir = tmpDir

	t.Run("empty name returns error", func(t *testing.T) {
		err := commandNameValidator("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "command name cannot be empty")
	})

	t.Run("invalid chars returns error", func(t *testing.T) {
		invalidNames := []string{"bad name", "name@", "na/me", "na*me"}
		for _, n := range invalidNames {
			err := commandNameValidator(n)
			assert.Error(t, err, "expected error for name %s", n)
			assert.Contains(t, err.Error(), "invalid characters in command name")
		}
	})

	t.Run("existing command returns error", func(t *testing.T) {
		existingFile := filepath.Join(tmpDir, "exists.js")
		err := os.WriteFile(existingFile, []byte("dummy"), 0644)
		require.NoError(t, err)

		err = commandNameValidator("exists")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "command already exists")
	})

	t.Run("valid name passes", func(t *testing.T) {
		err := commandNameValidator("validCommand")
		assert.NoError(t, err)
	})
}


func TestAddCommand(t *testing.T) {
	// Backup original functions
	originalOsStat := osStat
	originalOsMkdirAll := osMkdirAll
	originalOsGetwd := osGetwd
	originalOsWriteFile := osWriteFile
	originalFilepathAbs := filepathAbs
	originalFilepathRel := filepathRel
	originalSurveyAskOne := surveyAskOne
	originalRegexpMustCompile := regexpMustCompile
	originalConfigRC := cmd.ConfigRC

	// Cleanup function to restore mocks
	t.Cleanup(func() {
		osStat = originalOsStat
		osMkdirAll = originalOsMkdirAll
		osGetwd = originalOsGetwd
		osWriteFile = originalOsWriteFile
		filepathAbs = originalFilepathAbs
		filepathRel = originalFilepathRel
		surveyAskOne = originalSurveyAskOne
		regexpMustCompile = originalRegexpMustCompile
		cmd.ConfigRC = originalConfigRC
	})

	// Common setup
	cmd.ConfigRC = types.Config{
		Commands: types.ConfigCommands{
			Dir: "test_commands",
		},
	}
	// Mock regexp.MustCompile to return a predictable regexp
	validRegexp := regexp.MustCompile(`[\/:*?"<>|\s@]`)
	regexpMustCompile = func(str string) *regexp.Regexp {
		return validRegexp
	}

	t.Run("success: create new command", func(t *testing.T) {
		// Arrange
		commandName := "my-command"
		var writtenContent []byte
		var writtenPath string

		osStat = func(name string) (fs.FileInfo, error) {
			return nil, os.ErrNotExist // command does not exist
		}
		osMkdirAll = func(path string, perm fs.FileMode) error {
			assert.Equal(t, "test_commands", path)
			return nil
		}
		osWriteFile = func(name string, data []byte, perm fs.FileMode) error {
			writtenPath = name
			writtenContent = data
			assert.Equal(t, fs.FileMode(0644), perm)
			return nil
		}
		osGetwd = func() (string, error) { return "/project", nil }
		filepathAbs = func(path string) (string, error) {
			if path == "test_commands/my-command.js" {
				return "/project/test_commands/my-command.js", nil
			}
			if path == "/project" {
				return "/project", nil
			}
			return path, nil
		}
		filepathRel = func(basepath, targpath string) (string, error) {
			assert.Equal(t, "/project/test_commands", basepath)
			assert.Equal(t, "/project/prasmoid.d.ts", targpath)
			return "../prasmoid.d.ts", nil
		}

		// Act
		err := AddCommand(commandName)

		// Assert
		require.NoError(t, err)
		expectedContent := fmt.Sprintf(consts.JS_COMMAND_TEMPLATE, "../prasmoid.d.ts", commandName)
		assert.Equal(t, filepath.Join("test_commands", "my-command.js"), writtenPath)
		assert.Equal(t, expectedContent, string(writtenContent))
	})

	t.Run("error: command already exists", func(t *testing.T) {
		// Arrange
		commandName := "existing-command"
		osStat = func(name string) (fs.FileInfo, error) {
			return &MockFileInfo{}, nil // command exists
		}

		// Act
		err := AddCommand(commandName)

		// Assert
		require.Error(t, err)
		assert.Equal(t, "command already exists", err.Error())
	})

	t.Run("success: using survey to get command name", func(t *testing.T) {
		// Arrange
		commandName := "from-survey"
		var writtenContent []byte

		surveyAskOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
			*(response.(*string)) = commandName
			return nil
		}
		osStat = func(name string) (fs.FileInfo, error) {
			return nil, os.ErrNotExist
		}
		osMkdirAll = func(path string, perm fs.FileMode) error { return nil }
		osWriteFile = func(name string, data []byte, perm fs.FileMode) error {
			writtenContent = data
			return nil
		}
		osGetwd = func() (string, error) { return "/project", nil }
		filepathAbs = func(path string) (string, error) {
			if path == "test_commands/from-survey.js" {
				return "/project/test_commands/from-survey.js", nil
			}
			if path == "/project" {
				return "/project", nil
			}
			return path, nil
		}
		filepathRel = func(basepath, targpath string) (string, error) {
			return "../prasmoid.d.ts", nil
		}

		// Act
		err := AddCommand("") // Trigger survey

		// Assert
		require.NoError(t, err)
		expectedContent := fmt.Sprintf(consts.JS_COMMAND_TEMPLATE, "../prasmoid.d.ts", commandName)
		assert.Equal(t, expectedContent, string(writtenContent))
	})

	t.Run("error: survey fails", func(t *testing.T) {
		// Arrange
		surveyAskOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
			return errors.New("survey error")
		}

		// Act
		err := AddCommand("") // Trigger survey

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error asking for command name")
	})

	t.Run("error: mkdir fails", func(t *testing.T) {
		// Arrange
		commandName := "any-command"
		osStat = func(name string) (fs.FileInfo, error) { return nil, os.ErrNotExist }
		osMkdirAll = func(path string, perm fs.FileMode) error { return errors.New("mkdir failed") }

		// Act
		err := AddCommand(commandName)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create commands directory")
	})

	t.Run("error: write file fails", func(t *testing.T) {
		// Arrange
		commandName := "any-command"
		osStat = func(name string) (fs.FileInfo, error) { return nil, os.ErrNotExist }
		osMkdirAll = func(path string, perm fs.FileMode) error { return nil }
		osGetwd = func() (string, error) { return "/project", nil }
		filepathAbs = func(path string) (string, error) { return path, nil }
		filepathRel = func(basepath, targpath string) (string, error) { return "../prasmoid.d.ts", nil }
		osWriteFile = func(name string, data []byte, perm fs.FileMode) error {
			return errors.New("write failed")
		}

		// Act
		err := AddCommand(commandName)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error writing to file")
	})

	t.Run("error: relpath fails", func(t *testing.T) {
		// Arrange
		commandName := "any-command"
		osStat = func(name string) (fs.FileInfo, error) { return nil, os.ErrNotExist }
		osMkdirAll = func(path string, perm fs.FileMode) error { return nil }
		osGetwd = func() (string, error) { return "/project", nil }
		filepathAbs = func(path string) (string, error) { return path, nil }
		filepathRel = func(basepath, targpath string) (string, error) {
			return "", errors.New("relpath failed")
		}

		// Act
		err := AddCommand(commandName)

		// Assert
		require.Error(t, err)
		assert.Contains(t, err.Error(), "error calculating relative path")
	})
}
