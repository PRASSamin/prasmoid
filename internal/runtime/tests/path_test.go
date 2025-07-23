package runtime_tests

import (
	"testing"

	rtime "github.com/PRASSamin/prasmoid/internal/runtime"
)

func TestPathModule(t *testing.T) {
	vm := rtime.NewRuntime()
	_, err := vm.RunString(`const path = require('path');`)
	if err != nil {
		t.Fatalf("Failed to declare path: %v", err)
	}

	t.Run("resolve", func(t *testing.T) {
		script := `path.resolve('/foo/bar', './baz');`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		if val.String() != "/foo/bar/baz" {
			t.Errorf("Expected /foo/bar/baz, got %s", val.String())
		}
	})

	t.Run("normalize", func(t *testing.T) {
		script := `path.normalize('/foo/bar//baz/..');`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		if val.String() != "/foo/bar" {
			t.Errorf("Expected /foo/bar, got %s", val.String())
		}
	})

	t.Run("isAbsolute", func(t *testing.T) {
		script := `path.isAbsolute('/foo/bar');`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		if !val.ToBoolean() {
			t.Errorf("Expected true, got %t", val.ToBoolean())
		}
	})

	t.Run("join", func(t *testing.T) {
		script := `path.join('/foo', 'bar', 'baz/asdf', 'quux', '..');`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		if val.String() != "/foo/bar/baz/asdf" {
			t.Errorf("Expected /foo/bar/baz/asdf, got %s", val.String())
		}
	})

	t.Run("relative", func(t *testing.T) {
		script := `path.relative('/data/orandea/test/aaa', '/data/orandea/test/bbb');`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		if val.String() != "../bbb" {
			t.Errorf("Expected ../bbb, got %s", val.String())
		}
	})

	t.Run("dirname", func(t *testing.T) {
		script := `path.dirname('/foo/bar/baz/asdf/quux.html');`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		if val.String() != "/foo/bar/baz/asdf" {
			t.Errorf("Expected /foo/bar/baz/asdf, got %s", val.String())
		}
	})

	t.Run("basename", func(t *testing.T) {
		script := `path.basename('/foo/bar/baz/asdf/quux.html');`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		if val.String() != "quux.html" {
			t.Errorf("Expected quux.html, got %s", val.String())
		}
	})

	t.Run("extname", func(t *testing.T) {
		script := `path.extname('index.html');`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		if val.String() != ".html" {
			t.Errorf("Expected .html, got %s", val.String())
		}
	})

	t.Run("parse", func(t *testing.T) {
		script := `path.parse('/home/user/dir/file.txt');`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		obj := val.ToObject(vm)
		if obj.Get("root").String() != "/" ||
			obj.Get("dir").String() != "/home/user/dir" ||
			obj.Get("base").String() != "file.txt" ||
			obj.Get("ext").String() != ".txt" ||
			obj.Get("name").String() != "file" {
			t.Errorf("Parse result mismatch: %+v", obj.Export())
		}
	})

	t.Run("matchesGlob", func(t *testing.T) {
		script := `path.matchesGlob('foo.txt', '*.txt');`
		val, err := vm.RunString(script)
		if err != nil {
			t.Fatalf("vm.RunString() failed: %v", err)
		}
		if !val.ToBoolean() {
			t.Errorf("Expected true, got %t", val.ToBoolean())
		}
	})
}
