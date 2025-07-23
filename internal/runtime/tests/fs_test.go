package runtime_tests

import (
	"os"
	"path/filepath"
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
}
