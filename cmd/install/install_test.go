package install

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/PRASSamin/prasmoid/tests"
	"github.com/PRASSamin/prasmoid/utils"
	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInstallCmd(t *testing.T) {
	t.Run("not a valid plasmoid", func(t *testing.T) {
		// Arrange
		_, _, cleanup := tests.SetupTestEnvironment(t)
		defer cleanup()
		// remove metadata.json to make it invalid
		_ = os.Remove("metadata.json")

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		color.Output = w

		// Act
		InstallCmd.Run(InstallCmd, []string{})
		_ = w.Close()

		// Assert
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		assert.Contains(t, buf.String(), "Current directory is not a valid plasmoid.")
	})

	t.Run("install fails", func(t *testing.T) {
		// Arrange
		_, _, cleanup := tests.SetupTestEnvironment(t)
		defer cleanup()

		// Mock InstallPlasmoid to fail
		oldInstall := InstallPlasmoid
		InstallPlasmoid = func() error { return errors.New("install error") }
		defer func() { InstallPlasmoid = oldInstall }()

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		color.Output = w

		// Act
		InstallCmd.Run(InstallCmd, []string{})
		_ = w.Close()

		// Assert
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		assert.Contains(t, buf.String(), "Failed to install plasmoid:")
	})

	t.Run("install succeeds", func(t *testing.T) {
		// Arrange
		_, _, cleanup := tests.SetupTestEnvironment(t)
		defer cleanup()

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		color.Output = w

		// Act
		InstallCmd.Run(InstallCmd, []string{})
		_ = w.Close()

		// Assert
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		assert.Contains(t, buf.String(), "Plasmoid installed successfully")
	})
}

func TestInstallPlasmoid(t *testing.T) {
	t.Run("successful first time install", func(t *testing.T) {
		// Arrange
		_, _, cleanup := tests.SetupTestEnvironment(t)
		defer cleanup()

		// Act
		err := InstallPlasmoid()

		// Assert
		require.NoError(t, err)
		dest, _ := utils.GetDevDest()
		_, err = os.Stat(filepath.Join(dest, "metadata.json"))
		assert.NoError(t, err, "metadata.json should exist in destination")
	})

	t.Run("successful reinstall", func(t *testing.T) {
		// Arrange
		_, _, cleanup := tests.SetupTestEnvironment(t)
		defer cleanup()

		// Pre-install a version
		require.NoError(t, InstallPlasmoid())

		// Act
		err := InstallPlasmoid()

		// Assert
		require.NoError(t, err)
	})

	t.Run("fails when IsInstalled errors", func(t *testing.T) {
		// Arrange
		_, _, cleanup := tests.SetupTestEnvironment(t)
		defer cleanup()

		// Mock
		oldIsInstalled := utilsIsInstalled
		utilsIsInstalled = func() (bool, string, error) { return false, "", errors.New("isinstalled error") }
		defer func() { utilsIsInstalled = oldIsInstalled }()

		// Act & Assert
		assert.Error(t, InstallPlasmoid())
	})

	t.Run("warns when removing existing install fails", func(t *testing.T) {
		// Arrange
		_, _, cleanup := tests.SetupTestEnvironment(t)
		defer cleanup()

		// Mock
		oldRemoveAll := osRemoveAll
		osRemoveAll = func(path string) error { return errors.New("remove error") }
		defer func() { osRemoveAll = oldRemoveAll }()

		// Act & Assert
		assert.NoError(t, InstallPlasmoid(), "Should not fail, only warn")
	})

	t.Run("fails when creating install dir errors", func(t *testing.T) {
		// Arrange
		_, _, cleanup := tests.SetupTestEnvironment(t)
		defer cleanup()

		// Mock
		oldMkdirAll := osMkdirAll
		osMkdirAll = func(path string, perm os.FileMode) error { return errors.New("mkdir error") }
		defer func() { osMkdirAll = oldMkdirAll }()

		// Act & Assert
		assert.Error(t, InstallPlasmoid())
	})

	t.Run("fails when reading metadata errors", func(t *testing.T) {
		// Arrange
		_, _, cleanup := tests.SetupTestEnvironment(t)
		defer cleanup()

		// Mock
		oldReadFile := osReadFile
		osReadFile = func(name string) ([]byte, error) {
			if filepath.Base(name) == "metadata.json" {
				return nil, errors.New("read metadata error")
			}
			return oldReadFile(name)
		}
		defer func() { osReadFile = oldReadFile }()

		// Act & Assert
		assert.Error(t, InstallPlasmoid())
	})

	t.Run("fails when writing metadata errors", func(t *testing.T) {
		// Arrange
		_, _, cleanup := tests.SetupTestEnvironment(t)
		defer cleanup()

		// Mock
		oldWriteFile := osWriteFile
		osWriteFile = func(name string, data []byte, perm os.FileMode) error {
			if filepath.Base(name) == "metadata.json" {
				return errors.New("write metadata error")
			}
			return oldWriteFile(name, data, perm)
		}
		defer func() { osWriteFile = oldWriteFile }()

		// Act & Assert
		assert.Error(t, InstallPlasmoid())
	})

	t.Run("fails when copying contents errors", func(t *testing.T) {
		// Arrange
		tempDir, _, cleanup := tests.SetupTestEnvironment(t)
		defer cleanup()

		_ = os.RemoveAll(filepath.Join(tempDir, "contents"))

		// Act & Assert
		assert.Error(t, InstallPlasmoid())
	})
}

func TestCopyDir(t *testing.T) {
	t.Run("fails to create dest dir", func(t *testing.T) {
		// Mock
		oldMkdirAll := osMkdirAll
		osMkdirAll = func(path string, perm os.FileMode) error { return errors.New("mkdir error") }
		defer func() { osMkdirAll = oldMkdirAll }()

		// Act & Assert
		assert.Error(t, copyDir("a", "b"))
	})

	t.Run("fails to read src dir", func(t *testing.T) {
		// Mock
		oldReadDir := osReadDir
		osReadDir = func(name string) ([]fs.DirEntry, error) { return nil, errors.New("readdir error") }
		defer func() { osReadDir = oldReadDir }()

		// Act & Assert
		assert.Error(t, copyDir("a", "b"))
	})

	t.Run("fails to read src file", func(t *testing.T) {
		// Arrange
		_, _, cleanup := tests.SetupTestEnvironment(t)
		defer cleanup()

		// Mock
		oldReadFile := osReadFile
		osReadFile = func(name string) ([]byte, error) { return nil, errors.New("readfile error") }
		defer func() { osReadFile = oldReadFile }()

		// Act & Assert
		assert.Error(t, copyDir("contents", "dest"))
	})

	t.Run("fails to write dest file", func(t *testing.T) {
		// Arrange
		_, _, cleanup := tests.SetupTestEnvironment(t)
		defer cleanup()

		// Mock
		oldWriteFile := osWriteFile
		osWriteFile = func(name string, data []byte, perm os.FileMode) error { return errors.New("writefile error") }
		defer func() { osWriteFile = oldWriteFile }()

		// Act & Assert
		assert.Error(t, copyDir("contents", "dest"))
	})
}