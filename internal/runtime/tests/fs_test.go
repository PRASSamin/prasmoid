package runtime_tests

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	rtime "github.com/PRASSamin/prasmoid/internal/runtime"
	"github.com/dop251/goja"
)

func setupTestVM(t *testing.T, moduleName string, registerFunc func(*goja.Runtime, *goja.Object)) (*goja.Runtime, func()) {
	vm := rtime.NewRuntime()

	tmpDir, err := os.MkdirTemp("", "runtime-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	originalWd, _ := os.Getwd()
	os.Chdir(tmpDir)

	cleanup := func() {
		os.Chdir(originalWd)
		os.RemoveAll(tmpDir)
	}

	return vm, cleanup
}

func TestFSModule(t *testing.T) {
	t.Run("writeFileSync and readFileSync", func(t *testing.T) {
		vm, cleanup := setupTestVM(t, "fs", rtime.FS)
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
		vm, cleanup := setupTestVM(t, "fs", rtime.FS)
		defer cleanup()

		script := `
			fs.writeFileSync('test.txt', 'hello world');
			fs.appendFileSync('test.txt', ' hello');
			fs.readFileSync('test.txt');
		`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}

		if val.String() != "hello world hello" {
			t.Errorf("Expected to read 'hello world hello', but got '%s'", val.String())
		}
	})

	t.Run("existsSync", func(t *testing.T) {
		vm, cleanup := setupTestVM(t, "fs", rtime.FS)
		defer cleanup()

		script := `
			fs.writeFileSync('exists.txt', '');
			fs.existsSync('exists.txt');
		`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}

		if !val.ToBoolean() {
			t.Error("Expected existsSync to return true, but it returned false")
		}
	})

	t.Run("readdirSync", func(t *testing.T) {
		vm, cleanup := setupTestVM(t, "fs", rtime.FS)
		defer cleanup()

		if err := os.WriteFile("file1.txt", []byte(""), 0644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile("file2.txt", []byte(""), 0644); err != nil {
			t.Fatal(err)
		}

		script := `
			fs.readdirSync('.');
		`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}

		// Export the array to a Go slice
		exported := val.Export()
		if files, ok := exported.([]interface{}); ok {
			t.Logf("Files found by readdirSync: %v", files) // Added logging
			if len(files) != 2 {
				t.Errorf("Expected readdirSync to return 2 files, but got %d", len(files))
			}
		} else {
			t.Errorf("Expected readdirSync to return an array, but it did not")
		}
	})

	t.Run("mkdirSync and rmdirSync and rmSync with recursive option", func(t *testing.T) {
		vm, cleanup := setupTestVM(t, "fs", rtime.FS)
		defer cleanup()

		script := `
			fs.mkdirSync('newDir');
			fs.mkdirSync('newDir/newDir2');
			fs.mkdirSync('newDir/newDir2/newDir3');
			fs.mkdirSync('newDir/newDir2/newDir3/newDir4');
			fs.writeFileSync('newDir/newDir2/newDir3/newDir4/file.txt', '');
		`
		_, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}

		if _, err := os.Stat("newDir"); os.IsNotExist(err) {
			t.Error("Expected mkdirSync to create 'newDir', but it did not")
		}

		script = `
			fs.rmSync('newDir', { recursive: true });
		`
		_, err = vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}

		if _, err := os.Stat("newDir"); !os.IsNotExist(err) {
			t.Error("Expected rmSync to remove 'newDir', but it did not")
		}
	})

	t.Run("copyFileSync", func(t *testing.T) {
		vm, cleanup := setupTestVM(t, "fs", rtime.FS)
		defer cleanup()

		script := `
			fs.writeFileSync('test.txt', 'hello world');
			fs.copyFileSync('test.txt', 'test2.txt');
			fs.readFileSync('test2.txt');
		`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}

		if val.String() != "hello world" {
			t.Errorf("Expected to read 'hello world', but got '%s'", val.String())
		}
	})

	t.Run("renameSync", func(t *testing.T) {
		vm, cleanup := setupTestVM(t, "fs", rtime.FS)
		defer cleanup()

		script := `
			fs.writeFileSync('test.txt', 'hello world');
			fs.renameSync('test.txt', 'test2.txt');
			fs.readFileSync('test2.txt');
		`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}

		if val.String() != "hello world" {
			t.Errorf("Expected to read 'hello world', but got '%s'", val.String())
		}
	})

	t.Run("unlinkSync", func(t *testing.T) {
		vm, cleanup := setupTestVM(t, "fs", rtime.FS)
		defer cleanup()

		script := `
			fs.writeFileSync('test.txt', 'hello world');
			fs.unlinkSync('test.txt');
		`
		_, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}

		if _, err := os.Stat("test.txt"); !os.IsNotExist(err) {
			t.Error("Expected unlinkSync to remove 'test.txt', but it did not")
		}
	})

	t.Run("realpathSync", func(t *testing.T) {
		vm, cleanup := setupTestVM(t, "fs", rtime.FS)
		defer cleanup()

		script := `
			fs.writeFileSync('test.txt', 'hello world');
			fs.realpathSync('test.txt');
		`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}

		// Verify that the returned path is an absolute path
		absPath, _ := filepath.Abs(val.String())
		if !filepath.IsAbs(absPath) {
			t.Errorf("Expected absolute path, got: %s", absPath)
		}

		// Verify that the file exists at the returned path
		if _, err := os.Stat(absPath); err != nil {
			t.Errorf("File does not exist at returned path: %s", absPath)
		}
	})

	t.Run("readlinkSync", func(t *testing.T) {
		vm, cleanup := setupTestVM(t, "fs", rtime.FS)
		defer cleanup()

		// Create a target file
		if err := os.WriteFile("target.txt", []byte("hello world"), 0644); err != nil {
			t.Fatalf("Failed to create target file: %v", err)
		}
		defer os.Remove("target.txt")

		// Create a symbolic link
		if err := os.Symlink("target.txt", "link.txt"); err != nil {
			t.Fatalf("Failed to create symlink: %v", err)
		}
		defer os.Remove("link.txt")

		script := `
			fs.readlinkSync('link.txt');
		`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}

		// Verify that readlink returns the correct target path
		targetPath := val.String()
		if targetPath != "target.txt" {
			t.Errorf("Expected target.txt, got: %s", targetPath)
		}
	})

	t.Run("mkdtempSync", func(t *testing.T) {
		vm, cleanup := setupTestVM(t, "fs", rtime.FS)
		defer cleanup()

		script := `
			fs.mkdtempSync('prasmoid-');
		`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}

		// Verify that the directory exists
		if _, err := os.Stat(val.String()); os.IsNotExist(err) {
			t.Errorf("Directory does not exist: %s", val.String())
		}
	})

	t.Run("globSync", func(t *testing.T) {
		vm, cleanup := setupTestVM(t, "fs", rtime.FS)
		defer cleanup()
	
		// Create a test directory structure
		subDir := "subdir"
		if err := os.Mkdir(subDir, 0755); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}
	
		// Create test files
		testFiles := []string{
			"file1.go",
			"file2.go",
			"file3.txt",
			filepath.Join(subDir, "file4.go"),
			filepath.Join(subDir, "file5.txt"),
		}
	
		for _, file := range testFiles {
			if err := os.WriteFile(file, []byte("test content"), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}
		}
	
		// Test various patterns
		patterns := []struct {
			pattern    string
			expected   []string
			description string
		}{
			{
				"**/*.go",
				[]string{
					"file1.go",
					"file2.go",
					"subdir/file4.go",
				},
				"should match all .go files recursively",
			},
			{
				"*.go",
				[]string{
					"file1.go",
					"file2.go",
				},
				"should match .go files in current directory",
			},
			{
				"subdir/*.go",
				[]string{
					"subdir/file4.go",
				},
				"should match .go files in specific directory",
			},
			{
				"*.txt",
				[]string{
					"file3.txt",
				},
				"should match .txt files",
			},
			{
				"nonexistent/*.go",
				[]string{},
				"should return empty array for non-existent directory",
			},
		}
	
		for _, tc := range patterns {
			t.Run(tc.description, func(t *testing.T) {
				script := fmt.Sprintf("fs.globSync('%s');", tc.pattern)
				val, err := vm.RunString(script)
				if err != nil {
					t.Fatalf("vm.RunString() failed: %v", err)
				}
	
				// Convert Goja array to string slice
				gojaArray := val.Export()

				var result []string
				switch v := gojaArray.(type) {
				case []string:
					result = v
				case []interface{}:
					for _, item := range v {
						result = append(result, item.(string))
					}
				default:
					t.Fatalf("unexpected result type: %T", v)
				}
	
				// Sort both slices for comparison
				sort.Strings(result)
				sort.Strings(tc.expected)

				if len(tc.expected) != len(result) && !reflect.DeepEqual(result, tc.expected) {
					t.Errorf("Pattern %q: expected %v, got %v", tc.pattern, tc.expected, result)
				}
			})
		}

		t.Run("existsSync", func(t *testing.T) {
			vm, cleanup := setupTestVM(t, "fs", rtime.FS)
			defer cleanup()

			script := `
				fs.writeFileSync('test.txt', '');
				fs.existsSync('test.txt');
			`
			val, err := vm.RunString(script)
			if err != nil {
				t.Fatalf("vm.RunString() failed: %v", err)
			}

			if !val.ToBoolean() {
				t.Error("Expected existsSync to return true, but it returned false")
			}
		})

		t.Run("existsSync", func(t *testing.T) {
			vm, cleanup := setupTestVM(t, "fs", rtime.FS)
			defer cleanup()
		
			// Make sure we're testing in the correct dir
			originalDir, _ := os.Getwd()
			testDir := t.TempDir()
			os.Chdir(testDir)
			defer os.Chdir(originalDir)
		
			// Test both existing and non-existing file
			script := `
				fs.writeFileSync('file.txt', '');
				let exists1 = fs.existsSync('file.txt');
				let exists2 = fs.existsSync('nope.txt');
				exists1 && !exists2;
			`
		
			val, err := vm.RunString(script)
			if err != nil {
				t.Fatalf("vm.RunString() failed: %v", err)
			}
		
			if !val.ToBoolean() {
				t.Error("Expected existsSync to return true for existing file and false for missing file")
			}
		})
		
		t.Run("cpSync - file and directory copy", func(t *testing.T) {
			vm, cleanup := setupTestVM(t, "fs", rtime.FS)
			defer cleanup()
		
			originalDir, _ := os.Getwd()
			testDir := t.TempDir()
			os.Chdir(testDir)
			defer os.Chdir(originalDir)
		
			script := `
				// Create a single file
				fs.writeFileSync('original.txt', 'hello file');
		
				// Create a directory structure
				fs.mkdirSync('mydir');
				fs.writeFileSync('mydir/file1.txt', 'file one');
				fs.writeFileSync('mydir/file2.txt', 'file two');
		
				fs.mkdirSync('mydir/subdir');
				fs.writeFileSync('mydir/subdir/deep.txt', 'deep dive');
		
				// Copy the file
				fs.cpSync('original.txt', 'copy.txt');
		
				// Copy the whole directory
				fs.cpSync('mydir', 'mycopy', { recursive: true });
		
				// Read copied contents to verify
				let results = {
					copyText: fs.readFileSync('copy.txt').toString(),
					file1: fs.readFileSync('mycopy/file1.txt').toString(),
					file2: fs.readFileSync('mycopy/file2.txt').toString(),
					deep: fs.readFileSync('mycopy/subdir/deep.txt').toString()
				};
				JSON.stringify(results);
			`
		
			val, err := vm.RunString(script)
			if err != nil {
				t.Fatalf("vm.RunString() failed: %v", err)
			}
		
			expected := map[string]string{
				"copyText": "hello file",
				"file1":    "file one",
				"file2":    "file two",
				"deep":     "deep dive",
			}
		
			var got map[string]string
			if err := json.Unmarshal([]byte(val.String()), &got); err != nil {
				t.Fatalf("Failed to unmarshal script result: %v", err)
			}
		
			for key, expectedVal := range expected {
				if got[key] != expectedVal {
					t.Errorf("Mismatch for %s: got '%s', want '%s'", key, got[key], expectedVal)
				}
			}
		})		
	})

	t.Run("symlinkSync", func(t *testing.T) {
		vm, cleanup := setupTestVM(t, "fs", rtime.FS)
		defer cleanup()
		tmpDir := t.TempDir()
		origFile := filepath.Join(tmpDir, "original.txt")
		linkFile := filepath.Join(tmpDir, "link.txt")

		// Create the original file
		if err := os.WriteFile(origFile, []byte("hello world"), 0644); err != nil {
			t.Fatalf("failed to write original file: %v", err)
		}

		// Run symlinkSync JS code
		script := fmt.Sprintf(`fs.symlinkSync(%q, %q);`, origFile, linkFile)
		_, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("symlinkSync failed: %v", err)
		}

		// Check if symlink exists and points correctly
		info, err := os.Lstat(linkFile)
		if err != nil {
			t.Fatalf("link file not found: %v", err)
		}
		if info.Mode()&os.ModeSymlink == 0 {
			t.Error("expected a symlink but got a regular file")
		}

		// Read symlink target and verify
		target, err := os.Readlink(linkFile)
		if err != nil {
			t.Fatalf("failed to read symlink: %v", err)
		}
		if target != origFile {
			t.Errorf("expected symlink to point to %q but got %q", origFile, target)
		}
	})
}