package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/PRASSamin/prasmoid/internal"
	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

// --- Mocks and Test Setup ---

var testExeHash = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" // sha256 of empty string

func setupUpdateCheckerTests(t *testing.T) func() {
	// Backup originals
	origReadUpdateCache := readUpdateCache
	origWriteUpdateCache := writeUpdateCache
	origTimeParse := timeParse
	origTimeSince := timeSince
	origFetchURL := fetchURL
	origIsUpdateAvailable := isUpdateAvailable
	origPrintUpdateMessage := printUpdateMessage
	origOsExecutable := osExecutable
	origCalculateFileSHA256 := calculateFileSHA256
	origInternalVersion := internalAppMetaDataVersion
	origHttpGet := httpGet

	// Restore at the end of the test
	return func() {
		readUpdateCache = origReadUpdateCache
		writeUpdateCache = origWriteUpdateCache
		timeParse = origTimeParse
		timeSince = origTimeSince
		fetchURL = origFetchURL
		isUpdateAvailable = origIsUpdateAvailable
		printUpdateMessage = origPrintUpdateMessage
		osExecutable = origOsExecutable
		calculateFileSHA256 = origCalculateFileSHA256
		internal.AppMetaData.Version = origInternalVersion
		httpGet = origHttpGet
	}
}

// --- Individual Function Tests ---

func TestParseChecksums(t *testing.T) {
	checksums := `
e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855  prasmoid
9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08  prasmoid-portable
`
	t.Run("finds standard asset", func(t *testing.T) {
		hash := parseChecksums(checksums, "prasmoid")
		assert.Equal(t, "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", hash)
	})
	t.Run("finds portable asset", func(t *testing.T) {
		hash := parseChecksums(checksums, "prasmoid-portable")
		assert.Equal(t, "9f86d081884c7d659a2feaa0c55ad015a3bf4f1b2b0b822cd15d6c15b0f00a08", hash)
	})
	t.Run("returns empty for missing asset", func(t *testing.T) {
		hash := parseChecksums(checksums, "missing-asset")
		assert.Empty(t, hash)
	})
}

func TestIsUpdateAvailable(t *testing.T) {
	cleanup := setupUpdateCheckerTests(t)
	defer cleanup()

	t.Run("returns false if latest hash is empty", func(t *testing.T) {
		isAvailable, _ := isUpdateAvailable("")
		assert.False(t, isAvailable)
	})

	t.Run("returns false if os.Executable fails", func(t *testing.T) {
		osExecutable = func() (string, error) { return "", errors.New("exec error") }
		isAvailable, _ := isUpdateAvailable("somehash")
		assert.False(t, isAvailable)
	})

	t.Run("returns false if calculateFileSHA256 fails", func(t *testing.T) {
		osExecutable = func() (string, error) { return "/path/to/exe", nil }
		calculateFileSHA256 = func(filePath string) (string, error) { return "", errors.New("hash error") }
		isAvailable, _ := isUpdateAvailable("somehash")
		assert.False(t, isAvailable)
	})

	t.Run("returns false and current hash if hashes match", func(t *testing.T) {
		osExecutable = func() (string, error) { return "/path/to/exe", nil }
		calculateFileSHA256 = func(filePath string) (string, error) { return testExeHash, nil }
		isAvailable, currentHash := isUpdateAvailable(testExeHash)
		assert.False(t, isAvailable)
		assert.Equal(t, testExeHash, currentHash)
	})

	t.Run("returns true and current hash if hashes differ", func(t *testing.T) {
		osExecutable = func() (string, error) { return "/path/to/exe", nil }
		calculateFileSHA256 = func(filePath string) (string, error) { return testExeHash, nil }
		isAvailable, currentHash := isUpdateAvailable("differenthash")
		assert.True(t, isAvailable)
		assert.Equal(t, testExeHash, currentHash)
	})
}

func TestPrintUpdateMessage(t *testing.T) {
	cleanup := setupUpdateCheckerTests(t)
	defer cleanup()

	// Capture output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	color.Output = w

	printUpdateMessage("2.0.0", "newhash12345", "oldhash67890")

	_ = w.Close()
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	os.Stdout = oldStdout
	output := buf.String()

	assert.Contains(t, output, fmt.Sprintf("%s.oldh → 2.0.0.newh", strings.Split(internal.AppMetaData.Version, "-")[0]))
}

