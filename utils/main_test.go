package utils

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/AlecAivazis/survey/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsQmlFile(t *testing.T) {
	assert.True(t, IsQmlFile("main.qml"))
	assert.True(t, IsQmlFile("Component.qml"))
	assert.False(t, IsQmlFile("main.qml.txt"))
	assert.False(t, IsQmlFile("main.js"))
	assert.False(t, IsQmlFile(""))
}

func TestEnsureStringAndValid(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		s, err := EnsureStringAndValid("test", "value", nil)
		assert.NoError(t, err)
		assert.Equal(t, "value", s)
	})

	t.Run("error from input", func(t *testing.T) {
		expectedErr := errors.New("input error")
		s, err := EnsureStringAndValid("test", "value", expectedErr)
		assert.Equal(t, expectedErr, err)
		assert.Equal(t, "", s)
	})

	t.Run("value is not a string", func(t *testing.T) {
		s, err := EnsureStringAndValid("test", 123, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "test value is not a string")
		assert.Equal(t, "", s)
	})
}

func TestGetDataFromMetadata(t *testing.T) {
	setup := func(t *testing.T, content string) {
		tmpDir := t.TempDir()
		if content != "" {
			metaFile := filepath.Join(tmpDir, "metadata.json")
			require.NoError(t, os.WriteFile(metaFile, []byte(content), 0644))
		}

		oldWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tmpDir))
		t.Cleanup(func() { require.NoError(t, os.Chdir(oldWd)) })
	}

	t.Run("success", func(t *testing.T) {
		setup(t, `{"KPlugin": {"Id": "my-plasmoid"}}`)
		id, err := GetDataFromMetadata("Id")
		assert.NoError(t, err)
		assert.Equal(t, "my-plasmoid", id)
	})

	t.Run("file not found", func(t *testing.T) {
		setup(t, "")
		_, err := GetDataFromMetadata("Id")
		assert.Error(t, err)
	})

	t.Run("invalid json", func(t *testing.T) {
		setup(t, `{"KPlugin": {"Id": "my-plasmoid"}`)
		_, err := GetDataFromMetadata("Id")
		assert.Error(t, err)
	})

	t.Run("no KPlugin section", func(t *testing.T) {
		setup(t, `{}`)
		_, err := GetDataFromMetadata("Id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "KPlugin section not found")
	})

	t.Run("KPlugin not a map", func(t *testing.T) {
		setup(t, `{"KPlugin": "invalid"}`)
		_, err := GetDataFromMetadata("Id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "KPlugin section has unexpected structure")
	})

	t.Run("key not found", func(t *testing.T) {
		setup(t, `{"KPlugin": {"Name": "My Plasmoid"}}`)
		_, err := GetDataFromMetadata("Id")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "key 'Id' not found in KPlugin")
	})
}

func TestUpdateMetadata(t *testing.T) {
	setup := func(t *testing.T, initialContent string) string {
		tmpDir := t.TempDir()
		metaFile := filepath.Join(tmpDir, "metadata.json")
		if initialContent != "" {
			require.NoError(t, os.WriteFile(metaFile, []byte(initialContent), 0644))
		}

		oldWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tmpDir))
		t.Cleanup(func() { require.NoError(t, os.Chdir(oldWd)) })
		return metaFile
	}

	t.Run("update existing key", func(t *testing.T) {
		metaFile := setup(t, `{"KPlugin": {"Name": "Old Name"}}`)
		err := UpdateMetadata("Name", "New Name")
		assert.NoError(t, err)

		data, _ := os.ReadFile(metaFile)
		var meta map[string]map[string]interface{}
		_ = json.Unmarshal(data, &meta)
		assert.Equal(t, "New Name", meta["KPlugin"]["Name"])
	})

	t.Run("add new key", func(t *testing.T) {
		metaFile := setup(t, `{"KPlugin": {"Name": "My Plasmoid"}}`)
		err := UpdateMetadata("Version", "1.1")
		assert.NoError(t, err)

		data, _ := os.ReadFile(metaFile)
		var meta map[string]map[string]interface{}
		_ = json.Unmarshal(data, &meta)
		assert.Equal(t, "1.1", meta["KPlugin"]["Version"])
	})

	t.Run("add new section", func(t *testing.T) {
		metaFile := setup(t, `{}`)
		err := UpdateMetadata("Name", "My Plasmoid", "KPlugin")
		assert.NoError(t, err)

		data, _ := os.ReadFile(metaFile)
		var meta map[string]map[string]interface{}
		_ = json.Unmarshal(data, &meta)
		assert.Equal(t, "My Plasmoid", meta["KPlugin"]["Name"])
	})

	t.Run("update root key", func(t *testing.T) {
		metaFile := setup(t, `{"Name": "Old Name"}`)
		err := UpdateMetadata("Name", "New Name", ".")
		assert.NoError(t, err)

		data, _ := os.ReadFile(metaFile)
		var meta map[string]interface{}
		_ = json.Unmarshal(data, &meta)
		assert.Equal(t, "New Name", meta["Name"])
	})

	t.Run("file not found", func(t *testing.T) {
		setup(t, "")
		err := UpdateMetadata("Name", "New Name")
		assert.Error(t, err)
	})
}

