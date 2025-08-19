package preview

import (
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/PRASSamin/prasmoid/tests"
	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func TestPreviewCmdRun(t *testing.T) {
	// setup happy path mocks
	setupMocks := func() {
		utilsIsValidPlasmoid = func() bool { return true }
		utilsIsLinked = func() bool { return true }
		utilsIsPackageInstalled = func(pkg string) bool { return true }
		previewPlasmoid = func(watch bool) error { return nil }
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

	t.Run("failed plasmaviewer install", func(t *testing.T) {
		// Arrange
		_, _, cleanup := tests.SetupTestEnvironment(t)
		defer cleanup()
		setupMocks()
		orgUtilsIsPackageInstalled := utilsIsPackageInstalled
		orgUtilsIsLinked := utilsIsLinked
		orgUtilsInstallPackage := utilsInstallPackage
		orgSurveyAskOne := surveyAskOne
		
		utilsIsPackageInstalled = func(name string) bool {return false}
		utilsIsLinked = func() bool {return true}
		utilsInstallPackage = func(pm, pkg string, names map[string]string) error { return errors.New("install error") }
		surveyAskOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error { return nil }
		defer func() { utilsIsPackageInstalled = orgUtilsIsPackageInstalled }()
		defer func() { utilsIsLinked = orgUtilsIsLinked }()
		defer func() { utilsInstallPackage = orgUtilsInstallPackage }()
		defer func() { surveyAskOne = orgSurveyAskOne }()

		orgStdout := os.Stdout
		defer func() { os.Stdout = orgStdout }()
		r, w, _ := os.Pipe()
		os.Stdout = w
		color.Output = w

		confirmInstallation = true
		PreviewCmd.Run(PreviewCmd, []string{})
		_ = w.Close()

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		assert.Contains(t, buf.String(), "Failed to install plasmoidviewer:")
	})

	t.Run("failed to link", func(t *testing.T) {
		// Arrange
		_, _, cleanup := tests.SetupTestEnvironment(t)
		defer cleanup()
		setupMocks()
		orgUtilsIsLinked := utilsIsLinked
		orgSurveyAskOne := surveyAskOne
		orgLinkLinkPlasmoid := linkLinkPlasmoid
		
		utilsIsLinked = func() bool {return false}
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

	t.Run("viewer not installed, user confirms install", func(t *testing.T) {
		// Arrange
		_, _, cleanup := tests.SetupTestEnvironment(t)
		defer cleanup()
		setupMocks()
		utilsIsPackageInstalled = func(pkg string) bool { return false }
		surveyAskOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
			*(response.(*bool)) = true
			return nil
		}
		var installCalled bool
		utilsInstallPackage = func(pm, pkg string, names map[string]string) error {
			installCalled = true
			return nil
		}
		previewPlasmoid = func(watch bool) error { return nil }

		// Act
		PreviewCmd.Run(PreviewCmd, []string{})

		// Assert
		assert.True(t, installCalled, "InstallPackage should have been called")
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

func TestWatchOnChange(t *testing.T) {
	// --- Mocks & Originals ---
	originalExecCommand := execCommand
	originalTimeAfterFunc := timeAfterFunc
	originalSignalNotify := signalNotify
	originalUtilsIsQmlFile := utilsIsQmlFile

	var commands []*exec.Cmd
	var mu sync.Mutex

	t.Cleanup(func() {
		execCommand = originalExecCommand
		timeAfterFunc = originalTimeAfterFunc
		signalNotify = originalSignalNotify
		utilsIsQmlFile = originalUtilsIsQmlFile
	})

	// --- Mock Setup ---
	execCommand = func(name string, arg ...string) *exec.Cmd {
		mu.Lock()
		defer mu.Unlock()
		// Use a command that can be started and killed.
		cmd := exec.Command("sleep", "1")
		commands = append(commands, cmd)
		return cmd
	}

	timerChan := make(chan func(), 1)
	timeAfterFunc = func(d time.Duration, f func()) *time.Timer {
		timerChan <- f
		// Return a dummy timer that will not fire during the test.
		return time.NewTimer(1 * time.Hour)
	}

	signalChan := make(chan os.Signal, 1)
	signalNotify = func(c chan<- os.Signal, sig ...os.Signal) {
		// Forward signal from our test channel to the channel used by the function.
		go func() {
			s := <-signalChan
			c <- s
		}()
	}

	utilsIsQmlFile = func(path string) bool { return true }

	// --- Test Setup ---
	tempDir := t.TempDir()
	// The watcher needs a directory to watch.
	// The code walks the directory and adds subdirectories.
	// Let's create a subdirectory to make sure that works.
	subDir := filepath.Join(tempDir, "ui")
	err := os.Mkdir(subDir, 0755)
	assert.NoError(t, err)

	qmlFile := filepath.Join(subDir, "test.qml")
	err = os.WriteFile(qmlFile, []byte("initial content"), 0644)
	assert.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		watchOnChange(tempDir, "my-id")
	}()

	// Allow some time for the initial command to start and watcher to be set up
	time.Sleep(200 * time.Millisecond)

	// --- Assertions ---
	mu.Lock()
	assert.Equal(t, 1, len(commands), "initial command should have started")
	initialCmd := commands[0]
	mu.Unlock()
	assert.NotNil(t, initialCmd.Process, "initial process should be running")

	// Simulate file modification
	err = os.WriteFile(qmlFile, []byte("new content"), 0644)
	assert.NoError(t, err)

	// Get the debounced function
	var debouncedFunc func()
	select {
	case debouncedFunc = <-timerChan:
		// great
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for debouncer")
	}

	// Execute the debounced function
	debouncedFunc()

	// Allow some time for the new command to start
	time.Sleep(200 * time.Millisecond)

	mu.Lock()
	assert.Equal(t, 2, len(commands), "a new command should have started")
	newCmd := commands[1]
	mu.Unlock()
	assert.NotNil(t, newCmd.Process, "new process should be running")

	// Check if the initial process was killed.
	// Sending signal 0 to a process checks if it exists. An error means it doesn't.
	err = initialCmd.Process.Signal(syscall.Signal(0))
	assert.Error(t, err, "initial process should have been killed")

	// --- Cleanup ---
	// Terminate the watcher loop
	signalChan <- os.Interrupt
	wg.Wait()

	// Kill the last running process to be clean
	if newCmd.Process != nil {
		_ = newCmd.Process.Kill()
		_ = newCmd.Wait()
	}
}