func TestCheckForUpdates(t *testing.T) {
	cleanup := setupUpdateCheckerTests(t)
	defer cleanup()

	var printCalled bool
	printUpdateMessage = func(latest, latestHash, currentHash string) {
		printCalled = true
	}

	// Restore original isUpdateAvailable for this test block
	origIsUpdateAvailable := isUpdateAvailable
	defer func() { isUpdateAvailable = origIsUpdateAvailable }()

	t.Run("cache is fresh and no update available", func(t *testing.T) {
		printCalled = false
		readUpdateCache = func() (map[string]interface{}, error) {
			return map[string]interface{}{
				"last_checked": time.Now().Format(time.RFC3339),
				"latest_hash":  testExeHash,
			}, nil
		}
		timeParse = func(layout, value string) (time.Time, error) { return time.Now(), nil }
		timeSince = func(t time.Time) time.Duration { return 1 * time.Hour }
		isUpdateAvailable = func(latestHash string) (bool, string) { return false, testExeHash }

		CheckForUpdates()

		assert.False(t, printCalled)
	})

	t.Run("cache is fresh and update is available", func(t *testing.T) {
		printCalled = false
		readUpdateCache = func() (map[string]interface{}, error) {
			return map[string]interface{}{
				"last_checked": time.Now().Format(time.RFC3339),
				"latest_hash":  "new_hash",
				"latest_tag":   "2.0.0",
			}, nil
		}
		timeParse = func(layout, value string) (time.Time, error) { return time.Now(), nil }
		timeSince = func(t time.Time) time.Duration { return 1 * time.Hour }
		isUpdateAvailable = func(latestHash string) (bool, string) { return true, testExeHash }

		CheckForUpdates()

		assert.True(t, printCalled)
	})

	t.Run("cache is stale, network fetch succeeds, update available", func(t *testing.T) {
		printCalled = false
		readUpdateCache = func() (map[string]interface{}, error) { return nil, errors.New("cache miss") }
		httpGet = func(rawURL string) (*http.Response, error) {
			if strings.Contains(rawURL, "releases/latest") {
				json := `{"tag_name":"2.0.0", "assets":[{"name":"SHA256SUMS", "browser_download_url":"http://localhost/SHA256SUMS"}]}`
				body := io.NopCloser(strings.NewReader(json))
				return &http.Response{StatusCode: 200, Body: body}, nil
			}
			if strings.Contains(rawURL, "SHA256SUMS") {
				body := io.NopCloser(strings.NewReader("new_hash  prasmoid"))
				return &http.Response{StatusCode: 200, Body: body}, nil
			}
			return nil, errors.New("unexpected URL")
		}
		var cacheWritten bool
		writeUpdateCache = func(tag, hash string) { cacheWritten = true }
		isUpdateAvailable = func(latestHash string) (bool, string) { return true, testExeHash }

		CheckForUpdates()

		assert.True(t, cacheWritten)
		assert.True(t, printCalled)
	})
}

func TestCalculateFileSHA256(t *testing.T) {
	t.Run("calculates hash of a file", func(t *testing.T) {
		content := "hello world"
		tmpfile, err := os.CreateTemp("", "sha_test_*")
		assert.NoError(t, err)
		defer func() { _ = os.Remove(tmpfile.Name()) }()

		_, err = tmpfile.WriteString(content)
		assert.NoError(t, err)
		assert.NoError(t, tmpfile.Close())

		hash, err := calculateFileSHA256(tmpfile.Name())
		assert.NoError(t, err)
		// sha256sum of "hello world"
		assert.Equal(t, "b94d27b9934d3e08a52e52d7da7dabfac484efe37a5380ee9088f7ace2efcde9", hash)
	})

	t.Run("returns error for non-existent file", func(t *testing.T) {
		_, err := calculateFileSHA256("/non/existent/file")
		assert.Error(t, err)
	})
}

func TestGetCacheFilePath(t *testing.T) {
	cleanup := setupUpdateCheckerTests(t)
	defer cleanup()

	t.Run("uses user cache dir", func(t *testing.T) {
		osUserCacheDir = func() (string, error) { return "/fake/cache", nil }
		assert.Equal(t, "/fake/cache/.prasmoid", GetCacheFilePath())
	})

	t.Run("falls back to temp dir", func(t *testing.T) {
		osUserCacheDir = func() (string, error) { return "", errors.New("cache dir error") }
		osTempDir = func() string { return "/fake/tmp" }
		assert.Equal(t, "/fake/tmp/.prasmoid", GetCacheFilePath())
	})
}

func TestReadWriteUpdateCache(t *testing.T) {
	cleanup := setupUpdateCheckerTests(t)
	defer cleanup()

	tmpfile, err := os.CreateTemp("", "cache_test_*")
	assert.NoError(t, err)
	defer func() { _ = os.Remove(tmpfile.Name()) }()

	GetCacheFilePath = func() string {
		return tmpfile.Name()
	}

	t.Run("writes and reads cache successfully", func(t *testing.T) {
		testTag := "v1.2.3"
		testHash := "testhash123"
		writeUpdateCache(testTag, testHash)

		cache, err := readUpdateCache()
		assert.NoError(t, err)
		assert.Equal(t, testTag, cache["latest_tag"])
		assert.Equal(t, testHash, cache["latest_hash"])
		_, hasTimestamp := cache["last_checked"]
		assert.True(t, hasTimestamp)
	})

	t.Run("read returns error for non-existent file", func(t *testing.T) {
		GetCacheFilePath = func() string { return "/non/existent/cache/file" }
		_, err := readUpdateCache()
		assert.Error(t, err)
	})
}
