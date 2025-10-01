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
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/PRASSamin/prasmoid/tests"
	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockFileInfo is a mock implementation of os.FileInfo
type MockFileInfo struct {
	isDir bool
}

func (m *MockFileInfo) Name() string       { return "mock" }
func (m *MockFileInfo) Size() int64        { return 0 }
func (m *MockFileInfo) Mode() os.FileMode  { return 0 }
func (m *MockFileInfo) ModTime() time.Time { return time.Now() }
func (m *MockFileInfo) IsDir() bool        { return m.isDir }
func (m *MockFileInfo) Sys() interface{}   { return nil }

// mockWatcher is a mock implementation of the iWatcher interface for testing.
type mockWatcher struct {
	EventChan   chan fsnotify.Event
	ErrorChan   chan error
	addPaths    []string
	closeCalled bool
	mu          sync.Mutex
}

func newMockWatcher() *mockWatcher {
	return &mockWatcher{
		EventChan: make(chan fsnotify.Event),
		ErrorChan: make(chan error),
	}
}

func (m *mockWatcher) Add(path string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.addPaths = append(m.addPaths, path)
	return nil
}

func (m *mockWatcher) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closeCalled = true
	close(m.EventChan)
	close(m.ErrorChan)
	return nil
}

func (m *mockWatcher) Events() chan fsnotify.Event {
	return m.EventChan
}

func (m *mockWatcher) Errors() chan error {
	return m.ErrorChan
}

func TestPrettifyOnWatch(t *testing.T) {
	// Backup and restore original functions
	originalNewWatcher := newWatcher
	originalFilepathWalk := filepathWalk
	originalExecCommand := execCommand
	originalTimeAfterFunc := timeAfterFunc
	defer func() {
		newWatcher = originalNewWatcher
		filepathWalk = originalFilepathWalk
		execCommand = originalExecCommand
		timeAfterFunc = originalTimeAfterFunc
	}()

	// Redirect stdout and stderr to a buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w    // Redirect stderr as well, as color.RedString uses it
	color.Output = w // Redirect color output as well

	var buf bytes.Buffer
	var wg sync.WaitGroup // Add WaitGroup
	wg.Add(1)             // Increment counter for the goroutine

	// Start a goroutine to copy output from the pipe to the buffer
	go func() {
		defer wg.Done()
		_, _ = io.Copy(&buf, r)
	}()

	t.Run("success: formats a modified QML file", func(t *testing.T) {
		// Arrange
		mw := newMockWatcher()
		newWatcher = func() (iWatcher, error) {
			return mw, nil
		}
		filepathWalk = func(root string, walkFn filepath.WalkFunc) error {
			// Simulate walking and adding a directory
			return walkFn(root, &MockFileInfo{isDir: true}, nil)
		}

		// Mock utils.IsQmlFile
		originalUtilsIsQmlFile := utilsIsQmlFile
		defer func() { utilsIsQmlFile = originalUtilsIsQmlFile }()
		utilsIsQmlFile = func(file string) bool {
			return strings.HasSuffix(file, ".qml")
		}

		// Use WaitGroup for synchronization
		var wgExec sync.WaitGroup // Renamed to avoid conflict with outer wg
		wgExec.Add(1)             // Expect one call to execCommand

		execCommand = func(name string, arg ...string) *exec.Cmd {
			assert.Equal(t, "qmlformat", name)
			assert.Equal(t, "-i", arg[0])
			assert.Equal(t, "test.qml", arg[1])
			t.Logf("execCommand mock called for file: %s", arg[1])
			wgExec.Done() // Signal that execCommand was called
			return exec.Command("true")
		}

		// Make debounce immediate
		timeAfterFunc = func(d time.Duration, f func()) *time.Timer {
			f()
			return time.NewTimer(d)
		}

		done := make(chan bool)
		go prettifyOnWatch("/fake/path", done)
		time.Sleep(200 * time.Millisecond) // Give the goroutine time to start

		t.Log("Sending event to mock watcher")
		mw.EventChan <- fsnotify.Event{Name: "test.qml", Op: fsnotify.Write}

		// Assert using WaitGroup with timeout
		c := make(chan struct{})
		go func() {
			defer close(c)
			wgExec.Wait() // Wait for execCommand to be called
		}()

		select {
		case <-c:
			t.Log("Received signal from WaitGroup")
			// Success, execCommand was called
		case <-time.After(500 * time.Millisecond):
			t.Fatal("execCommand was not called within timeout")
		}

		// Cleanup
		close(done)
	})

	t.Run("error: newWatcher fails", func(t *testing.T) {
		// Arrange
		expectedError := "watcher failed"
		newWatcher = func() (iWatcher, error) {
			return nil, errors.New(expectedError)
		}

		// Act
		prettifyOnWatch("/fake/path", make(chan bool))
	})

	t.Run("error: filepathWalk fails", func(t *testing.T) {
		// Arrange
		mw := newMockWatcher()
		newWatcher = func() (iWatcher, error) {
			return mw, nil
		}
		expectedError := "walk failed"
		filepathWalk = func(root string, walkFn filepath.WalkFunc) error {
			return errors.New(expectedError)
		}

		// Act
		done := make(chan bool)
		prettifyOnWatch("/fake/path", done)

		// Cleanup
		close(done)
	})

	t.Run("error: newWatcher fails", func(t *testing.T) {
		// Arrange
		expectedError := "watcher failed"
		newWatcher = func() (iWatcher, error) {
			return nil, errors.New(expectedError)
		}

		// Act
		prettifyOnWatch("/fake/path", make(chan bool))

		// Assert
		// Output is captured by the main test's buffer
		// No need to capture here
	})

	t.Run("error: watcher.Errors() receives an error", func(t *testing.T) {
		// Arrange
		mw := newMockWatcher()
		newWatcher = func() (iWatcher, error) {
			return mw, nil
		}
		// Mock filepathWalk to avoid actual file system interaction
		filepathWalk = func(root string, walkFn filepath.WalkFunc) error {
			return nil // Don't add any paths, just let it proceed
		}

		// Act
		done := make(chan bool)
		go prettifyOnWatch("/fake/path", done)

		// Send an error to the watcher's error channel
		expectedError := errors.New("simulated watcher error")
		mw.ErrorChan <- expectedError

		// Give the goroutine time to process the error
		time.Sleep(50 * time.Millisecond)

		// Cleanup: Close the done channel to ensure the main prettifyOnWatch function exits
		close(done)
	})

	// After all subtests, close the pipe writer to signal EOF to the reader goroutine
	_ = w.Close()
	wg.Wait()
	os.Stdout = oldStdout

	output := buf.String()
	assert.Contains(t, output, "Formatter running in watch mode ...")
	assert.Contains(t, output, "Formatted: test.qml")
	assert.Contains(t, output, "Failed to start watcher: watcher failed")
	assert.Contains(t, output, "Failed to watch directory: walk failed")
	assert.Contains(t, output, "Watcher error: simulated watcher error")
}

