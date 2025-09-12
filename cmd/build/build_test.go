package build

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/PRASSamin/prasmoid/cmd/i18n"
	"github.com/PRASSamin/prasmoid/tests"
	"github.com/PRASSamin/prasmoid/types"
	"github.com/PRASSamin/prasmoid/utils"
	"github.com/stretchr/testify/assert"
)

type mockFileInfo struct {
	isDir bool
	name  string
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return 0 }
func (m *mockFileInfo) Mode() os.FileMode  { return 0 }
func (m *mockFileInfo) ModTime() time.Time { return time.Now() }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() interface{}   { return nil }

func TestBuildCommand(t *testing.T) {
	t.Run("successful build", func(t *testing.T) {
		// Arrange
		oldAddFileToZip := AddFileToZip
		oldAddDirToZip := AddDirToZip
		t.Cleanup(func() {
			utilsIsValidPlasmoid = utils.IsValidPlasmoid
			i18nCompileI18n = i18n.CompileI18n
			utilsGetDataFromMetadata = utils.GetDataFromMetadata
			osRemoveAll = os.RemoveAll
			osMkdirAll = os.MkdirAll
			osCreate = os.Create
			AddFileToZip = oldAddFileToZip
			AddDirToZip = oldAddDirToZip
		})

		utilsIsValidPlasmoid = func() bool { return true }
		i18nCompileI18n = func(config types.Config, silent bool) error { return nil }
		utilsGetDataFromMetadata = func(key string) (interface{}, error) {
			if key == "Id" {
				return "org.kde.testplasmoid", nil
			}
			if key == "Version" {
				return "1.0.0", nil
			}
			return nil, errors.New("not found")
		}
		osRemoveAll = func(path string) error { return nil }
		osMkdirAll = func(path string, perm os.FileMode) error { return nil }
		osCreate = func(name string) (*os.File, error) {
			// Return a dummy file that writes to a buffer
			return os.NewFile(0, "dummy.zip"), nil
		}
		AddFileToZip = func(zipWriter *zip.Writer, filename string) error { return nil }
		AddDirToZip = func(zipWriter *zip.Writer, baseDir string) error { return nil }

		// Act
		err := BuildPlasmoid()

		// Assert
		assert.NoError(t, err)
	})

	t.Run("build without valid plasmoid", func(t *testing.T) {
		// Arrange
		t.Cleanup(func() { utilsIsValidPlasmoid = utils.IsValidPlasmoid })
		utilsIsValidPlasmoid = func() bool { return false }

		// Act
		err := BuildPlasmoid()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "current directory is not a valid plasmoid")
	})

	t.Run("build without id and version", func(t *testing.T) {
		// Arrange
		t.Cleanup(func() {
			utilsIsValidPlasmoid = utils.IsValidPlasmoid
			i18nCompileI18n = i18n.CompileI18n
			utilsGetDataFromMetadata = utils.GetDataFromMetadata
		})
		utilsIsValidPlasmoid = func() bool { return true }
		i18nCompileI18n = func(config types.Config, silent bool) error { return nil }
		utilsGetDataFromMetadata = func(key string) (interface{}, error) {
			return nil, errors.New("metadata error")
		}

		// Act
		err := BuildPlasmoid()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid metadata")
	})

	t.Run("build with failing translation compilation", func(t *testing.T) {
		oldAddFileToZip := AddFileToZip
		oldAddDirToZip := AddDirToZip
		// Arrange
		t.Cleanup(func() {
			i18nCompileI18n = i18n.CompileI18n
			// Restore other mocks from successful build
			utilsIsValidPlasmoid = utils.IsValidPlasmoid
			utilsGetDataFromMetadata = utils.GetDataFromMetadata
			osRemoveAll = os.RemoveAll
			osMkdirAll = os.MkdirAll
			osCreate = os.Create
			AddFileToZip = oldAddFileToZip
			AddDirToZip = oldAddDirToZip
		})

		utilsIsValidPlasmoid = func() bool { return true }
		i18nCompileI18n = func(config types.Config, silent bool) error { return errors.New("i18n error") }
		utilsGetDataFromMetadata = func(key string) (interface{}, error) {
			if key == "Id" {
				return "org.kde.testplasmoid", nil
			}
			if key == "Version" {
				return "1.0.0", nil
			}
			return nil, nil
		}
		osRemoveAll = func(path string) error { return nil }
		osMkdirAll = func(path string, perm os.FileMode) error { return nil }
		osCreate = func(name string) (*os.File, error) { return os.NewFile(0, "dummy.zip"), nil }
		AddFileToZip = func(zipWriter *zip.Writer, filename string) error { return nil }
		AddDirToZip = func(zipWriter *zip.Writer, baseDir string) error { return nil }

		// Act
		err := BuildPlasmoid()

		// Assert
		assert.NoError(t, err, "Build should succeed even if translation fails")
	})

	t.Run("failed to clear existing build dir", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()
		// Arrange
		t.Cleanup(func() { osRemoveAll = os.RemoveAll })
		osRemoveAll = func(path string) error { return errors.New("clean error") }

		// Act
		err := BuildPlasmoid()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to clean build dir")
	})

	t.Run("failed to create build dir", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()
		// Arrange
		t.Cleanup(func() { osMkdirAll = os.MkdirAll })
		osRemoveAll = func(path string) error { return nil } // This should pass
		osMkdirAll = func(path string, perm os.FileMode) error { return errors.New("mkdir error") }

		// Act
		err := BuildPlasmoid()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create build dir")
	})
}

