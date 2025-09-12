package i18n

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AlecAivazis/survey/v2"
	"github.com/PRASSamin/prasmoid/cmd"
	"github.com/PRASSamin/prasmoid/tests"
	"github.com/PRASSamin/prasmoid/utils"
	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockExecCommand is a helper to mock exec.Command for testing command failures
func mockExecCommand(t *testing.T, failingCmd string) {
	t.Helper()
	oldExecCommand := execCommand
	execCommand = func(name string, arg ...string) *exec.Cmd {
		if name == failingCmd {
			return exec.Command("non-existent-command") // This will fail
		}
		return oldExecCommand(name, arg...)
	}
	t.Cleanup(func() { execCommand = oldExecCommand })
}

func TestI18nExtractCommand(t *testing.T) {
	t.Run("invalid plasmoid", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()
		_ = os.Remove("metadata.json")

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		I18nExtractCmd.Run(I18nExtractCmd, []string{})

		require.NoError(t, w.Close())
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		color.Output = os.Stdout
		output := buf.String()
		assert.Contains(t, output, "Current directory is not a valid plasmoid")
	})

	t.Run("successfully extracts and creates .po files", func(t *testing.T) {
		// Arrange
		projectDir, cleanup := tests.SetupTestProject(t)
		defer cleanup()
		cmd.ConfigRC = utils.LoadConfigRC()
		qmlDir := filepath.Join(projectDir, "contents", "ui")
		_ = os.MkdirAll(qmlDir, 0755)
		_ = os.WriteFile(filepath.Join(qmlDir, "main.qml"), []byte(`Text { text: i18n("Hello") }`), 0644)
		_ = I18nExtractCmd.Flags().Set("no-po", "false")

		// Act
		I18nExtractCmd.Run(I18nExtractCmd, []string{})

		// Assert
		potFile := filepath.Join(projectDir, "translations", "template.pot")
		enPoFile := filepath.Join(projectDir, "translations", "en.po")
		require.FileExists(t, potFile)
		require.FileExists(t, enPoFile)
		potContent, _ := os.ReadFile(potFile)
		assert.Contains(t, string(potContent), `msgid "Hello"`)
	})

	t.Run("skips .po generation with --no-po flag", func(t *testing.T) {
		// Arrange
		projectDir, cleanup := tests.SetupTestProject(t)
		defer cleanup()
		cmd.ConfigRC = utils.LoadConfigRC()
		qmlDir := filepath.Join(projectDir, "contents", "ui")
		_ = os.MkdirAll(qmlDir, 0755)
		_ = os.WriteFile(filepath.Join(qmlDir, "main.qml"), []byte(`Text { text: i18n("Hello") }`), 0644)
		_ = I18nExtractCmd.Flags().Set("no-po", "true")

		// Act
		I18nExtractCmd.Run(I18nExtractCmd, []string{})

		// Assert
		potFile := filepath.Join(projectDir, "translations", "template.pot")
		enPoFile := filepath.Join(projectDir, "translations", "en.po")
		assert.FileExists(t, potFile)
		assert.NoFileExists(t, enPoFile)
	})

	t.Run("error on gettext missing", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		// Save & override PATH to raise error
		oldIsPackageInstalled := utilsIsPackageInstalled
		utilsIsPackageInstalled = func(packageName string) bool {
			return false
		}
		t.Cleanup(func() { utilsIsPackageInstalled = oldIsPackageInstalled })

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		I18nExtractCmd.Run(I18nExtractCmd, []string{})
		_ = w.Close()

		os.Stdout = oldStdout
		color.Output = os.Stdout

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		output := buf.String()

		require.Contains(t, output, "mgettext is not installed. Do you want to install it first?")
	})

	t.Run("successfull install missing gettext package", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		oldIsValidPlasmoid := utilsIsValidPlasmoid
		oldIsPackageInstalled := utilsIsPackageInstalled
		oldInstallPackage := utilsInstallPackage
		oldSurveyAskOne := surveyAskOne

		t.Cleanup(func() {
			utilsIsPackageInstalled = oldIsPackageInstalled
			utilsIsValidPlasmoid = oldIsValidPlasmoid
			utilsInstallPackage = oldInstallPackage
			surveyAskOne = oldSurveyAskOne
		})

		utilsIsPackageInstalled = func(packageName string) bool {
			return false
		}
		utilsIsValidPlasmoid = func() bool {
			return true
		}
		utilsInstallPackage = func(pm string, binName string, pkgNames map[string]string) error {
			return errors.New("install failed")
		}
		surveyAskOne = func(prompt survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
			*(response.(*bool)) = true
			return nil
		}

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		I18nExtractCmd.Run(I18nExtractCmd, []string{})
		_ = w.Close()

		os.Stdout = oldStdout
		color.Output = os.Stdout

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		output := buf.String()

		require.Contains(t, output, "Failed to install gettext")
	})
}

