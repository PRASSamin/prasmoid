package i18n

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/PRASSamin/prasmoid/tests"
	"github.com/PRASSamin/prasmoid/types"
	"github.com/PRASSamin/prasmoid/utils"
	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestI18nCompileCommand(t *testing.T) {
	// Set up a temporary project
	t.Run("successfully compiles .po files", func(t *testing.T) {
		projectDir, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		// Create a dummy config
		config := types.Config{
			I18n: types.ConfigI18n{
				Dir:     "translations",
				Locales: []string{"en"},
			},
		}

		// Create a dummy .po file
		poDir := filepath.Join(projectDir, config.I18n.Dir)
		_ = os.MkdirAll(poDir, 0755)
		poContent := `
msgid "Hello World"
msgstr "Hello World"
`
		if err := os.WriteFile(filepath.Join(poDir, "en.po"), []byte(poContent), 0644); err != nil {
			t.Fatalf("Failed to write en.po: %v", err)
		}

		_ = I18nCompileCmd.Flags().Set("restart", "true")
		// 2. Execute the CompileMessages function
		I18nCompileCmd.Run(I18nCompileCmd, []string{})

		// 3. Verify the output
		plasmoidId, _ := utils.GetDataFromMetadata("Id")
		moFile := filepath.Join(projectDir, "contents", "locale", "en", "LC_MESSAGES", "plasma_applet_"+plasmoidId.(string)+".mo")

		if _, err := os.Stat(moFile); os.IsNotExist(err) {
			t.Fatalf("Expected .mo file to be created at %s, but it was not", moFile)
		}

		// Check that the .mo file is not empty
		info, _ := os.Stat(moFile)
		if info.Size() == 0 {
			t.Errorf("Expected .mo file to not be empty, but it was")
		}
	})

	t.Run("invalid plasmoid", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()
		_ = os.Remove("metadata.json")

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		I18nCompileCmd.Run(I18nCompileCmd, []string{})

		require.NoError(t, w.Close())
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		color.Output = os.Stdout
		output := buf.String()
		assert.Contains(t, output, "Current directory is not a valid plasmoid")
	})

	t.Run("error on gettext missing", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		// Save & override PATH to raise error
		oldPath := os.Getenv("PATH")
		defer os.Setenv("PATH", oldPath)
		_ = os.Setenv("PATH", "/nonexistent")

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		confirm = true
		I18nCompileCmd.Run(I18nCompileCmd, []string{})
		_ = w.Close()

		os.Stdout = oldStdout
		color.Output = os.Stdout

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		output := buf.String()

		require.Contains(t, output, "Failed to install gettext")
	})

}

func TestCompileI18n(t *testing.T) {
	t.Run("no po files found", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		config := types.Config{
			I18n: types.ConfigI18n{
				Dir:     "translations",
				Locales: []string{"en"},
			},
		}

		err := CompileI18n(config, true)
		assert.NoError(t, err)
	})

	t.Run("error getting plasmoid id", func(t *testing.T) {
		tmpDir, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		metadataPath := filepath.Join(tmpDir, "metadata.json")
		data, _ := json.MarshalIndent(map[string]interface{}{}, "", "  ")
		require.NoError(t, os.WriteFile(metadataPath, data, 0644))

		config := types.Config{}
		err := CompileI18n(config, true)
		assert.Error(t, err)
	})

	t.Run("glob returns error", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		oldFilepathGlob := filepathGlob
		filepathGlob = func(pattern string) ([]string, error) {
			return nil, errors.New("glob error")
		}
		t.Cleanup(func() { filepathGlob = oldFilepathGlob })

		config := types.Config{}
		err := CompileI18n(config, true)
		assert.Error(t, err)
	})

	t.Run("skip po file not in config", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		config := types.Config{
			I18n: types.ConfigI18n{
				Dir:     "translations",
				Locales: []string{"fr"},
			},
		}

		poDir := filepath.Join(".", config.I18n.Dir)
		_ = os.MkdirAll(poDir, 0755)
		_ = os.WriteFile(filepath.Join(poDir, "en.po"), []byte(""), 0644)

		err := CompileI18n(config, true)
		assert.NoError(t, err)
	})

	t.Run("mkdirall returns error", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		config := types.Config{
			I18n: types.ConfigI18n{
				Dir:     "translations",
				Locales: []string{"en"},
			},
		}

		poDir := filepath.Join(".", config.I18n.Dir)
		_ = os.MkdirAll(poDir, 0755)
		_ = os.WriteFile(filepath.Join(poDir, "en.po"), []byte(""), 0644)

		oldOsMkdirAll := osMkdirAll
		osMkdirAll = func(path string, perm os.FileMode) error {
			return errors.New("mkdir error")
		}
		t.Cleanup(func() { osMkdirAll = oldOsMkdirAll })

		err := CompileI18n(config, true)
		assert.Error(t, err)
	})

	t.Run("msgfmt command fails", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		config := types.Config{
			I18n: types.ConfigI18n{
				Dir:     "translations",
				Locales: []string{"en"},
			},
		}

		poDir := filepath.Join(".", config.I18n.Dir)
		_ = os.MkdirAll(poDir, 0755)
		_ = os.WriteFile(filepath.Join(poDir, "en.po"), []byte(""), 0644)

		oldExecCommand := execCommand
		execCommand = func(name string, arg ...string) *exec.Cmd {
			return exec.Command("non-existent-command")
		}
		t.Cleanup(func() { execCommand = oldExecCommand })

		err := CompileI18n(config, true)
		assert.Error(t, err)
	})
}

func TestRestartPlasmashell(t *testing.T) {
	t.Run("killall fails", func(t *testing.T) {
		oldExecCommand := execCommand
		execCommand = func(name string, arg ...string) *exec.Cmd {
			if name == "killall" {
				return exec.Command("non-existent-command")
			}
			return exec.Command(name, arg...)
		}
		t.Cleanup(func() { execCommand = oldExecCommand })

		// Should not return an error, just print a warning
		err := restartPlasmashell()
		assert.NoError(t, err)
	})

	t.Run("kstart5 not found, fallback to kstart", func(t *testing.T) {
		var calledWith string
		oldExecCommand := execCommand
		execCommand = func(name string, arg ...string) *exec.Cmd {
			calledWith = name
			return exec.Command("echo") // return a command that exists and does nothing
		}
		t.Cleanup(func() { execCommand = oldExecCommand })

		oldExecLookPath := execLookPath
		execLookPath = func(file string) (string, error) {
			if file == "kstart5" {
				return "", errors.New("not found")
			}
			return "/bin/kstart", nil
		}
		t.Cleanup(func() { execLookPath = oldExecLookPath })

		err := restartPlasmashell()
		assert.NoError(t, err)
		assert.Equal(t, "kstart", calledWith)
	})
}