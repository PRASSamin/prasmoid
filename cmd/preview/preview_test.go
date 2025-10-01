package preview

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

	"github.com/AlecAivazis/survey/v2"
	"github.com/PRASSamin/prasmoid/tests"
	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"github.com/stretchr/testify/assert"
)

var previewTestMutex sync.Mutex

func TestPreviewCmdRun(t *testing.T) {
	previewTestMutex.Lock()
	defer previewTestMutex.Unlock()

	t.Run("plasmoidviewer not installed", func(t *testing.T) {
		// Arrange
		_, _, cleanup := tests.SetupTestEnvironment(t)
		defer cleanup()

		originalIsPackageInstalled := utilsIsPackageInstalled
		utilsIsPackageInstalled = func(pkg string) bool { return false }
		defer func() { utilsIsPackageInstalled = originalIsPackageInstalled }()

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		color.Output = w

		// Act
		PreviewCmd.Run(PreviewCmd, []string{})
		_ = w.Close()

		// Assert
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		assert.Contains(t, buf.String(), "preview command is disabled due to missing dependencies.")
	})


	originalPreviewPlasmoid := previewPlasmoid
	defer func() {
		previewPlasmoid = originalPreviewPlasmoid
	}()

	// setup happy path mocks
	setupMocks := func() {
		utilsIsValidPlasmoid = func() bool { return true }
		utilsIsLinked = func() bool { return true }
		previewPlasmoid = func(watch bool) error { return nil }
		utilsIsPackageInstalled = func(pkg string) bool { return true }
	}

	t.Run("invalid plasmoid", func(t *testing.T) {
		// Arrange
		_, _, cleanup := tests.SetupTestEnvironment(t)
		defer cleanup()
		utilsIsValidPlasmoid = func() bool { return false }
		defer func() { utilsIsValidPlasmoid = func() bool { return true } }()

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		color.Output = w

		// Act
		PreviewCmd.Run(PreviewCmd, []string{})
		_ = w.Close()

		// Assert
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		assert.Contains(t, buf.String(), "Current directory is not a valid plasmoid.")
	})

	t.Run("failed to link", func(t *testing.T) {
		// Arrange
		_, _, cleanup := tests.SetupTestEnvironment(t)
		defer cleanup()
		setupMocks()
		orgUtilsIsLinked := utilsIsLinked
		orgSurveyAskOne := surveyAskOne
		orgLinkLinkPlasmoid := linkLinkPlasmoid

		utilsIsLinked = func() bool { return false }
		surveyAskOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error { return nil }
		linkLinkPlasmoid = func(dest string) error { return errors.New("link error") }
		defer func() { utilsIsLinked = orgUtilsIsLinked }()
		defer func() { surveyAskOne = orgSurveyAskOne }()
		defer func() { linkLinkPlasmoid = orgLinkLinkPlasmoid }()

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		color.Output = w

		confirmLink = true
		PreviewCmd.Run(PreviewCmd, []string{})
		_ = w.Close()

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		color.Output = os.Stdout
		output := buf.String()
		assert.Contains(t, output, "Failed to link plasmoid:")
	})

	t.Run("not linked, user confirms link, success", func(t *testing.T) {
		// Arrange
		_, _, cleanup := tests.SetupTestEnvironment(t)
		defer cleanup()
		setupMocks()
		utilsIsLinked = func() bool { return false }
		surveyAskOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
			*(response.(*bool)) = true
			return nil
		}
		var linkCalled bool
		linkLinkPlasmoid = func(dest string) error {
			linkCalled = true
			return nil
		}
		previewPlasmoid = func(watch bool) error { return nil }

		// Act
		PreviewCmd.Run(PreviewCmd, []string{})

		// Assert
		assert.True(t, linkCalled, "LinkPlasmoid should have been called")
	})

	t.Run("not linked, user cancels", func(t *testing.T) {
		// Arrange
		_, _, cleanup := tests.SetupTestEnvironment(t)
		defer cleanup()
		setupMocks()
		utilsIsLinked = func() bool { return false }
		surveyAskOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
			*(response.(*bool)) = false
			return nil
		}
		var linkCalled bool
		linkLinkPlasmoid = func(dest string) error {
			linkCalled = true
			return nil
		}
		previewPlasmoid = func(watch bool) error { return nil }

		// Act
		PreviewCmd.Run(PreviewCmd, []string{})

		// Assert
		assert.False(t, linkCalled, "LinkPlasmoid should not have been called")
	})

	t.Run("previewPlasmoid fails", func(t *testing.T) {
		// Arrange
		_, _, cleanup := tests.SetupTestEnvironment(t)
		defer cleanup()
		setupMocks()
		previewPlasmoid = func(watch bool) error { return errors.New("preview error") }

		// Capture output
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		color.Output = w

		// Act
		PreviewCmd.Run(PreviewCmd, []string{})
		_ = w.Close()

		// Assert
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		os.Stdout = oldStdout
		assert.Contains(t, buf.String(), "Failed to preview plasmoid:")
	})
}

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