func TestGeneratePoFiles(t *testing.T) {
	t.Run("updates existing .po file", func(t *testing.T) {
		// Arrange
		projectDir, cleanup := tests.SetupTestProject(t)
		defer cleanup()
		cmd.ConfigRC = utils.LoadConfigRC()
		translationsDir := filepath.Join(projectDir, "translations")
		_ = os.MkdirAll(translationsDir, 0755)
		potContent := `msgid ""
msgstr ""
"MIME-Version: 1.0\n"
"Content-Type: text/plain; charset=UTF-8\n"
"Content-Transfer-Encoding: 8bit\n"

msgid "Hello"
msgstr ""`
		poContent := `msgid ""
msgstr ""
"MIME-Version: 1.0\n"
"Content-Type: text/plain; charset=UTF-8\n"
"Content-Transfer-Encoding: 8bit\n"

msgid "Hello"
msgstr "Bonjour"`
		_ = os.WriteFile(filepath.Join(translationsDir, "template.pot"), []byte(potContent), 0644)
		_ = os.WriteFile(filepath.Join(translationsDir, "fr.po"), []byte(poContent), 0644)

		// Act
		err := generatePoFiles(translationsDir)

		// Assert
		require.NoError(t, err)
		finalPo, _ := os.ReadFile(filepath.Join(translationsDir, "fr.po"))
		assert.Contains(t, string(finalPo), `msgstr "Bonjour"`)
	})

	t.Run("handles msginit failure", func(t *testing.T) {
		// Arrange
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()
		cmd.ConfigRC = utils.LoadConfigRC()
		translationsDir := "translations"
		_ = os.MkdirAll(translationsDir, 0755)
		_ = os.WriteFile(filepath.Join(translationsDir, "template.pot"), []byte(""), 0644)
		mockExecCommand(t, "msginit")

		// Act
		err := generatePoFiles(translationsDir)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create")
	})

	t.Run("handles msgmerge failure", func(t *testing.T) {
		// Arrange
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()
		cmd.ConfigRC = utils.LoadConfigRC()
		translationsDir := "translations"
		_ = os.MkdirAll(translationsDir, 0755)
		_ = os.WriteFile(filepath.Join(translationsDir, "template.pot"), []byte(""), 0644)
		_ = os.WriteFile(filepath.Join(translationsDir, "en.po"), []byte(""), 0644)
		mockExecCommand(t, "msgmerge")

		// Act
		err := generatePoFiles(translationsDir)

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to merge")
	})
}

func TestCleanupBackupFiles(t *testing.T) {
	t.Run("removes backup files", func(t *testing.T) {
		// Arrange
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()
		translationsDir := "translations"
		_ = os.MkdirAll(translationsDir, 0755)
		backupFile := filepath.Join(translationsDir, "test.po~")
		_ = os.WriteFile(backupFile, []byte(""), 0644)

		// Act
		err := cleanupBackupFiles(translationsDir)

		// Assert
		require.NoError(t, err)
		assert.NoFileExists(t, backupFile)
	})

	t.Run("handles glob error", func(t *testing.T) {
		// Arrange
		oldGlob := doublestarGlob
		doublestarGlob = func(fs fs.FS, pattern string) ([]string, error) {
			return nil, errors.New("glob error")
		}
		t.Cleanup(func() { doublestarGlob = oldGlob })

		// Act
		err := cleanupBackupFiles(".")

		// Assert
		assert.Error(t, err)
	})

	t.Run("handles remove error", func(t *testing.T) {
		// Arrange
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()
		translationsDir := "translations"
		_ = os.MkdirAll(translationsDir, 0755)
		_ = os.WriteFile(filepath.Join(translationsDir, "test.po~"), []byte(""), 0644)

		oldRemove := osRemove
		osRemove = func(name string) error { return errors.New("remove error") }
		t.Cleanup(func() { osRemove = oldRemove })

		// Act & Assert: Should not return an error, only a warning
		assert.NoError(t, cleanupBackupFiles(translationsDir))
	})
}

