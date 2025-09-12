package cmd

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/PRASSamin/prasmoid/internal"
	"github.com/PRASSamin/prasmoid/types"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRootCmd_Run(t *testing.T) {
	// Capture stdout
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

	// Mock os.Exit to prevent actual exit during test
	oldOsExit := osExit
	defer func() { osExit = oldOsExit }()
	var exitCode int
	osExit = func(code int) { exitCode = code }

	t.Run("version flag set", func(t *testing.T) {
		// Arrange
		cmd := &cobra.Command{}
		cmd.Flags().BoolP("version", "v", false, "show Prasmoid version")
		_ = cmd.Flags().Set("version", "true")
		buf, restore := captureOutput()

		// Act
		RootCmd.Run(cmd, []string{})

		// Assert
		restore()
		assert.Contains(t, buf.String(), internal.AppMetaData.Version)
		assert.Equal(t, 0, exitCode)
	})
}

func TestExecute(t *testing.T) {
	// Save original functions
	originalDiscover := extendcliDiscoverAndRegisterCustomCommands
	originalCheckForUpdates := CheckForUpdates
	originalRootCmdExecute := rootCmdExecute
	originalOsExit := osExit

	t.Cleanup(func() {
		extendcliDiscoverAndRegisterCustomCommands = originalDiscover
		CheckForUpdates = originalCheckForUpdates
		rootCmdExecute = originalRootCmdExecute
		osExit = originalOsExit
	})

	// Mock os.Exit to prevent actual exit during test
	var exitCode int
	osExit = func(code int) { exitCode = code }

	t.Run("successful execution", func(t *testing.T) {
		// Arrange
		var discoverCalled, checkForUpdatesCalled, rootCmdExecuteCalled bool
		extendcliDiscoverAndRegisterCustomCommands = func(*cobra.Command, types.Config) { discoverCalled = true }
		CheckForUpdates = func() { checkForUpdatesCalled = true }
		rootCmdExecute = func() error { rootCmdExecuteCalled = true; return nil }

		// Act
		Execute()

		// Assert
		assert.True(t, discoverCalled)
		assert.True(t, checkForUpdatesCalled)
		assert.True(t, rootCmdExecuteCalled)
		assert.Equal(t, 0, exitCode)
	})

	t.Run("rootCmdExecute returns error", func(t *testing.T) {
		// Arrange
		extendcliDiscoverAndRegisterCustomCommands = func(*cobra.Command, types.Config) {}
		CheckForUpdates = func() {}
		rootCmdExecute = func() error { return errors.New("execute error") }

		// Act
		Execute()

		// Assert
		assert.Equal(t, 1, exitCode)
	})
}