func TestWatchOnChange(t *testing.T) {
	previewTestMutex.Lock()
	defer previewTestMutex.Unlock()

	// Backup and restore
	originalFsnotifyNewWatcher := fsnotifyNewWatcher
	originalFilepathWalk := filepathWalk
	originalExecCommand := execCommand
	originalTimeAfterFunc := timeAfterFunc
	originalUtilsIsQmlFile := utilsIsQmlFile
	originalSignalNotify := signalNotify
	defer func() {
		fsnotifyNewWatcher = originalFsnotifyNewWatcher
		filepathWalk = originalFilepathWalk
		execCommand = originalExecCommand
		timeAfterFunc = originalTimeAfterFunc
		utilsIsQmlFile = originalUtilsIsQmlFile
		signalNotify = originalSignalNotify
	}()

	// Redirect output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w
	color.Output = w

	var buf bytes.Buffer
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, _ = io.Copy(&buf, r)
	}()

	t.Run("success: restarts on change", func(t *testing.T) {
		// Arrange
		mw := newMockWatcher()
		fsnotifyNewWatcher = func() (iWatcher, error) {
			return mw, nil
		}
		filepathWalk = func(root string, walkFn filepath.WalkFunc) error {
			return walkFn(root, &MockFileInfo{isDir: true}, nil)
		}
		utilsIsQmlFile = func(file string) bool {
			return strings.HasSuffix(file, ".qml")
		}

		var execCount int
		execCommand = func(name string, arg ...string) *exec.Cmd {
			execCount++
			cmd := exec.Command("sleep", "0.1")
			return cmd
		}

		timeAfterFunc = func(d time.Duration, f func()) *time.Timer {
			f() // immediate execution
			return time.NewTimer(d)
		}

		quitChan := make(chan os.Signal, 1)
		signalNotify = func(c chan<- os.Signal, sig ...os.Signal) {
			// Redirect signals to our channel
			go func() {
				for s := range quitChan {
					c <- s
				}
			}()
		}

		// Act
		done := make(chan bool)
		go func() {
			watchOnChange("/fake/path", "my-plasmoid")
			close(done)
		}()

		time.Sleep(100 * time.Millisecond) // allow initial start
		mw.EventChan <- fsnotify.Event{Name: "test.qml", Op: fsnotify.Write}
		time.Sleep(100 * time.Millisecond) // allow restart

		quitChan <- os.Interrupt
		<-done

		// Assert
		assert.Equal(t, 2, execCount, "execCommand should be called twice")
	})

	t.Run("error: newWatcher fails", func(t *testing.T) {
		fsnotifyNewWatcher = func() (iWatcher, error) {
			return nil, errors.New("watcher failed")
		}
		watchOnChange("/fake/path", "id")
	})

	t.Run("error: filepathWalk fails", func(t *testing.T) {
		mw := newMockWatcher()
		fsnotifyNewWatcher = func() (iWatcher, error) {
			return mw, nil
		}
		filepathWalk = func(root string, walkFn filepath.WalkFunc) error {
			return errors.New("walk failed")
		}
		execCommand = func(name string, arg ...string) *exec.Cmd {
			return exec.Command("true")
		}

		quitChan := make(chan os.Signal, 1)
		signalNotify = func(c chan<- os.Signal, sig ...os.Signal) {
			go func() {
				for s := range quitChan {
					c <- s
				}
			}()
		}
		done := make(chan bool)
		go func() {
			watchOnChange("/fake/path", "id")
			close(done)
		}()
		quitChan <- os.Interrupt
		<-done
	})

	t.Run("error: watcher.Errors() receives an error", func(t *testing.T) {
		// Arrange
		mw := newMockWatcher()
		fsnotifyNewWatcher = func() (iWatcher, error) {
			return mw, nil
		}
		filepathWalk = func(root string, walkFn filepath.WalkFunc) error {
			return nil
		}
		execCommand = func(name string, arg ...string) *exec.Cmd {
			return exec.Command("true")
		}

		quitChan := make(chan os.Signal, 1)
		signalNotify = func(c chan<- os.Signal, sig ...os.Signal) {
			go func() {
				for s := range quitChan {
					c <- s
				}
			}()
		}

		// Act
		done := make(chan bool)
		go func() {
			watchOnChange("/fake/path", "id")
			close(done)
		}()

		expectedError := errors.New("simulated watcher error")
		mw.ErrorChan <- expectedError

		time.Sleep(50 * time.Millisecond)

		quitChan <- os.Interrupt
		<-done
	})

	// Cleanup
	_ = w.Close()
	wg.Wait()
	os.Stdout = oldStdout

	output := buf.String()
	assert.Contains(t, output, "Previewer running in watch mode")
	assert.Contains(t, output, "Failed to start watcher: watcher failed")
	assert.Contains(t, output, "Failed to watch directory: walk failed")
	assert.Contains(t, output, "Watcher error: simulated watcher error")
}

