package changeset

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlecAivazis/survey/v2"
	"github.com/PRASSamin/prasmoid/tests"
	"github.com/PRASSamin/prasmoid/utils"
	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetNextVersion(t *testing.T) {
	testCases := []struct {
		name           string
		currentVersion string
		bump           string
		expected       string
		expectError    bool
	}{
		{
			name:           "patch bump",
			currentVersion: "1.2.3",
			bump:           "patch",
			expected:       "1.2.4",
			expectError:    false,
		},
		{
			name:           "minor bump",
			currentVersion: "1.2.3",
			bump:           "minor",
			expected:       "1.3.0",
			expectError:    false,
		},
		{
			name:           "major bump",
			currentVersion: "1.2.3",
			bump:           "major",
			expected:       "2.0.0",
			expectError:    false,
		},
		{
			name:           "invalid version format",
			currentVersion: "1.2",
			bump:           "patch",
			expected:       "",
			expectError:    true,
		},
		{
			name:           "invalid bump type",
			currentVersion: "1.2.3",
			bump:           "invalid",
			expected:       "",
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nextVersion, err := GetNextVersion(tc.currentVersion, tc.bump)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, nextVersion)
			}
		})
	}
}

func TestOpenEditor(t *testing.T) {
	t.Run("successful edit with content", func(t *testing.T) {
		// Arrange
		t.Cleanup(func() {
			osGetenv = os.Getenv
			osCreateTemp = os.CreateTemp
			osRemove = os.Remove
			execCommand = exec.Command
			osReadFile = os.ReadFile
		})

		osGetenv = func(key string) string { return "vim" }
		osCreateTemp = func(dir, pattern string) (*os.File, error) {
			_, w, err := os.Pipe()
			return w, err
		}
		osRemove = func(name string) error { return nil }
		execCommand = func(name string, arg ...string) *exec.Cmd {
			return exec.Command("true")
		}
		osReadFile = func(name string) ([]byte, error) {
			return []byte("This is a test changelog entry."), nil
		}

		// Act
		content, err := OpenEditor()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "This is a test changelog entry.", content)
	})

	t.Run("missing editor env var", func(t *testing.T) {
		// Arrange
		t.Cleanup(func() {
			osGetenv = os.Getenv
			osCreateTemp = os.CreateTemp
			osRemove = os.Remove
			execCommand = exec.Command
			osReadFile = os.ReadFile
		})

		osGetenv = func(key string) string { return "" }
		osCreateTemp = func(dir, pattern string) (*os.File, error) {
			_, w, err := os.Pipe()
			return w, err
		}
		osRemove = func(name string) error { return nil }
		execCommand = func(name string, arg ...string) *exec.Cmd {
			return exec.Command("true")
		}
		osReadFile = func(name string) ([]byte, error) {
			return []byte("This is a test changelog entry with nano."), nil
		}

		// Act
		content, err := OpenEditor()

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, "This is a test changelog entry with nano.", content)
	})

	t.Run("editor returns an error", func(t *testing.T) {
		// Arrange
		t.Cleanup(func() { execCommand = exec.Command })
		execCommand = func(name string, arg ...string) *exec.Cmd {
			return exec.Command("false") // command that fails
		}

		// Act
		_, err := OpenEditor()

		// Assert
		assert.Error(t, err)
	})
}

func TestAddChangeset(t *testing.T) {
	normalize := func(s string) string {
		return strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", "\n"), "\n", ""))
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

	t.Run("successful changeset creation with flags", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()
		// Arrange
		t.Cleanup(func() {
			utilsIsValidPlasmoid = utils.IsValidPlasmoid
			utilsGetDataFromMetadata = utils.GetDataFromMetadata
			osMkdirAll = os.MkdirAll
			osWriteFile = os.WriteFile
		})

		utilsIsValidPlasmoid = func() bool { return true }
		utilsGetDataFromMetadata = func(key string) (interface{}, error) {
			if key == "Id" {
				return "org.kde.test", nil
			}
			if key == "Version" {
				return "1.0.0", nil
			}
			return nil, nil
		}
		osMkdirAll = func(path string, perm os.FileMode) error { return nil }
		var writtenFile string
		osWriteFile = func(name string, data []byte, perm os.FileMode) error {
			writtenFile = string(data)
			return nil
		}
		_ = changesetAddCmd.Flags().Set("bump", "patch")
		_ = changesetAddCmd.Flags().Set("summary", "Test summary")
		_ = changesetAddCmd.Flags().Set("apply", "false")

		// Act
		changesetAddCmd.Run(changesetAddCmd, []string{})

		// Assert
		assert.Contains(t, writtenFile, "bump: patch")
		assert.Contains(t, writtenFile, "next: 1.0.1")
		assert.Contains(t, writtenFile, "Test summary")
	})

	t.Run("successful changeset creation with prompts", func(t *testing.T) {
		tmpDir, cleanup := tests.SetupTestProject(t)
		defer cleanup()
		// Arrange
		t.Cleanup(func() {
			surveyAskOne = survey.AskOne
		})

		surveyAskOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
			switch v := p.(type) {
			case *survey.Select:
				*(response.(*string)) = v.Options[0] // select patch
			case *survey.Input:
				*(response.(*string)) = "Prompted summary"
			}
			return nil
		}
		OpenEditor = func() (string, error) { return "", errors.New("editor failed") }

		// Act
		AddChangeset("", "", true)

		// verify summary in changelog
		changelogPath := filepath.Join(tmpDir, "CHANGELOG.md")
		content, err := os.ReadFile(changelogPath)
		require.NoError(t, err)
		actual := normalize(string(content))
		require.Contains(t, actual, "Prompted summary")

		// Check version after applying changeset
		version, err := utils.GetDataFromMetadata("Version")
		require.NoError(t, err)
		require.Equal(t, "1.0.1", version)
	})

	t.Run("invalid plasmoid", func(t *testing.T) {
		// Arrange
		t.Cleanup(func() { utilsIsValidPlasmoid = utils.IsValidPlasmoid })
		utilsIsValidPlasmoid = func() bool { return false }
		buf, restore := captureOutput()

		// Act
		AddChangeset("patch", "summary", false)

		restore()
		output := buf.String()
		// Assert
		assert.Contains(t, output, "Current directory is not a valid plasmoid")
	})
}