func TestHandlePotFileUpdate(t *testing.T) {
	t.Run("renames new file if old does not exist", func(t *testing.T) {
		// Arrange
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()
		oldPath := "test.pot"
		newPath := "test.pot.new"
		_ = os.WriteFile(newPath, []byte(""), 0644)

		// Act
		err := handlePotFileUpdate(oldPath, newPath)

		// Assert
		require.NoError(t, err)
		assert.FileExists(t, oldPath)
		assert.NoFileExists(t, newPath)
	})

	t.Run("removes new file if content is identical", func(t *testing.T) {
		// Arrange
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()
		content := []byte(`
"POT-Creation-Date: 2023-01-01 10:00+0000\n"
msgid "Hello"
msgstr ""
`)
		oldPath := "test.pot"
		newPath := "test.pot.new"
		_ = os.WriteFile(oldPath, content, 0644)
		_ = os.WriteFile(newPath, content, 0644)

		// Act
		err := handlePotFileUpdate(oldPath, newPath)

		// Assert
		require.NoError(t, err)
		assert.FileExists(t, oldPath)
		assert.NoFileExists(t, newPath)
	})

	t.Run("renames new file if content differs", func(t *testing.T) {
		// Arrange
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()
		oldPath := "test.pot"
		newPath := "test.pot.new"

		oldContent := []byte(`
	"POT-Creation-Date: 2023-01-01 10:00+0000\n"
	msgid "Hello"
	msgstr ""
	`)
		newContent := []byte(`
	"POT-Creation-Date: 2024-05-05 12:00+0000\n"
	msgid "Hello World"
	msgstr ""
	`)

		// write different contents
		_ = os.WriteFile(oldPath, oldContent, 0644)
		_ = os.WriteFile(newPath, newContent, 0644)

		// Act
		err := handlePotFileUpdate(oldPath, newPath)

		// Assert
		require.NoError(t, err)
		assert.FileExists(t, oldPath)
		assert.NoFileExists(t, newPath)

		// check the old file got replaced with new content
		finalContent, _ := os.ReadFile(oldPath)
		assert.Contains(t, string(finalContent), "Hello World")
	})

}

func TestStripCreationDate(t *testing.T) {
	input := []byte(`
# Some comment
"POT-Creation-Date: 2023-10-27 10:25+0000\n"
"PO-Revision-Date: YEAR-MO-DA HO:MI+ZONE\n"
`)
	expected := []byte(`
# Some comment
"PO-Revision-Date: YEAR-MO-DA HO:MI+ZONE\n"
`)
	output := stripCreationDate(input)
	assert.Equal(t, strings.TrimSpace(string(expected)), strings.TrimSpace(string(output)))
}

func TestPostProcessPotFile(t *testing.T) {
	t.Run("returns early if file read fails", func(t *testing.T) {
		// Try to read a non-existent file (osReadFile will fail)
		postProcessPotFile("nonexistent.pot", "demo", nil)
		// Nothing to assert, just checking it doesn't panic
	})

	t.Run("default author/email used when authors is nil", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		path := "test.pot"
		content := []byte(`charset=CHARSET
SOME DESCRIPTIVE TITLE.
Copyright (C) YEAR THE PACKAGE'S COPYRIGHT HOLDER
FIRST AUTHOR <EMAIL@ADDRESS>, YEAR.`)
		require.NoError(t, os.WriteFile(path, content, 0644))

		postProcessPotFile(path, "MyApp", nil)

		final, _ := os.ReadFile(path)
		out := string(final)
		assert.Contains(t, out, "charset=UTF-8")
		assert.Contains(t, out, "Translation of MyApp in $__LANGUAGE__$")
		assert.Contains(t, out, "FIRST AUTHOR")
		assert.Contains(t, out, "EMAIL@ADDRESS")
	})

	t.Run("author and email replaced when provided", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		path := "test.pot"
		content := []byte(`charset=CHARSET
SOME DESCRIPTIVE TITLE.
Copyright (C) YEAR THE PACKAGE'S COPYRIGHT HOLDER
FIRST AUTHOR <EMAIL@ADDRESS>, YEAR.`)
		require.NoError(t, os.WriteFile(path, content, 0644))

		authors := []interface{}{
			map[string]interface{}{
				"Name":  "Alice",
				"Email": "alice@example.com",
			},
		}
		postProcessPotFile(path, "CoolApp", authors)

		final, _ := os.ReadFile(path)
		out := string(final)
		assert.Contains(t, out, "Alice <alice@example.com>")
		assert.Contains(t, out, "Translation of CoolApp in $__LANGUAGE__$")
	})

	t.Run("ignores non-list authors value", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		path := "test.pot"
		content := []byte(`charset=CHARSET
SOME DESCRIPTIVE TITLE.
Copyright (C) YEAR THE PACKAGE'S COPYRIGHT HOLDER
FIRST AUTHOR <EMAIL@ADDRESS>, YEAR.`)
		require.NoError(t, os.WriteFile(path, content, 0644))

		postProcessPotFile(path, "OtherApp", "not-a-list")

		final, _ := os.ReadFile(path)
		out := string(final)
		// Still defaults
		assert.Contains(t, out, "FIRST AUTHOR <EMAIL@ADDRESS>")
	})

	t.Run("handles partial author info (only name)", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		path := "test.pot"
		content := []byte(`charset=CHARSET
SOME DESCRIPTIVE TITLE.
Copyright (C) YEAR THE PACKAGE'S COPYRIGHT HOLDER
FIRST AUTHOR <EMAIL@ADDRESS>, YEAR.`)
		require.NoError(t, os.WriteFile(path, content, 0644))

		authors := []interface{}{
			map[string]interface{}{
				"Name": "Bob",
			},
		}
		postProcessPotFile(path, "HalfApp", authors)

		final, _ := os.ReadFile(path)
		out := string(final)
		assert.Contains(t, out, "Bob <EMAIL@ADDRESS>")
	})
}