func TestIsValidPlasmoid(t *testing.T) {
	setup := func(t *testing.T, createMeta bool, createContents bool) {
		tmpDir := t.TempDir()
		if createMeta {
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "metadata.json"), []byte("{}"), 0644))
		}
		if createContents {
			require.NoError(t, os.Mkdir(filepath.Join(tmpDir, "contents"), 0755))
		}

		oldWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tmpDir))
		t.Cleanup(func() { require.NoError(t, os.Chdir(oldWd)) })
	}

	t.Run("valid", func(t *testing.T) {
		setup(t, true, true)
		assert.True(t, IsValidPlasmoid())
	})

	t.Run("missing metadata.json", func(t *testing.T) {
		setup(t, false, true)
		assert.False(t, IsValidPlasmoid())
	})

	t.Run("missing contents dir", func(t *testing.T) {
		setup(t, true, false)
		assert.False(t, IsValidPlasmoid())
	})
}

func TestGetDevDest(t *testing.T) {
	setup := func(t *testing.T, metaContent string) (homeDir string) {
		tmpDir := t.TempDir()
		homeDir = t.TempDir()
		t.Setenv("HOME", homeDir)

		if metaContent != "" {
			metaFile := filepath.Join(tmpDir, "metadata.json")
			require.NoError(t, os.WriteFile(metaFile, []byte(metaContent), 0644))
		}

		oldWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tmpDir))
		t.Cleanup(func() { require.NoError(t, os.Chdir(oldWd)) })
		return homeDir
	}

	t.Run("success", func(t *testing.T) {
		homeDir := setup(t, `{"KPlugin": {"Id": "my-plasmoid"}}`)
		dest, err := GetDevDest()
		assert.NoError(t, err)
		expectedDest := filepath.Join(homeDir, ".local/share/plasma/plasmoids/my-plasmoid")
		assert.Equal(t, expectedDest, dest)
		// Check if parent dir was created
		parentDir := filepath.Dir(expectedDest)
		_, err = os.Stat(parentDir)
		assert.NoError(t, err)
	})

	t.Run("GetDataFromMetadata fails", func(t *testing.T) {
		setup(t, "")
		_, err := GetDevDest()
		assert.Error(t, err)
	})
}

func TestIsLinked(t *testing.T) {
	setup := func(t *testing.T, metaContent string, createLink bool) {
		tmpDir := t.TempDir()
		homeDir := t.TempDir()
		t.Setenv("HOME", homeDir)

		if metaContent != "" {
			metaFile := filepath.Join(tmpDir, "metadata.json")
			require.NoError(t, os.WriteFile(metaFile, []byte(metaContent), 0644))
		}

		if createLink {
			destDir := filepath.Join(homeDir, ".local/share/plasma/plasmoids")
			require.NoError(t, os.MkdirAll(destDir, 0755))
			require.NoError(t, os.Symlink(tmpDir, filepath.Join(destDir, "my-plasmoid")))
		}

		oldWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tmpDir))
		t.Cleanup(func() { require.NoError(t, os.Chdir(oldWd)) })
	}

	t.Run("linked", func(t *testing.T) {
		setup(t, `{"KPlugin": {"Id": "my-plasmoid"}}`, true)
		assert.True(t, IsLinked())
	})

	t.Run("not linked", func(t *testing.T) {
		setup(t, `{"KPlugin": {"Id": "my-plasmoid"}}`, false)
		assert.False(t, IsLinked())
	})

	t.Run("GetDevDest fails", func(t *testing.T) {
		setup(t, "", false)
		assert.False(t, IsLinked())
	})
}