func TestCheckForUpdates_AllBranches(t *testing.T) {
	// Backup originals
	origReadUpdateCache := readUpdateCache
	origTimeParse := timeParse
	origTimeSince := timeSince
	origTlsDial := tlsDial
	origConnWrite := connWrite
	origConnClose := connClose
	origIoReadAll := ioReadAll
	origGetLatestTag := getLatestTag
	origWriteUpdateCache := writeUpdateCache
	origIsUpdateAvailable := isUpdateAvailable
	origPrintUpdateMessage := printUpdateMessage
	origLogPrintf := logPrintf

	t.Cleanup(func() {
		readUpdateCache = origReadUpdateCache
		timeParse = origTimeParse
		timeSince = origTimeSince
		tlsDial = origTlsDial
		connWrite = origConnWrite
		connClose = origConnClose
		ioReadAll = origIoReadAll
		getLatestTag = origGetLatestTag
		writeUpdateCache = origWriteUpdateCache
		isUpdateAvailable = origIsUpdateAvailable
		printUpdateMessage = origPrintUpdateMessage
		logPrintf = origLogPrintf
	})

	logPrintf = func(format string, v ...interface{}) { /* silence */ }

	t.Run("cache expired due to parse error", func(t *testing.T) {
		readUpdateCache = func() (map[string]interface{}, error) {
			return map[string]interface{}{"last_checked": "bad-time"}, nil
		}
		timeParse = func(layout, value string) (time.Time, error) { return time.Time{}, errors.New("bad time") }
		tlsDial = func(_, _ string, _ *tls.Config) (*tls.Conn, error) { return nil, errors.New("dial fail") }

		CheckForUpdates() // should early return, no panic
	})

	t.Run("tlsDial works, but connWrite fails", func(t *testing.T) {
		readUpdateCache = func() (map[string]interface{}, error) { return nil, errors.New("miss") }
		mockConn := &tls.Conn{}
		tlsDial = func(_, _ string, _ *tls.Config) (*tls.Conn, error) { return mockConn, nil }
		connWrite = func(_ *tls.Conn, _ []byte) (int, error) { return 0, errors.New("write fail") }
		connClose = func(_ *tls.Conn) error { return nil }

		CheckForUpdates() // hit connWrite fail branch
	})

	t.Run("tlsDial ok, connWrite ok, ioReadAll fails", func(t *testing.T) {
		readUpdateCache = func() (map[string]interface{}, error) { return nil, errors.New("miss") }
		mockConn := &tls.Conn{}
		tlsDial = func(_, _ string, _ *tls.Config) (*tls.Conn, error) { return mockConn, nil }
		connWrite = func(_ *tls.Conn, _ []byte) (int, error) { return 1, nil }
		ioReadAll = func(_ io.Reader) ([]byte, error) { return nil, errors.New("read fail") }
		connClose = func(_ *tls.Conn) error { return errors.New("close fail") }

		var logged bool
		logPrintf = func(format string, v ...interface{}) { logged = true }

		CheckForUpdates()
		assert.True(t, logged, "logPrintf should be called when connClose fails")
	})

	t.Run("response malformed (<2 parts)", func(t *testing.T) {
		readUpdateCache = func() (map[string]interface{}, error) { return nil, errors.New("miss") }
		mockConn := &tls.Conn{}
		tlsDial = func(_, _ string, _ *tls.Config) (*tls.Conn, error) { return mockConn, nil }
		connWrite = func(_ *tls.Conn, _ []byte) (int, error) { return 1, nil }
		ioReadAll = func(_ io.Reader) ([]byte, error) { return []byte("no-body-here"), nil }
		connClose = func(_ *tls.Conn) error { return nil }

		CheckForUpdates() // hits malformed response branch
	})

	t.Run("response not 200 OK", func(t *testing.T) {
		readUpdateCache = func() (map[string]interface{}, error) { return nil, errors.New("miss") }
		mockConn := &tls.Conn{}
		tlsDial = func(_, _ string, _ *tls.Config) (*tls.Conn, error) { return mockConn, nil }
		connWrite = func(_ *tls.Conn, _ []byte) (int, error) { return 1, nil }
		ioReadAll = func(_ io.Reader) ([]byte, error) {
			return []byte("HTTP/1.1 404 Not Found\r\n\r\n{}"), nil
		}
		connClose = func(_ *tls.Conn) error { return nil }

		CheckForUpdates()
	})

	t.Run("cached, update available", func(t *testing.T) {
		readUpdateCache = func() (map[string]interface{}, error) {
			return map[string]interface{}{
					"last_checked": time.Now().Format(time.RFC3339),
					"latest_tag":   "2.0.0",
				},
				nil
		}
		timeParse = func(layout, value string) (time.Time, error) { return time.Now(), nil }
		timeSince = func(t time.Time) time.Duration { return 1 * time.Hour }
		isUpdateAvailable = func(tag string) bool { return true }
		var tlsDialCalled bool
		tlsDial = func(network, addr string, config *tls.Config) (*tls.Conn, error) {
			tlsDialCalled = true
			return nil, errors.New("mock error")
		}

		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		color.Output = w // Redirect color output as well
		buf := new(bytes.Buffer)

		// Act
		CheckForUpdates()

		// Assert
		_ = w.Close()
		_, _ = io.Copy(buf, r)
		os.Stdout = oldStdout
		color.Output = oldStdout
		assert.False(t, tlsDialCalled, "tlsDial should not be called if cache is valid and update available")
		require.Contains(t, buf.String(), "update available")
		require.Contains(t, buf.String(), "2.0.0")
	})

	t.Run("happy flow, update available", func(t *testing.T) {
		readUpdateCache = func() (map[string]interface{}, error) { return nil, errors.New("miss") }
		mockConn := &tls.Conn{}
		tlsDial = func(_, _ string, _ *tls.Config) (*tls.Conn, error) { return mockConn, nil }
		connWrite = func(_ *tls.Conn, _ []byte) (int, error) { return 1, nil }
		ioReadAll = func(_ io.Reader) ([]byte, error) {
			return []byte("HTTP/1.1 200 OK\r\n\r\n{\"tag_name\":\"3.0.0\"}"), nil
		}
		connClose = func(_ *tls.Conn) error { return nil }
		getLatestTag = func(b []byte) string { return "3.0.0" }
		writeUpdateCache = func(tag string, b []byte) {}
		isUpdateAvailable = func(tag string) bool { return true }
		var printed bool
		printUpdateMessage = func(tag string) { printed = true }

		CheckForUpdates()
		assert.True(t, printed)
	})

	t.Run("happy flow, update available", func(t *testing.T) {
		readUpdateCache = func() (map[string]interface{}, error) { return nil, errors.New("miss") }
		mockConn := &tls.Conn{}
		tlsDial = func(_, _ string, _ *tls.Config) (*tls.Conn, error) { return mockConn, nil }
		connWrite = func(_ *tls.Conn, _ []byte) (int, error) { return 1, nil }
		ioReadAll = func(_ io.Reader) ([]byte, error) {
			return []byte("HTTP/1.1 200 OK\r\n\r\n{\"tag_name\":\"3.0.0\"}"), nil
		}
		connClose = func(_ *tls.Conn) error { return nil }
		getLatestTag = func(b []byte) string { return "3.0.0" }
		writeUpdateCache = func(tag string, b []byte) {}
		isUpdateAvailable = func(tag string) bool { return true }
		var printed bool
		printUpdateMessage = func(tag string) { printed = true }

		CheckForUpdates()
		assert.True(t, printed)
	})

	t.Run("happy flow, no update available", func(t *testing.T) {
		readUpdateCache = func() (map[string]interface{}, error) { return nil, errors.New("miss") }
		mockConn := &tls.Conn{}
		tlsDial = func(_, _ string, _ *tls.Config) (*tls.Conn, error) { return mockConn, nil }
		connWrite = func(_ *tls.Conn, _ []byte) (int, error) { return 1, nil }
		ioReadAll = func(_ io.Reader) ([]byte, error) {
			return []byte("HTTP/1.1 200 OK\r\n\r\n{\"tag_name\":\"3.0.0\"}"), nil
		}
		connClose = func(_ *tls.Conn) error { return nil }
		getLatestTag = func(b []byte) string { return "3.0.0" }
		writeUpdateCache = func(tag string, b []byte) {}
		isUpdateAvailable = func(tag string) bool { return false }
		printUpdateMessage = func(tag string) { t.Fatal("should not print") }

		CheckForUpdates()
	})
}