func TestPreviewPlasmoid(t *testing.T) {
	previewTestMutex.Lock()
	defer previewTestMutex.Unlock()

	// Backup and restore
	originalUtilsGetDataFromMetadata := utilsGetDataFromMetadata
	originalExecCommand := execCommand
	originalWatchOnChange := watchOnChange
	defer func() {
		utilsGetDataFromMetadata = originalUtilsGetDataFromMetadata
		execCommand = originalExecCommand
		watchOnChange = originalWatchOnChange
	}()

	t.Run("success: runs plasmoidviewer without watch", func(t *testing.T) {
		// Arrange
		utilsGetDataFromMetadata = func(key string) (interface{}, error) {
			return "my-plasmoid", nil
		}
		var execCalled bool
		execCommand = func(name string, arg ...string) *exec.Cmd {
			execCalled = true
			assert.Equal(t, "plasmoidviewer", name)
			assert.Equal(t, []string{"-a", "my-plasmoid"}, arg)
			cmd := exec.Command("true") // command that does nothing and succeeds
			return cmd
		}

		// Act
		err := previewPlasmoid(false)

		// Assert
		assert.NoError(t, err)
		assert.True(t, execCalled)
	})

	t.Run("success: calls watchOnChange with watch flag", func(t *testing.T) {
		// Arrange
		utilsGetDataFromMetadata = func(key string) (interface{}, error) {
			return "my-plasmoid", nil
		}
		var watchCalled bool
		watchOnChange = func(path string, id string) {
			watchCalled = true
			assert.Equal(t, "./contents", path)
			assert.Equal(t, "my-plasmoid", id)
		}

		// Act
		err := previewPlasmoid(true)

		// Assert
		assert.NoError(t, err)
		assert.True(t, watchCalled)
	})

	t.Run("error: GetDataFromMetadata fails", func(t *testing.T) {
		// Arrange
		expectedErr := errors.New("metadata error")
		utilsGetDataFromMetadata = func(key string) (interface{}, error) {
			return nil, expectedErr
		}

		// Act
		err := previewPlasmoid(false)

		// Assert
		assert.Equal(t, expectedErr, err)
	})

	t.Run("error: exec.Command fails", func(t *testing.T) {
		// Arrange
		utilsGetDataFromMetadata = func(key string) (interface{}, error) {
			return "my-plasmoid", nil
		}
		execCommand = func(name string, arg ...string) *exec.Cmd {
			return exec.Command("non-existent-command")
		}

		// Act
		err := previewPlasmoid(false)

		// Assert
		assert.Error(t, err)
	})
}