func TestIsInstalled(t *testing.T) {
	setup := func(t *testing.T, metaContent string, installInUser bool) (homeDir string) {
		tmpDir := t.TempDir()
		homeDir = t.TempDir()
		t.Setenv("HOME", homeDir)

		if metaContent != "" {
			metaFile := filepath.Join(tmpDir, "metadata.json")
			require.NoError(t, os.WriteFile(metaFile, []byte(metaContent), 0644))
		}

		if installInUser {
			userPath := filepath.Join(homeDir, ".local/share/plasma/plasmoids", "my-plasmoid")
			require.NoError(t, os.MkdirAll(userPath, 0755))
		}

		oldWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tmpDir))
		t.Cleanup(func() { require.NoError(t, os.Chdir(oldWd)) })
		return homeDir
	}

	t.Run("installed in user path", func(t *testing.T) {
		homeDir := setup(t, `{"KPlugin": {"Id": "my-plasmoid"}}`, true)
		installed, path, err := IsInstalled()
		assert.NoError(t, err)
		assert.True(t, installed)
		expectedPath := filepath.Join(homeDir, ".local/share/plasma/plasmoids", "my-plasmoid")
		assert.Equal(t, expectedPath, path)
	})

	t.Run("not installed", func(t *testing.T) {
		setup(t, `{"KPlugin": {"Id": "my-plasmoid"}}`, false)
		installed, _, err := IsInstalled()
		assert.NoError(t, err)
		assert.False(t, installed)
	})

	t.Run("GetDataFromMetadata fails", func(t *testing.T) {
		setup(t, "", false)
		_, _, err := IsInstalled()
		assert.Error(t, err)
	})
}

func TestIsPackageInstalled(t *testing.T) {
	setup := func(t *testing.T, executables ...string) {
		tmpDir := t.TempDir()
		for _, exec := range executables {
			err := os.WriteFile(filepath.Join(tmpDir, exec), []byte("#!/bin/sh\n"), 0755)
			require.NoError(t, err)
		}
		t.Setenv("PATH", tmpDir)
	}

	t.Run("package is installed", func(t *testing.T) {
		setup(t, "my-package")
		assert.True(t, IsPackageInstalled("my-package"))
	})

	t.Run("package is not installed", func(t *testing.T) {
		setup(t)
		assert.False(t, IsPackageInstalled("my-package"))
	})
}

func TestAskForLocales(t *testing.T) {
	originalSurveyAskOne := surveyAskOne
	defer func() {
		surveyAskOne = originalSurveyAskOne
	}()

	t.Run("success", func(t *testing.T) {
		surveyAskOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
			s, ok := response.(*[]string)
			require.True(t, ok)
			*s = []string{"en (English)", "fr (French)"}
			return nil
		}

		locales := AskForLocales()
		assert.Equal(t, []string{"en", "fr"}, locales)
	})

	t.Run("survey returns error", func(t *testing.T) {
		surveyAskOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
			return errors.New("survey error")
		}

		locales := AskForLocales()
		assert.Nil(t, locales)
	})
}

func TestLoadConfigRC(t *testing.T) {
	setup := func(t *testing.T, content *string) {
		tmpDir := t.TempDir()
		if content != nil {
			configFile := filepath.Join(tmpDir, "prasmoid.config.js")
			require.NoError(t, os.WriteFile(configFile, []byte(*content), 0644))
		}

		oldWd, err := os.Getwd()
		require.NoError(t, err)
		require.NoError(t, os.Chdir(tmpDir))
		t.Cleanup(func() { require.NoError(t, os.Chdir(oldWd)) })
	}

	t.Run("success", func(t *testing.T) {
		configContent := `var config = { i18n: { locales: ["en", "de"] } };`
		setup(t, &configContent)
		config := LoadConfigRC()
		assert.Equal(t, []string{"en", "de"}, config.I18n.Locales)
	})

	t.Run("file not found", func(t *testing.T) {
		setup(t, nil)
		config := LoadConfigRC()
		assert.Equal(t, []string{"en"}, config.I18n.Locales) // default
	})

	t.Run("invalid js", func(t *testing.T) {
		configContent := `var config = {`
		setup(t, &configContent)
		config := LoadConfigRC()
		assert.Equal(t, []string{"en"}, config.I18n.Locales) // default
	})

	t.Run("no config object", func(t *testing.T) {
		configContent := `var myconfig = {};`
		setup(t, &configContent)
		config := LoadConfigRC()
		assert.Equal(t, []string{"en"}, config.I18n.Locales) // default
	})
}