func TestRunXGettext(t *testing.T) {
	t.Run("fails when Name metadata is missing", func(t *testing.T) {
		// Mock GetDataFromMetadata to return invalid name
		oldGetData := GetDataFromMetadata
		GetDataFromMetadata = func(key string) (interface{}, error) {
			if key == "Name" {
				return nil, fmt.Errorf("missing")
			}
			return "", nil
		}
		t.Cleanup(func() { GetDataFromMetadata = oldGetData })

		err := runXGettext("translations")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing")
	})

	t.Run("fails when Version metadata invalid", func(t *testing.T) {
		oldGetData := GetDataFromMetadata
		GetDataFromMetadata = func(key string) (interface{}, error) {
			if key == "Name" {
				return "MyPlasmoid", nil
			}
			if key == "Version" {
				return nil, fmt.Errorf("invalid")
			}
			return "", nil
		}
		t.Cleanup(func() { GetDataFromMetadata = oldGetData })

		err := runXGettext("translations")
		assert.Error(t, err)
	})

	t.Run("returns nil if no translatable files found", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		oldWalk := filepathWalk
		filepathWalk = func(root string, fn filepath.WalkFunc) error {
			return nil // no files added
		}
		t.Cleanup(func() { filepathWalk = oldWalk })

		oldGetData := GetDataFromMetadata
		GetDataFromMetadata = func(key string) (interface{}, error) {
			switch key {
			case "Name":
				return "MyPlasmoid", nil
			case "Version":
				return "1.0", nil
			default:
				return "", nil
			}
		}
		t.Cleanup(func() { GetDataFromMetadata = oldGetData })

		err := runXGettext("translations")
		assert.NoError(t, err) // should just warn & exit
	})

	t.Run("fails when xgettext command errors", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		// Create dummy source file so it doesn’t exit early
		_ = os.WriteFile("main.qml", []byte(`Text { text: i18n("Hello") }`), 0644)

		// Mock execCommand to fail
		mockExecCommand(t, "xgettext")

		oldGetData := GetDataFromMetadata
		GetDataFromMetadata = func(key string) (interface{}, error) {
			switch key {
			case "Name":
				return "MyPlasmoid", nil
			case "Version":
				return "1.0", nil
			default:
				return "", nil
			}
		}
		t.Cleanup(func() { GetDataFromMetadata = oldGetData })

		err := runXGettext("translations")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "xgettext for source files failed")
	})

	t.Run("fails when potFileNew not created", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		_ = os.WriteFile("main.qml", []byte(`Text { text: i18n("Hello") }`), 0644)

		// Mock runCommand to succeed but we won’t create template.pot.new
		oldRunCmd := runCommand
		runCommand = func(cmd *exec.Cmd) error { return nil }
		t.Cleanup(func() { runCommand = oldRunCmd })

		oldGetData := GetDataFromMetadata
		GetDataFromMetadata = func(key string) (interface{}, error) {
			if key == "Name" {
				return "MyPlasmoid", nil
			}
			if key == "Version" {
				return "1.0", nil
			}
			return "", nil
		}
		t.Cleanup(func() { GetDataFromMetadata = oldGetData })

		err := runXGettext("translations")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no translatable strings")
	})

	t.Run("happy path creates pot file and processes it", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		translations := "translations"
		_ = os.MkdirAll(translations, 0755)

		_ = os.WriteFile("main.qml", []byte(`Text { text: i18n("Hello") }`), 0644)

		// Mock runCommand to simulate creating potFileNew
		oldRunCmd := runCommand
		runCommand = func(cmd *exec.Cmd) error {
			_ = os.WriteFile(filepath.Join(translations, "template.pot.new"), []byte(`dummy`), 0644)
			return nil
		}
		t.Cleanup(func() { runCommand = oldRunCmd })

		oldGetData := GetDataFromMetadata
		GetDataFromMetadata = func(key string) (interface{}, error) {
			switch key {
			case "Name":
				return "MyPlasmoid", nil
			case "Version":
				return "1.0", nil
			case "BugReportUrl":
				return "http://bugs.test", nil
			default:
				return "", nil
			}
		}
		t.Cleanup(func() { GetDataFromMetadata = oldGetData })

		err := runXGettext(translations)
		require.NoError(t, err)

		// final template.pot should exist
		assert.FileExists(t, filepath.Join(translations, "template.pot"))
	})
}
