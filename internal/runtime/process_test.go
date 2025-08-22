package runtime

import (
	"fmt"
	"os"
	"testing"

	"path/filepath"

	"github.com/dop251/goja"
	"github.com/stretchr/testify/require"
)

func TestProcessModule(t *testing.T) {
	vm := NewRuntime()
	_, err := vm.RunString(`const process = require('process');`)
	require.NoError(t, err)

	t.Run("cwd", func(t *testing.T) {
		script := `process.cwd();`
		val, err := vm.RunString(script)
		require.NoError(t, err)
		wd, _ := os.Getwd()
		require.Equal(t, wd, val.String())
	})

	t.Run("chdir", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "chdir-test-")
		require.NoError(t, err)
		defer func() {
			require.NoError(t, os.RemoveAll(tmpDir))
		}()

		script := `process.chdir('` + tmpDir + `');`
		_, err = vm.RunString(script)
		require.NoError(t, err)
		wd, _ := os.Getwd()
		require.Equal(t, tmpDir, wd)

		t.Run("no arguments", func(t *testing.T) {
			script := `process.chdir();`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "chdir: path is required", val.String())
		})

		t.Run("invalid path", func(t *testing.T) {
			script := `process.chdir('/non/existent/path');`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Contains(t, val.String(), "chdir error:")
		})
	})

	t.Run("uptime", func(t *testing.T) {
		script := `process.uptime();`
		val, err := vm.RunString(script)
		require.NoError(t, err)
		require.Greater(t, val.ToFloat(), float64(0))
	})

	t.Run("memoryUsage", func(t *testing.T) {
		script := `process.memoryUsage();`
		val, err := vm.RunString(script)
		require.NoError(t, err)
		obj := val.ToObject(vm)
		require.Greater(t, obj.Get("rss").ToInteger(), int64(0))
		require.Greater(t, obj.Get("heapTotal").ToInteger(), int64(0))
		require.Greater(t, obj.Get("heapUsed").ToInteger(), int64(0))
		require.Greater(t, obj.Get("external").ToInteger(), int64(0))
	})

	t.Run("kill", func(t *testing.T) {
		t.Run("no arguments", func(t *testing.T) {
			script := `process.kill();`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "kill: pid required", val.String())
		})

		t.Run("invalid pid", func(t *testing.T) {
			script := `process.kill(-1);`
			val, err := vm.RunString(script)
			require.NoError(t, err)
			require.Contains(t, val.String(), "kill error:")
		})
	})

	t.Run("getuid", func(t *testing.T) {
		script := `process.getuid();`
		val, err := vm.RunString(script)
		require.NoError(t, err)
		require.Equal(t, int64(os.Getuid()), val.ToInteger())
	})

	t.Run("getgid", func(t *testing.T) {
		script := `process.getgid();`
		val, err := vm.RunString(script)
		require.NoError(t, err)
		require.Equal(t, int64(os.Getgid()), val.ToInteger())
	})

	t.Run("geteuid", func(t *testing.T) {
		script := `process.geteuid();`
		val, err := vm.RunString(script)
		require.NoError(t, err)
		require.Equal(t, int64(os.Geteuid()), val.ToInteger())
	})

	t.Run("getegid", func(t *testing.T) {
		script := `process.getegid();`
		val, err := vm.RunString(script)
		require.NoError(t, err)
		require.Equal(t, int64(os.Getegid()), val.ToInteger())
	})

	t.Run("env", func(t *testing.T) {
		script := `process.env.HOME;`
		val, err := vm.RunString(script)
		require.NoError(t, err)
		require.Equal(t, os.Getenv("HOME"), val.String())

		t.Run("load from .env files", func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "env-test-")
			require.NoError(t, err)
			defer func() {
				require.NoError(t, os.RemoveAll(tmpDir))
			}()

			require.NoError(t, os.Chdir(tmpDir))

			// Create a dummy .env file
			envContent := `
PRASMOID_TEST_VAR_1=value1
# This is a comment
PRASMOID_TEST_VAR_2="value2 with spaces"
ANOTHER_VAR=should_be_ignored
PRASMOID_TEST_VAR_3=value3_no_quotes
`
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".env"), []byte(envContent), 0644))

			// Re-initialize the runtime to pick up new envs
			vm = NewRuntime()
			_, err = vm.RunString(`const process = require('process');`)
			require.NoError(t, err)

			script := `process.env.PRASMOID_TEST_VAR_1;`
			val, err = vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "value1", val.String())

			script = `process.env.PRASMOID_TEST_VAR_2;`
			val, err = vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "value2 with spaces", val.String())

			script = `process.env.PRASMOID_TEST_VAR_3;`
			val, err = vm.RunString(script)
			require.NoError(t, err)
			require.Equal(t, "value3_no_quotes", val.String())

			script = `process.env.ANOTHER_VAR;`
			val, err = vm.RunString(script)
			require.NoError(t, err)
			require.True(t, goja.IsUndefined(val))
		})

		t.Run("empty .env file", func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "env-empty-test-")
			require.NoError(t, err)
			defer func() {
				require.NoError(t, os.RemoveAll(tmpDir))
			}()

			require.NoError(t, os.Chdir(tmpDir))

			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".env"), []byte(``), 0644))

			vm = NewRuntime()
			_, err = vm.RunString(`const process = require('process');`)
			require.NoError(t, err)

			script := `process.env.PRASMOID_EMPTY_VAR;`
			val, err = vm.RunString(script)
			require.NoError(t, err)
			require.True(t, goja.IsUndefined(val))
		})

		t.Run(".env file with no PRASMOID_ prefix", func(t *testing.T) {
			tmpDir, err := os.MkdirTemp("", "env-no-prefix-test-")
			require.NoError(t, err)
			defer func() {
				require.NoError(t, os.RemoveAll(tmpDir))
			}()

			require.NoError(t, os.Chdir(tmpDir))

			envContent := `
MY_VAR=test
ANOTHER_VAR=test2
`
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, ".env"), []byte(envContent), 0644))

			vm = NewRuntime()
			_, err = vm.RunString(`const process = require('process');`)
			require.NoError(t, err)

			script := `process.env.MY_VAR;`
			val, err = vm.RunString(script)
			require.NoError(t, err)
			require.True(t, goja.IsUndefined(val))
		})
	})

		t.Run("not implemented functions", func(t *testing.T) {
			notImplementedFuncs := []string{
				"nextTick", "binding", "dlopen", "getActiveResourcesInfo", "reallyExit", "loadEnvFile",
				"cpuUsage", "resourceUsage", "constrainedMemory", "availableMemory", "execve",
				"ref", "unref", "hrtime", "openStdin", "getgroups", "assert",
				"setUncaughtExceptionCaptureCallback", "hasUncaughtExceptionCaptureCallback",
				"emitWarning", "setSourceMapsEnabled", "getBuiltinModule", "abort",
				"initgroups", "setgroups", "setegid", "seteuid", "setgid", "setuid",
			}

			for _, name := range notImplementedFuncs {
				t.Run(name, func(t *testing.T) {
					script := fmt.Sprintf(`process.%s();`, name)
					val, err := vm.RunString(script)
					require.NoError(t, err)
					expected := fmt.Sprintf("process.%s is not implemented in this runtime", name)
					require.Equal(t, expected, val.String())
				})
			}
		})
	}