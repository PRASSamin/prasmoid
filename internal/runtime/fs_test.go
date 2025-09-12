package runtime

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/dop251/goja"
)

func setupTestVM(t *testing.T) (*goja.Runtime, func()) {
	vm := NewRuntime()

	tmpDir, err := os.MkdirTemp("", "runtime-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	originalWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory to %s: %v", tmpDir, err)
	}

	cleanup := func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("Failed to restore original directory: %v", err)
		}
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("Failed to remove temporary directory: %v", err)
		}
	}

	return vm, cleanup
}

func TestFSModule(t *testing.T) {
	t.Run("writeFileSync and readFileSync", func(t *testing.T) {
		vm, cleanup := setupTestVM(t)
		defer cleanup()

		script := `
			fs.writeFileSync('test.txt', 'hello world');
			fs.readFileSync('test.txt');
		`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}

		if val.String() != "hello world" {
			t.Errorf("Expected to read 'hello world', but got '%s'", val.String())
		}
	})

	t.Run("appendFileSync", func(t *testing.T) {
		vm, cleanup := setupTestVM(t)
		defer cleanup()

		script := `
			fs.writeFileSync('test.txt', 'hello');
			fs.appendFileSync('test.txt', ' world');
			fs.readFileSync('test.txt');
		`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}

		if val.String() != "hello world" {
			t.Errorf("Expected to read 'hello world', but got '%s'", val.String())
		}
	})

	t.Run("existsSync", func(t *testing.T) {
		vm, cleanup := setupTestVM(t)
		defer cleanup()

		script := `
			fs.writeFileSync('exists.txt', '');
			const exists1 = fs.existsSync('exists.txt');
			const exists2 = fs.existsSync('does-not-exist.txt');
			exists1 && !exists2;
		`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}

		if !val.ToBoolean() {
			t.Error("Expected existsSync to return true for existing and false for non-existing file")
		}
	})

	t.Run("readdirSync", func(t *testing.T) {
		vm, cleanup := setupTestVM(t)
		defer cleanup()

		if err := os.WriteFile("file1.txt", []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile("file2.txt", []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir("dir1", 0755); err != nil {
			t.Fatal(err)
		}

		script := `fs.readdirSync('.');`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}

		exported := val.Export()
		var files []string
		if exportedSlice, ok := exported.([]interface{}); ok {
			for _, item := range exportedSlice {
				files = append(files, item.(string))
			}
		} else {
			t.Fatalf("Expected readdirSync to return an array, but it did not")
		}

		sort.Strings(files)
		expectedFiles := []string{"dir1", "file1.txt", "file2.txt"}
		sort.Strings(expectedFiles)

		if !reflect.DeepEqual(files, expectedFiles) {
			t.Errorf("Expected readdirSync to return %v, but got %v", expectedFiles, files)
		}
	})

	t.Run("mkdirSync, rmdirSync, rmSync", func(t *testing.T) {
		vm, cleanup := setupTestVM(t)
		defer cleanup()

		// Test mkdirSync
		script := `fs.mkdirSync('newDir');`
		_, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() for mkdirSync failed: %v", err)
		}
		if _, err := os.Stat("newDir"); os.IsNotExist(err) {
			t.Error("Expected mkdirSync to create 'newDir', but it did not")
		}

		// Test rmdirSync on empty dir
		script = `fs.rmdirSync('newDir');`
		_, err = vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() for rmdirSync failed: %v", err)
		}
		if _, err := os.Stat("newDir"); !os.IsNotExist(err) {
			t.Error("Expected rmdirSync to remove 'newDir', but it did not")
		}

		// Test rmSync recursive
		script = `
			fs.mkdirSync('newDir');
			fs.writeFileSync('newDir/file.txt', 'content');
			fs.mkdirSync('newDir/subDir');
			fs.rmSync('newDir', { recursive: true });
		`
		_, err = vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() for rmSync failed: %v", err)
		}
		if _, err := os.Stat("newDir"); !os.IsNotExist(err) {
			t.Error("Expected rmSync recursive to remove 'newDir', but it did not")
		}
	})

	t.Run("copyFileSync", func(t *testing.T) {
		vm, cleanup := setupTestVM(t)
		defer cleanup()

		script := `
			fs.writeFileSync('source.txt', 'hello copy');
			fs.copyFileSync('source.txt', 'dest.txt');
			fs.readFileSync('dest.txt');
		`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}

		if val.String() != "hello copy" {
			t.Errorf("Expected to read 'hello copy', but got '%s'", val.String())
		}
	})

	t.Run("renameSync", func(t *testing.T) {
		vm, cleanup := setupTestVM(t)
		defer cleanup()

		script := `
			fs.writeFileSync('old.txt', 'hello rename');
			fs.renameSync('old.txt', 'new.txt');
			fs.readFileSync('new.txt');
		`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}

		if val.String() != "hello rename" {
			t.Errorf("Expected to read 'hello rename', but got '%s'", val.String())
		}
		if _, err := os.Stat("old.txt"); !os.IsNotExist(err) {
			t.Error("Expected old.txt to be removed after rename, but it still exists")
		}
	})

	t.Run("unlinkSync", func(t *testing.T) {
		vm, cleanup := setupTestVM(t)
		defer cleanup()

		script := `
			fs.writeFileSync('deleteme.txt', 'content');
			fs.unlinkSync('deleteme.txt');
			fs.existsSync('deleteme.txt');
		`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}

		if val.ToBoolean() {
			t.Error("Expected unlinkSync to remove file, but it still exists")
		}
	})

	t.Run("realpathSync", func(t *testing.T) {
		vm, cleanup := setupTestVM(t)
		defer cleanup()

		targetFile := "target.txt"
		linkFile := "link.txt"
		if err := os.WriteFile(targetFile, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(targetFile, linkFile); err != nil {
			t.Fatal(err)
		}

		script := fmt.Sprintf(`fs.realpathSync('%s');`, linkFile)
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}

		expectedPath, _ := filepath.Abs(targetFile)
		if val.String() != expectedPath {
			t.Errorf("Expected realpath to be '%s', but got '%s'", expectedPath, val.String())
		}
	})

	t.Run("readlinkSync", func(t *testing.T) {
		vm, cleanup := setupTestVM(t)
		defer cleanup()

		target := "target.txt"
		link := "link.txt"
		if err := os.WriteFile(target, []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(target, link); err != nil {
			t.Fatal(err)
		}

		script := fmt.Sprintf(`fs.readlinkSync('%s');`, link)
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}

		if val.String() != target {
			t.Errorf("Expected readlinkSync to return '%s', got '%s'", target, val.String())
		}
	})

	t.Run("cpSync", func(t *testing.T) {
		vm, cleanup := setupTestVM(t)
		defer cleanup()

		script := `
			// File copy
			fs.writeFileSync('original.txt', 'hello file');
			fs.cpSync('original.txt', 'copy.txt');

			// Dir copy
			fs.mkdirSync('mydir');
			fs.writeFileSync('mydir/file1.txt', 'file one');
			fs.mkdirSync('mydir/subdir');
			fs.writeFileSync('mydir/subdir/deep.txt', 'deep dive');
			fs.cpSync('mydir', 'mycopy');

			// Verify
			JSON.stringify({
				copyText: fs.readFileSync('copy.txt'),
				file1: fs.readFileSync('mycopy/file1.txt'),
				deep: fs.readFileSync('mycopy/subdir/deep.txt')
			});
		`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}

		var result map[string]string
		if err := json.Unmarshal([]byte(val.String()), &result); err != nil {
			t.Fatalf("Failed to unmarshal result: %v", err)
		}

		expected := map[string]string{
			"copyText": "hello file",
			"file1":    "file one",
			"deep":     "deep dive",
		}
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("cpSync result mismatch. Got %v, want %v", result, expected)
		}
	})

	t.Run("globSync", func(t *testing.T) {
		vm, cleanup := setupTestVM(t)
		defer cleanup()

		if err := os.Mkdir("a", 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile("a/f1.js", nil, 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile("a/f2.ts", nil, 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir("b", 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile("b/f3.js", nil, 0644); err != nil {
			t.Fatal(err)
		}

		script := `fs.globSync('**/*.js')`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}

		var result []string
		if err := vm.ExportTo(val, &result); err != nil {
			t.Fatalf("Failed to export glob result: %v", err)
		}

		sort.Strings(result)
		expected := []string{"a/f1.js", "b/f3.js"}
		// On some systems, path separator might be different
		for i, p := range expected {
			expected[i] = filepath.FromSlash(p)
		}
		sort.Strings(expected)

		if !reflect.DeepEqual(result, expected) {
			t.Errorf("Expected glob result %v, got %v", expected, result)
		}
	})

	t.Run("mkdtempSync", func(t *testing.T) {
		vm, cleanup := setupTestVM(t)
		defer cleanup()

		script := `fs.mkdtempSync('test-prefix-');`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}

		dirPath := val.String()
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			t.Errorf("mkdtempSync should have created directory '%s', but it doesn't exist", dirPath)
		}
		if !strings.HasPrefix(filepath.Base(dirPath), "test-prefix-") {
			t.Errorf("Expected dir to have prefix 'test-prefix-', but got '%s'", filepath.Base(dirPath))
		}
	})

	t.Run("symlinkSync", func(t *testing.T) {
		vm, cleanup := setupTestVM(t)
		defer cleanup()

		target := "original.txt"
		link := "link.txt"
		if err := os.WriteFile(target, []byte("hello"), 0644); err != nil {
			t.Fatal(err)
		}

		script := fmt.Sprintf(`fs.symlinkSync('%s', '%s');`, target, link)
		_, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("symlinkSync failed: %v", err)
		}

		info, err := os.Lstat(link)
		if err != nil {
			t.Fatalf("link file not found: %v", err)
		}
		if info.Mode()&os.ModeSymlink == 0 {
			t.Error("expected a symlink but got a regular file")
		}
		readTarget, err := os.Readlink(link)
		if err != nil {
			t.Fatalf("failed to read symlink: %v", err)
		}
		if readTarget != target {
			t.Errorf("expected symlink to point to %q but got %q", target, readTarget)
		}
	})

	t.Run("statSync and Stats class", func(t *testing.T) {
		vm, cleanup := setupTestVM(t)
		defer cleanup()

		if err := os.WriteFile("testfile.txt", []byte("hello"), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir("testdir", 0755); err != nil {
			t.Fatal(err)
		}

		script := `
			const fileStats = fs.statSync('testfile.txt');
			const dirStats = fs.statSync('testdir');
			JSON.stringify({
				file: {
					isFile: fileStats.isFile(),
					isDirectory: fileStats.isDirectory(),
					size: fileStats.size
				},
				dir: {
					isFile: dirStats.isFile(),
					isDirectory: dirStats.isDirectory(),
				}
			})
		`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}

		type statsResult struct {
			IsFile      bool `json:"isFile"`
			IsDirectory bool `json:"isDirectory"`
			Size        int  `json:"size"`
		}
		var result struct {
			File statsResult `json:"file"`
			Dir  statsResult `json:"dir"`
		}
		if err := json.Unmarshal([]byte(val.String()), &result); err != nil {
			t.Fatalf("Failed to unmarshal stat result: %v", err)
		}

		if !result.File.IsFile {
			t.Error("fileStats.isFile() should be true")
		}
		if result.File.IsDirectory {
			t.Error("fileStats.isDirectory() should be false")
		}
		if result.File.Size != 5 {
			t.Errorf("fileStats.size should be 5, got %d", result.File.Size)
		}
		if result.Dir.IsFile {
			t.Error("dirStats.isFile() should be false")
		}
		if !result.Dir.IsDirectory {
			t.Error("dirStats.isDirectory() should be true")
		}
	})

	t.Run("error handling", func(t *testing.T) {
		vm, cleanup := setupTestVM(t)
		defer cleanup()

		// Test cases where the function returns an error string to JS
		t.Run("JS error strings", func(t *testing.T) {
			testCases := []struct {
				name          string
				script        string
				expectedValue string // The string value returned by the JS function
			}{
				{"readFileSync missing path", `fs.readFileSync()`, "fs.readFileSync: missing path"},
				{"readFileSync non-existent", `fs.readFileSync('no-such-file.txt')`, "no such file or directory"},
				{"writeFileSync missing content", `fs.writeFileSync('file.txt')`, "fs.writeFileSync: missing path or content"},
				{"appendFileSync missing content", `fs.appendFileSync('file.txt')`, "fs.appendFileSync: missing path or content"},
				{"existsSync missing path", `fs.existsSync()`, "fs.existsSync: missing path"},
				{"readdirSync missing path", `fs.readdirSync()`, "fs.readdirSync: missing path"},
				{"mkdirSync missing path", `fs.mkdirSync()`, "fs.mkdirSync: missing path"},
				{"rmSync missing path", `fs.rmSync()`, "fs.rmSync: missing path"},
				{"copyFileSync missing args", `fs.copyFileSync('src')`, "fs.copyFileSync: missing src or dest"},
				{"renameSync missing args", `fs.renameSync('old')`, "fs.renameSync: missing oldPath or newPath"},
				{"unlinkSync missing path", `fs.unlinkSync()`, "fs.unlinkSync: missing path"},
				{"realpathSync missing path", `fs.realpathSync()`, "fs.realpathSync: missing path"},
				{"readlinkSync missing path", `fs.readlinkSync()`, "fs.readlinkSync: missing path"},
				{"cpSync missing args", `fs.cpSync('src')`, "fs.cpSync: missing src or dest"},
				{"globSync missing pattern", `fs.globSync()`, "fs.globSync: missing pattern"},
				{"mkdtempSync missing prefix", `fs.mkdtempSync()`, "fs.mkdtempSync: missing prefix"},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					val, err := vm.RunString(tc.script)
					if err != nil {
						t.Fatalf("vm.RunString() unexpectedly failed: %v", err)
					}
					if !strings.Contains(val.String(), tc.expectedValue) {
						t.Errorf("Expected JS value to contain '%s', but got '%s'", tc.expectedValue, val.String())
					}
				})
			}
		})

		// Test cases where the function causes a Goja error (panic in Go)
		t.Run("Goja errors (panics)", func(t *testing.T) {
			testCases := []struct {
				name          string
				script        string
				expectedError string // Substring of the Goja error message
			}{
				{"symlinkSync missing args", `fs.symlinkSync('target')`, "fs.symlinkSync requires 2 arguments: target and path"},
				{"statSync missing path", `fs.statSync()`, "fs.statSync: missing path"},
				{"statSync non-existent", `fs.statSync('no-such-file')`, "fs.statSync error: stat no-such-file: no such file or directory"},
				{"watch missing listener", `fs.watch('.', null)`, "watch: last argument must be a function"},
				{"watchFile missing listener", `fs.watchFile('.', null)`, "watchFile: last argument must be a function"},
			}

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					_, err := vm.RunString(tc.script)
					if err == nil {
						t.Errorf("Expected a Goja error, but got none")
						return
					}
					if !strings.Contains(err.Error(), tc.expectedError) {
						t.Errorf("Expected Goja error to contain '%s', but got '%v'", tc.expectedError, err)
					}
				})
			}
		})
	})

	t.Run("not implemented functions", func(t *testing.T) {
		vm, cleanup := setupTestVM(t)
		defer cleanup()

		script := `fs.accessSync('a.txt')`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		expected := "fs.accessSync is not implemented in this runtime"
		if val.String() != expected {
			t.Errorf("Expected '%s', got '%s'", expected, val.String())
		}
	})
}