func TestEnsureBinaryLinked(t *testing.T) {
	originalExecLookPath := execLookPath
	originalExecCommand := execCommand
	originalOsLstat := osLstat
	originalOsSymlink := osSymlink
	defer func() {
		execLookPath = originalExecLookPath
		execCommand = originalExecCommand
		osLstat = originalOsLstat
		osSymlink = originalOsSymlink
	}()

	t.Run("binary already in path", func(t *testing.T) {
		execLookPath = func(file string) (string, error) {
			return "/usr/bin/my-bin", nil
		}
		err := ensureBinaryLinked("my-bin", "/usr/local/bin")
		assert.NoError(t, err)
	})

	t.Run("binary found and symlinked", func(t *testing.T) {
		execLookPath = func(file string) (string, error) {
			return "", errors.New("not found")
		}
		execCommand = func(name string, args ...string) *exec.Cmd {
			return exec.Command("echo", "/opt/my-bin/my-bin")
		}
		osLstat = func(name string) (os.FileInfo, error) {
			return nil, os.ErrNotExist
		}
		var symlinkCalled bool
		osSymlink = func(oldname, newname string) error {
			symlinkCalled = true
			assert.Equal(t, "/opt/my-bin/my-bin", oldname)
			assert.Equal(t, "/usr/local/bin/my-bin", newname)
			return nil
		}

		err := ensureBinaryLinked("my-bin", "/usr/local/bin")
		assert.NoError(t, err)
		assert.True(t, symlinkCalled)
	})

	t.Run("binary found but symlink exists", func(t *testing.T) {
		execLookPath = func(file string) (string, error) {
			return "", errors.New("not found")
		}
		execCommand = func(name string, args ...string) *exec.Cmd {
			return exec.Command("echo", "/opt/my-bin/my-bin")
		}
		osLstat = func(name string) (os.FileInfo, error) {
			return nil, nil // file exists
		}
		var symlinkCalled bool
		osSymlink = func(oldname, newname string) error {
			symlinkCalled = true
			return nil
		}

		err := ensureBinaryLinked("my-bin", "/usr/local/bin")
		assert.NoError(t, err)
		assert.False(t, symlinkCalled)
	})

	t.Run("binary not found", func(t *testing.T) {
		execLookPath = func(file string) (string, error) {
			return "", errors.New("not found")
		}
		execCommand = func(name string, args ...string) *exec.Cmd {
			return exec.Command("echo", "") // empty output
		}

		err := ensureBinaryLinked("my-bin", "/usr/local/bin")
		assert.Error(t, err)
	})

	t.Run("find command fails", func(t *testing.T) {
		execLookPath = func(file string) (string, error) {
			return "", errors.New("not found")
		}
		execCommand = func(name string, args ...string) *exec.Cmd {
			return exec.Command("false") // command that fails
		}

		err := ensureBinaryLinked("my-bin", "/usr/local/bin")
		assert.Error(t, err)
	})

	t.Run("symlink fails", func(t *testing.T) {
		execLookPath = func(file string) (string, error) {
			return "", errors.New("not found")
		}
		execCommand = func(name string, args ...string) *exec.Cmd {
			return exec.Command("echo", "/opt/my-bin/my-bin")
		}
		osLstat = func(name string) (os.FileInfo, error) {
			return nil, os.ErrNotExist
		}
		osSymlink = func(oldname, newname string) error {
			return errors.New("symlink error")
		}

		err := ensureBinaryLinked("my-bin", "/usr/local/bin")
		assert.Error(t, err)
	})
}