func TestGetCacheFilePath(t *testing.T) {
	// Save original functions
	originalOsUserCacheDir := osUserCacheDir
	originalOsTempDir := osTempDir

	t.Cleanup(func() {
		osUserCacheDir = originalOsUserCacheDir
		osTempDir = originalOsTempDir
	})

	t.Run("UserCacheDir succeeds", func(t *testing.T) {
		// Arrange
		osUserCacheDir = func() (string, error) { return "/home/user/.cache", nil }

		// Act
		path := GetCacheFilePath()

		// Assert
		assert.Equal(t, filepath.Join("/home/user/.cache", "prasmoid_update.json"), path)
	})

	t.Run("UserCacheDir fails, TempDir succeeds", func(t *testing.T) {
		// Arrange
		osUserCacheDir = func() (string, error) { return "", errors.New("cache dir error") }
		osTempDir = func() string { return "/tmp" }

		// Act
		path := GetCacheFilePath()

		// Assert
		assert.Equal(t, filepath.Join("/tmp", "prasmoid_update.json"), path)
	})
}

func TestReadUpdateCache(t *testing.T) {
	// Save original functions
	originalGetCacheFilePath := GetCacheFilePath
	originalOsReadFile := osReadFile
	originalJsonUnmarshal := jsonUnmarshal

	t.Cleanup(func() {
		GetCacheFilePath = originalGetCacheFilePath
		osReadFile = originalOsReadFile
		jsonUnmarshal = originalJsonUnmarshal
	})

	t.Run("file read fails", func(t *testing.T) {
		// Arrange
		GetCacheFilePath = func() string { return "/nonexistent/path" }
		osReadFile = func(name string) ([]byte, error) { return nil, errors.New("read error") }

		// Act
		_, err := readUpdateCache()

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "read error")
	})

	t.Run("json unmarshal fails", func(t *testing.T) {
		// Arrange
		GetCacheFilePath = func() string { return "/mock/path" }
		osReadFile = func(name string) ([]byte, error) { return []byte("invalid json"), nil }
		jsonUnmarshal = func(data []byte, v interface{}) error { return errors.New("unmarshal error") }

		// Act
		_, err := readUpdateCache()

		// Assert
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		// Arrange
		GetCacheFilePath = func() string { return "/mock/path" }
		osReadFile = func(name string) ([]byte, error) { return []byte(`{"key":"value"}`), nil }
		jsonUnmarshal = func(data []byte, v interface{}) error {
			if m, ok := v.(*map[string]interface{}); ok {
				*m = map[string]interface{}{
					"key": "value",
				}
			}
			return nil
		}

		// Act
		cache, err := readUpdateCache()

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, cache)
		assert.Equal(t, "value", cache["key"])
	})
}