func TestPrettify(t *testing.T) {
	_, cleanup := tests.SetupTestProject(t)
	defer cleanup()

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	prettify("contents")

	require.NoError(t, w.Close())
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	os.Stdout = oldStdout
	color.Output = os.Stdout
	output := buf.String()
	assert.Contains(t, output, "Formatted 1 files")
}

func TestFormatCmdRun(t *testing.T) {
	t.Run("qmlformat not installed", func(t *testing.T) {
		// Arrange
		originalIsPackageInstalled := utilsIsPackageInstalled
		utilsIsPackageInstalled = func(pkg string) bool { return false }
		defer func() { utilsIsPackageInstalled = originalIsPackageInstalled }()

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		color.Output = w

		// Act
		FormatCmd.Run(FormatCmd, []string{})
		_ = w.Close()

		// Assert
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		output := buf.String()
		assert.Contains(t, output, "format command is disabled due to missing qmlformat dependency.")
	})

	t.Run("Not a valid plasmoid", func(t *testing.T) {
		// Arrange
		originalIsPackageInstalled := utilsIsPackageInstalled
		utilsIsPackageInstalled = func(pkg string) bool { return true } // Mock as installed
		defer func() { utilsIsPackageInstalled = originalIsPackageInstalled }()

		tmpDir, err := os.MkdirTemp("", "format-invalid-*")
		require.NoError(t, err)
		defer func() { require.NoError(t, os.RemoveAll(tmpDir)) }()

		oldWd, _ := os.Getwd()
		require.NoError(t, os.Chdir(tmpDir))
		defer func() { require.NoError(t, os.Chdir(oldWd)) }()

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		color.Output = w

		// Act
		FormatCmd.Run(FormatCmd, []string{})
		_ = w.Close()

		// Assert
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		color.Output = os.Stdout
		assert.Contains(t, buf.String(), "Current directory is not a valid plasmoid")
	})

	t.Run("Run format successfully", func(t *testing.T) {
		// Arrange
		originalIsPackageInstalled := utilsIsPackageInstalled
		utilsIsPackageInstalled = func(pkg string) bool { return true } // Mock as installed
		defer func() { utilsIsPackageInstalled = originalIsPackageInstalled }()

		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		color.Output = w

		// Act
		FormatCmd.Run(FormatCmd, []string{})
		_ = w.Close()

		// Assert
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		color.Output = os.Stdout
		output := buf.String()
		assert.Contains(t, output, "Formatted 1 files")
	})
}

func TestPrettifyError(t *testing.T) {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	prettify("/non-existent-path")

	require.NoError(t, w.Close())
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
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
	_, _ = io.Copy(&buf, r)
	os.Stdout = oldStdout
	color.Output = os.Stdout
	output := buf.String()
	assert.Contains(t, output, "Failed to format qml files")
}