func TestAddFileToZip(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()
		// Arrange
		t.Cleanup(func() {
			osOpen = os.Open
			ioCopy = io.Copy
		})

		var copied bool
		osOpen = func(name string) (*os.File, error) {
			return os.NewFile(0, "dummy.txt"), nil
		}
		ioCopy = func(dst io.Writer, src io.Reader) (int64, error) {
			copied = true
			return 0, nil
		}

		buf := new(bytes.Buffer)
		zipWriter := zip.NewWriter(buf)

		// Act
		err := AddFileToZip(zipWriter, "test.txt")
		assert.NoError(t, err)

		err = zipWriter.Close()
		assert.NoError(t, err)
		assert.True(t, copied)
	})

	t.Run("file does not exist", func(t *testing.T) {
		// Arrange
		t.Cleanup(func() { osOpen = os.Open })
		osOpen = func(name string) (*os.File, error) {
			return nil, os.ErrNotExist
		}

		buf := new(bytes.Buffer)
		zipWriter := zip.NewWriter(buf)

		// Act
		err := AddFileToZip(zipWriter, "non-existent-file.txt")

		// Assert
		assert.Error(t, err)
		assert.Equal(t, os.ErrNotExist, err)
	})
}

func TestAddDirToZip(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()
		// Arrange
		t.Cleanup(func() { filepathWalk = filepath.Walk })

		var walkCalled bool
		filepathWalk = func(root string, walkFn filepath.WalkFunc) error {
			walkCalled = true
			// Simulate walking a directory with one file
			if err := walkFn("contents", &mockFileInfo{isDir: true, name: "contents"}, nil); err != nil {
				return err
			}
			if err := walkFn("contents/ui/main.qml", &mockFileInfo{isDir: false, name: "ui.qml"}, nil); err != nil {
				return err
			}
			return nil
		}

		buf := new(bytes.Buffer)
		zipWriter := zip.NewWriter(buf)

		// Act
		err := AddDirToZip(zipWriter, "contents")

		// Assert
		assert.NoError(t, err)
		assert.True(t, walkCalled)
	})

	t.Run("directory does not exist", func(t *testing.T) {
		// Arrange
		t.Cleanup(func() { filepathWalk = filepath.Walk })
		expectedErr := errors.New("walk error")
		filepathWalk = func(root string, walkFn filepath.WalkFunc) error {
			return expectedErr
		}

		buf := new(bytes.Buffer)
		zipWriter := zip.NewWriter(buf)

		// Act
		err := AddDirToZip(zipWriter, "non-existent-dir")

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}