func TestWriteUpdateCache(t *testing.T) {
	// Save original functions
	originalGetCacheFilePath := GetCacheFilePath
	originalOsWriteFile := osWriteFile
	originalJsonMarshal := jsonMarshal
	originalJsonUnmarshal := jsonUnmarshal // Used by writeUpdateCache internally
	originalTimeNow := timeNow

	t.Cleanup(func() {
		GetCacheFilePath = originalGetCacheFilePath
		osWriteFile = originalOsWriteFile
		jsonMarshal = originalJsonMarshal
		jsonUnmarshal = originalJsonUnmarshal
		timeNow = originalTimeNow
	})

	t.Run("json unmarshal fails (releaseData)", func(t *testing.T) {
		// Arrange
		jsonUnmarshal = func(data []byte, v interface{}) error { return errors.New("unmarshal error") }
		var writeFileCalled bool
		osWriteFile = func(name string, data []byte, perm os.FileMode) error { writeFileCalled = true; return nil }

		// Act
		writeUpdateCache("v1.0.0", []byte("invalid json"))

		// Assert: Should not panic, but write file might still be called with empty data
		assert.True(t, writeFileCalled)
	})

	t.Run("json marshal fails (cache data)", func(t *testing.T) {
		// Arrange
		jsonMarshal = func(v interface{}) ([]byte, error) { return nil, errors.New("marshal error") }
		var writeFileCalled bool
		osWriteFile = func(name string, data []byte, perm os.FileMode) error { writeFileCalled = true; return nil }

		// Act
		writeUpdateCache("v1.0.0", []byte(`{"tag_name":"v1.0.0"}`))

		// Assert: Should not panic, but write file might still be called with empty data
		assert.True(t, writeFileCalled)
	})

	t.Run("osWriteFile fails", func(t *testing.T) {
		// Arrange
		osWriteFile = func(name string, data []byte, perm os.FileMode) error { return errors.New("write error") }

		// Act
		writeUpdateCache("v1.0.0", []byte(`{"tag_name":"v1.0.0"}`))

		// Assert: Should not panic
	})

	t.Run("success", func(t *testing.T) {
		// Arrange
		var writtenData []byte
		osWriteFile = func(name string, data []byte, perm os.FileMode) error {
			writtenData = data
			return nil
		}

		// Mock jsonUnmarshal to parse the input body
		jsonUnmarshal = func(data []byte, v interface{}) error {
			if m, ok := v.(*map[string]interface{}); ok {
				*m = map[string]interface{}{
					"tag_name": "v1.0.0",
					"other":    "data",
				}
			}
			return nil
		}

		// Mock jsonMarshal to return test data
		jsonMarshal = func(v interface{}) ([]byte, error) {
			return json.Marshal(map[string]interface{}{
				"last_checked": "2025-01-01T00:00:00Z",
				"latest_tag":   "v1.0.0",
				"data": map[string]interface{}{
					"tag_name": "v1.0.0",
					"other":    "data",
				},
			})
		}

		timeNow = func() time.Time { return time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC) }

		// Act
		writeUpdateCache("v1.0.0", []byte(`{"tag_name":"v1.0.0","other":"data"}`))

		// Assert
		assert.NotNil(t, writtenData)
		assert.Contains(t, string(writtenData), `"latest_tag":"v1.0.0"`)
		assert.Contains(t, string(writtenData), `"last_checked":"2025-01-01T00:00:00Z"`)
	})
}

