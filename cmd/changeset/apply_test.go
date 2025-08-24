package changeset

import (
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/PRASSamin/prasmoid/utils"
	"github.com/fatih/color"
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

func TestUpdateChangelog(t *testing.T) {
	t.Run("create new changelog", func(t *testing.T) {
		// Arrange
		t.Cleanup(func() {
			osStat = os.Stat
			osWriteFile = os.WriteFile
		})

		osStat = func(name string) (os.FileInfo, error) {
			return nil, os.ErrNotExist
		}
		var writtenContent string
		osWriteFile = func(name string, data []byte, perm os.FileMode) error {
			writtenContent = string(data)
			return nil
		}

		// Act
		err := UpdateChangelog("1.0.0", "2025-01-01", "Initial release")

		// Assert
		assert.NoError(t, err)
		assert.Contains(t, writtenContent, "# CHANGELOG")
		assert.Contains(t, writtenContent, "[v1.0.0] - 2025-01-01")
		assert.Contains(t, writtenContent, "Initial release")
	})

	t.Run("append to existing changelog", func(t *testing.T) {
		// Arrange
		t.Cleanup(func() {
			osStat = os.Stat
			osReadFile = os.ReadFile
			osWriteFile = os.WriteFile
		})

		osStat = func(name string) (os.FileInfo, error) { return nil, nil }
		osReadFile = func(name string) ([]byte, error) {
			return []byte("# CHANGELOG\n\n## [v1.0.0] - 2025-01-01\n\nOld entry."), nil
		}
		var writtenContent string
		osWriteFile = func(name string, data []byte, perm os.FileMode) error {
			writtenContent = string(data)
			return nil
		}

		// Act
		err := UpdateChangelog("1.1.0", "2025-01-02", "New entry")

		// Assert
		assert.NoError(t, err)
		assert.True(t, strings.HasPrefix(writtenContent, "# CHANGELOG\n\n## [v1.1.0] - 2025-01-02\n\nNew entry"))
		assert.Contains(t, writtenContent, "## [v1.0.0] - 2025-01-01")
	})
}

func TestApplyChanges(t *testing.T) {
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

	t.Run("successful apply", func(t *testing.T) {
		// Arrange
		t.Cleanup(func() {
			filepathWalk = filepath.Walk
			osReadFile = os.ReadFile
			utilsUpdateMetadata = utils.UpdateMetadata
			osRemove = os.Remove
			osStat = os.Stat
			osWriteFile = os.WriteFile
		})

		filepathWalk = func(root string, walkFn filepath.WalkFunc) error {
			return walkFn(".changes/change1.mdx", &mockFileInfo{}, nil)
		}
		osReadFile = func(name string) ([]byte, error) {
			if strings.HasSuffix(name, ".mdx") {
				return []byte("---\nid: test\nbump: patch\nnext: 1.0.1\ndate: 2025-01-01\n---\n- A change."), nil
			}
			return nil, nil
		}
		utilsUpdateMetadata = func(key string, value interface{}, sectionOpt ...string) error { return nil }
		osRemove = func(name string) error { return nil }
		osStat = func(name string) (os.FileInfo, error) { return nil, os.ErrNotExist } // for new changelog
		osWriteFile = func(name string, data []byte, perm os.FileMode) error { return nil }
		buf, restore := captureOutput()

		changesetApplyCmd.Run(changesetApplyCmd, []string{})

		// Assert
		restore()
		output := buf.String()
		assert.Contains(t, output, "All changesets applied successfully!")
	})

	t.Run("no changeset files", func(t *testing.T) {
		// Arrange
		t.Cleanup(func() { filepathWalk = filepath.Walk })
		filepathWalk = func(root string, walkFn filepath.WalkFunc) error { return nil } // no files found
		buf, restore := captureOutput()

		// Act
		ApplyChanges()

		// Assert
		restore()
		output := buf.String()
		assert.Contains(t, output, "No changeset files found.")
		assert.Contains(t, output, "run `prasmoid changeset add` to create a changeset.")
	})

	t.Run("changeset dir not found", func(t *testing.T) {
		// Arrange
		t.Cleanup(func() { filepathWalk = filepath.Walk })
		filepathWalk = func(root string, walkFn filepath.WalkFunc) error { return errors.New("dir not found") }
		buf, restore := captureOutput()

		// Act
		ApplyChanges()

		// Assert
		restore()
		output := buf.String()
		assert.Contains(t, output, "Failed to walk changes directory")
	})

	t.Run("changeset file read error", func(t *testing.T) {
		// Arrange
		t.Cleanup(func() { filepathWalk = filepath.Walk })
		filepathWalk = func(root string, walkFn filepath.WalkFunc) error {
			return walkFn(".changes/change1.mdx", &mockFileInfo{}, nil)
		}
		osReadFile = func(name string) ([]byte, error) { return nil, errors.New("file not found") }
		buf, restore := captureOutput()

		// Act
		ApplyChanges()

		// Assert
		restore()
		output := buf.String()
		assert.Contains(t, output, "Failed to read changeset file")
	})
}

func TestMatterParse(t *testing.T) {
	t.Run("valid frontmatter", func(t *testing.T) {
		data := []byte(`---
id: test
bump: patch
next: 1.0.1
date: 2025-01-01
---
This is the body.`)
		meta, body, err := matterParse(data)
		assert.NoError(t, err)
		assert.Equal(t, "test", meta.ID)
		assert.Equal(t, "patch", meta.Bump)
		assert.Equal(t, "1.0.1", meta.Next)
		assert.Equal(t, "2025-01-01", meta.Date)
		assert.Equal(t, "This is the body.", strings.TrimSpace(body))
	})

	t.Run("invalid yaml", func(t *testing.T) {
		data := []byte(`---
id: test
invalid-yaml: :
---
This is the body.`)
		_, _, err := matterParse(data)
		assert.Error(t, err)
	})

}
