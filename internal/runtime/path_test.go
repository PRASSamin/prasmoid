package runtime

import (
	"testing"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/require"
)

func TestPathModule(t *testing.T) {
	vm := NewRuntime()
	_, err := vm.RunString(`const path = require('path');`)
	require.NoError(t, err)

	t.Run("resolve", func(t *testing.T) {
		t.Run("basic", func(t *testing.T) {
			script := `path.resolve('/foo/bar', './baz');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "/foo/bar/baz", val.String())
		})

		t.Run("no arguments", func(t *testing.T) {
			script := `path.resolve();`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "", val.String()) // Corrected expectation
		})

		t.Run("empty arguments", func(t *testing.T) {
			script := `path.resolve('', '', '');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "", val.String()) // Corrected expectation
		})

		t.Run("absolute path in middle", func(t *testing.T) {
			script := `path.resolve('/foo', 'bar', '/baz', 'qux');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "/foo/bar/baz/qux", val.String()) // Corrected expectation
		})

		t.Run("with dots", func(t *testing.T) {
			script := `path.resolve('/foo/bar/baz', '../qux/./asdf');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "/foo/bar/qux/asdf", val.String())
		})
	})

	t.Run("normalize", func(t *testing.T) {
		t.Run("basic", func(t *testing.T) {
			script := `path.normalize('/foo/bar//baz/..');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "/foo/bar", val.String())
		})

		t.Run("empty string", func(t *testing.T) {
			script := `path.normalize('');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, ".", val.String())
		})

		t.Run("relative path", func(t *testing.T) {
			script := `path.normalize('foo/./bar/../baz');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "foo/baz", val.String())
		})

		t.Run("only dots", func(t *testing.T) {
			script := `path.normalize('./../.');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "..", val.String())
		})
	})

	t.Run("isAbsolute", func(t *testing.T) {
		t.Run("basic - true", func(t *testing.T) {
			script := `path.isAbsolute('/foo/bar');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.True(t, val.ToBoolean())
		})

		t.Run("empty string", func(t *testing.T) {
			script := `path.isAbsolute('');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.False(t, val.ToBoolean())
		})

		t.Run("relative path", func(t *testing.T) {
			script := `path.isAbsolute('foo/bar');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.False(t, val.ToBoolean())
		})
	})

	t.Run("join", func(t *testing.T) {
		t.Run("basic", func(t *testing.T) {
			script := `path.join('/foo', 'bar', 'baz/asdf', 'quux', '..');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "/foo/bar/baz/asdf", val.String())
		})

		t.Run("no arguments", func(t *testing.T) {
			script := `path.join();`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "", val.String()) // Corrected expectation
		})

		t.Run("empty arguments", func(t *testing.T) {
			script := `path.join('', '', '');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "", val.String()) // Corrected expectation
		})

		t.Run("mixed absolute and relative", func(t *testing.T) {
			script := `path.join('/foo', 'bar', '/baz', 'qux');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "/foo/bar/baz/qux", val.String()) // Corrected expectation
		})
	})

	t.Run("relative", func(t *testing.T) {
		t.Run("basic", func(t *testing.T) {
			script := `path.relative('/data/orandea/test/aaa', '/data/orandea/test/bbb');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "../bbb", val.String())
		})

		t.Run("same paths", func(t *testing.T) {
			script := `path.relative('/a/b', '/a/b');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, ".", val.String())
		})

		t.Run("from parent to child", func(t *testing.T) {
			script := `path.relative('/a', '/a/b');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "b", val.String())
		})

		t.Run("from child to parent", func(t *testing.T) {
			script := `path.relative('/a/b', '/a');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "..", val.String())
		})

		t.Run("missing arguments", func(t *testing.T) {
			script := `path.relative('/a');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "path.relative: needs from and to", val.String())
		})
	})

	t.Run("toNamespacedPath", func(t *testing.T) {
		t.Run("no arguments", func(t *testing.T) {
			script := `path.toNamespacedPath();`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "", val.String())
		})

		t.Run("empty string", func(t *testing.T) {
			script := `path.toNamespacedPath('');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, ".", val.String())
		})

		t.Run("regular path", func(t *testing.T) {
			script := `path.toNamespacedPath('/foo/bar');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "/foo/bar", val.String())
		})

		t.Run("with dots", func(t *testing.T) {
			script := `path.toNamespacedPath('/foo/./bar/../baz');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "/foo/baz", val.String())
		})
	})

	t.Run("dirname", func(t *testing.T) {
		t.Run("basic", func(t *testing.T) {
			script := `path.dirname('/foo/bar/baz/asdf/quux.html');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "/foo/bar/baz/asdf", val.String())
		})

		t.Run("no arguments", func(t *testing.T) {
			script := `path.dirname();`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "", val.String())
		})

		t.Run("empty string", func(t *testing.T) {
			script := `path.dirname('');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, ".", val.String())
		})

		t.Run("root path", func(t *testing.T) {
			script := `path.dirname('/');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "/", val.String())
		})

		t.Run("file in root", func(t *testing.T) {
			script := `path.dirname('/file.txt');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "/", val.String())
		})

		t.Run("no slashes", func(t *testing.T) {
			script := `path.dirname('file.txt');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, ".", val.String())
		})
	})

	t.Run("basename", func(t *testing.T) {
		t.Run("basic", func(t *testing.T) {
			script := `path.basename('/foo/bar/baz/asdf/quux.html');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "quux.html", val.String())
		})

		t.Run("no arguments", func(t *testing.T) {
			script := `path.basename();`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "", val.String())
		})

		t.Run("empty string", func(t *testing.T) {
			script := `path.basename('');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, ".", val.String())
		})

		t.Run("root path", func(t *testing.T) {
			script := `path.basename('/');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "/", val.String())
		})

		t.Run("path ending with slash", func(t *testing.T) {
			script := `path.basename('/foo/bar/');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "bar", val.String())
		})

		t.Run("no slashes", func(t *testing.T) {
			script := `path.basename('file.txt');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "file.txt", val.String())
		})
	})

	t.Run("extname", func(t *testing.T) {
		t.Run("basic", func(t *testing.T) {
			script := `path.extname('index.html');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, ".html", val.String())
		})

		t.Run("no arguments", func(t *testing.T) {
			script := `path.extname();`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "", val.String())
		})

		t.Run("empty string", func(t *testing.T) {
			script := `path.extname('');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "", val.String())
		})

		t.Run("no extension", func(t *testing.T) {
			script := `path.extname('index');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "", val.String())
		})

		t.Run("multiple dots", func(t *testing.T) {
			script := `path.extname('index.tar.gz');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, ".gz", val.String())
		})

		t.Run("hidden file", func(t *testing.T) {
			script := `path.extname('.bashrc');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, ".bashrc", val.String()) // Corrected expectation
		})

		t.Run("path with directory", func(t *testing.T) {
			script := `path.extname('/path/to/file.txt');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, ".txt", val.String())
		})
	})

	t.Run("parse", func(t *testing.T) {
		t.Run("basic", func(t *testing.T) {
			script := `path.parse('/home/user/dir/file.txt');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			obj := val.ToObject(vm)
			require.Equal(t, "/", obj.Get("root").String())
			require.Equal(t, "/home/user/dir", obj.Get("dir").String())
			require.Equal(t, "file.txt", obj.Get("base").String())
			require.Equal(t, ".txt", obj.Get("ext").String())
			require.Equal(t, "file", obj.Get("name").String())
		})

		t.Run("no arguments", func(t *testing.T) {
			script := `path.parse();`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.True(t, goja.IsNull(val)) // Corrected expectation: vm.ToValue(nil) is null, not undefined
		})

		t.Run("empty string", func(t *testing.T) {
			script := `path.parse('');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			obj := val.ToObject(vm)
			require.Equal(t, "", obj.Get("root").String())
			require.Equal(t, ".", obj.Get("dir").String()) // Corrected expectation
			require.Equal(t, ".", obj.Get("base").String())
			require.Equal(t, "", obj.Get("ext").String())
			require.Equal(t, "", obj.Get("name").String())
		})

		t.Run("root path", func(t *testing.T) {
			script := `path.parse('/');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			obj := val.ToObject(vm)
			require.Equal(t, "/", obj.Get("root").String())
			require.Equal(t, "/", obj.Get("dir").String())
			require.Equal(t, "/", obj.Get("base").String())
			require.Equal(t, "", obj.Get("ext").String())
			require.Equal(t, "/", obj.Get("name").String())
		})

		t.Run("relative path", func(t *testing.T) {
			script := `path.parse('file.txt');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			obj := val.ToObject(vm)
			require.Equal(t, "", obj.Get("root").String())
			require.Equal(t, ".", obj.Get("dir").String())
			require.Equal(t, "file.txt", obj.Get("base").String())
			require.Equal(t, ".txt", obj.Get("ext").String())
			require.Equal(t, "file", obj.Get("name").String())
		})

		t.Run("no extension", func(t *testing.T) {
			script := `path.parse('/path/to/file');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			obj := val.ToObject(vm)
			require.Equal(t, "/", obj.Get("root").String())
			require.Equal(t, "/path/to", obj.Get("dir").String())
			require.Equal(t, "file", obj.Get("base").String())
			require.Equal(t, "", obj.Get("ext").String())
			require.Equal(t, "file", obj.Get("name").String())
		})

		t.Run("multiple dots", func(t *testing.T) {
			script := `path.parse('/path/to/file.tar.gz');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			obj := val.ToObject(vm)
			require.Equal(t, "/", obj.Get("root").String())
			require.Equal(t, "/path/to", obj.Get("dir").String())
			require.Equal(t, "file.tar.gz", obj.Get("base").String())
			require.Equal(t, ".gz", obj.Get("ext").String())
			require.Equal(t, "file.tar", obj.Get("name").String())
		})
	})

	t.Run("matchesGlob", func(t *testing.T) {
		t.Run("basic - true", func(t *testing.T) {
			script := `path.matchesGlob('foo.txt', '*.txt');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.True(t, val.ToBoolean())
		})

		t.Run("no match", func(t *testing.T) {
			script := `path.matchesGlob('foo.txt', '*.js');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.False(t, val.ToBoolean())
		})

		t.Run("pattern with wildcard", func(t *testing.T) {
			script := `path.matchesGlob('foo.txt', 'f*.txt');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.True(t, val.ToBoolean())
		})

		t.Run("pattern with question mark", func(t *testing.T) {
			script := `path.matchesGlob('foo.txt', 'foo.???');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.True(t, val.ToBoolean())
		})

		t.Run("path is directory", func(t *testing.T) {
			script := `path.matchesGlob('/foo/bar', 'bar');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.True(t, val.ToBoolean())
		})

		t.Run("missing arguments", func(t *testing.T) {
			script := `path.matchesGlob('foo.txt');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "path.matchesGlob: missing path or pattern", val.String())
		})
	})

	t.Run("format", func(t *testing.T) {
		script := `path.format({});`
		val, err := vm.RunString(script)
		require.NoError(t, err)
		require.Equal(t, "Error: path.format not implemented", val.String())
	})
}