func TestGetLatestTag(t *testing.T) {
	// Save original function
	originalJsonUnmarshal := jsonUnmarshal
	t.Cleanup(func() {
		jsonUnmarshal = originalJsonUnmarshal
	})

	t.Run("json unmarshal fails", func(t *testing.T) {
		// Arrange
		jsonUnmarshal = func(data []byte, v interface{}) error { return errors.New("unmarshal error") }

		// Act
		tag := getLatestTag([]byte("invalid json"))

		// Assert
		assert.Empty(t, tag)
	})

	t.Run("tag_name not found", func(t *testing.T) {
		// Arrange
		jsonUnmarshal = func(data []byte, v interface{}) error {
			if m, ok := v.(*map[string]interface{}); ok {
				*m = map[string]interface{}{
					"some_other_key": "value",
				}
			}
			return nil
		}

		// Act
		tag := getLatestTag([]byte(`{"some_other_key":"value"}`))

		// Assert
		assert.Empty(t, tag)
	})

	t.Run("tag_name is not string", func(t *testing.T) {
		// Arrange
		jsonUnmarshal = func(data []byte, v interface{}) error {
			if m, ok := v.(*map[string]interface{}); ok {
				*m = map[string]interface{}{
					"tag_name": 123,
				}
			}
			return nil
		}

		// Act
		tag := getLatestTag([]byte(`{"tag_name":123}`))

		// Assert
		assert.Empty(t, tag)
	})

	t.Run("success with v prefix", func(t *testing.T) {
		// Arrange
		jsonUnmarshal = func(data []byte, v interface{}) error {
			if m, ok := v.(*map[string]interface{}); ok {
				*m = map[string]interface{}{
					"tag_name": "v1.2.3",
				}
			}
			return nil
		}

		// Act
		tag := getLatestTag([]byte(`{"tag_name":"v1.2.3"}`))

		// Assert
		assert.Equal(t, "1.2.3", tag)
	})

	t.Run("success without v prefix", func(t *testing.T) {
		// Arrange
		jsonUnmarshal = func(data []byte, v interface{}) error {
			if m, ok := v.(*map[string]interface{}); ok {
				*m = map[string]interface{}{
					"tag_name": "1.2.3",
				}
			}
			return nil
		}

		// Act
		tag := getLatestTag([]byte(`{"tag_name":"1.2.3"}`))

		// Assert
		assert.Equal(t, "1.2.3", tag)
	})
}

func TestCompareVersions(t *testing.T) {
	t.Run("current less than latest", func(t *testing.T) {
		assert.Equal(t, -1, compareVersions("1.0.0", "1.0.1"))
		assert.Equal(t, -1, compareVersions("1.0.0", "1.1.0"))
		assert.Equal(t, -1, compareVersions("1.0.0", "2.0.0"))
		assert.Equal(t, -1, compareVersions("v1.0.0", "v1.0.1"))
	})

	t.Run("current equal to latest", func(t *testing.T) {
		assert.Equal(t, 0, compareVersions("1.0.0", "1.0.0"))
		assert.Equal(t, 0, compareVersions("v1.0.0", "1.0.0"))
		assert.Equal(t, 0, compareVersions("1.0.0", "v1.0.0"))
	})

	t.Run("current greater than latest", func(t *testing.T) {
		assert.Equal(t, 1, compareVersions("1.0.1", "1.0.0"))
		assert.Equal(t, 1, compareVersions("1.1.0", "1.0.0"))
		assert.Equal(t, 1, compareVersions("2.0.0", "1.0.0"))
		assert.Equal(t, 1, compareVersions("v2.0.0", "v1.0.0"))
	})

	t.Run("malformed versions", func(t *testing.T) {
		assert.Equal(t, -1, compareVersions("abc", "1.0.0")) // 0.0.0 < 1.0.0
		assert.Equal(t, 1, compareVersions("1.0.0", "abc"))  // 1.0.0 > 0.0.0
		assert.Equal(t, 0, compareVersions("1.0", "1.0.0"))  // 1.0.0 == 1.0.0 (missing parts default to 0)
		assert.Equal(t, 0, compareVersions("1", "1.0.0"))    // 1.0.0 == 1.0.0 (missing parts default to 0)
	})
}

func TestIsUpdateAvailable(t *testing.T) {
	// Save original function
	originalCompareVersions := compareVersions
	originalInternalAppMetaDataVersion := internalAppMetaDataVersion

	t.Cleanup(func() {
		compareVersions = originalCompareVersions
		internalAppMetaDataVersion = originalInternalAppMetaDataVersion
	})

	t.Run("latestTag is empty", func(t *testing.T) {
		assert.False(t, isUpdateAvailable(""))
	})

	t.Run("update available", func(t *testing.T) {
		internalAppMetaDataVersion = "1.0.0"
		compareVersions = func(current, latest string) int { return -1 }
		assert.True(t, isUpdateAvailable("1.0.1"))
	})

	t.Run("no update available", func(t *testing.T) {
		internalAppMetaDataVersion = "1.0.0"
		compareVersions = func(current, latest string) int { return 0 }
		assert.False(t, isUpdateAvailable("1.0.0"))
	})

	t.Run("current is newer", func(t *testing.T) {
		internalAppMetaDataVersion = "1.0.1"
		compareVersions = func(current, latest string) int { return 1 }
		assert.False(t, isUpdateAvailable("1.0.0"))
	})
}

func TestPrintUpdateMessage(t *testing.T) {
	// Save original functions
	originalTermGetSize := termGetSize
	originalInternalAppMetaDataVersion := internalAppMetaDataVersion

	t.Cleanup(func() {
		termGetSize = originalTermGetSize
		internalAppMetaDataVersion = originalInternalAppMetaDataVersion
	})

	// Helper to capture stdout
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

	t.Run("prints update message with default width", func(t *testing.T) {
		// Arrange
		termGetSize = func(fd int) (width, height int, err error) { return 0, 0, errors.New("error getting size") }
		internalAppMetaDataVersion = "1.0.0"

		buf, restoreOutput := captureOutput()
		// Act
		printUpdateMessage("1.0.1")

		// Assert
		restoreOutput()
		output := buf.String()
		assert.Contains(t, output, "Prasmoid update available! 1.0.0 → 1.0.1")
		assert.Contains(t, output, "Run `prasmoid upgrade` to update")
		// Check for default width (70) - this is tricky with color codes, but we can check line length roughly
		lines := strings.Split(output, "\n")
		assert.GreaterOrEqual(t, len(lines[0]), 70) // First line should be at least 70 chars wide
	})

	t.Run("prints update message with custom width", func(t *testing.T) {
		// Arrange
		termGetSize = func(fd int) (width, height int, err error) { return 100, 20, nil }
		internalAppMetaDataVersion = "1.0.0"
		buf, restoreOutput := captureOutput()

		// Act
		printUpdateMessage("1.0.1")

		// Assert
		restoreOutput()
		output := buf.String()
		assert.Contains(t, output, "Prasmoid update available! 1.0.0 → 1.0.1")
		assert.Contains(t, output, "Run `prasmoid upgrade` to update")
		// Check for custom width (100)
		lines := strings.Split(output, "\n")
		assert.GreaterOrEqual(t, len(lines[0]), 100) // First line should be at least 100 chars wide
	})
}
