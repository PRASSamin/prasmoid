package runtime

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConsoleModule(t *testing.T) {
	vm := NewRuntime()
	_, err := vm.RunString(`const console = require('console');`)
	if err != nil {
		t.Fatalf("Failed to declare console: %v", err)
	}

	// Helper to capture stdout
	captureOutput := func(f func()) string {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		f()

		_ = w.Close()
		os.Stdout = old
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		return strings.TrimSpace(buf.String())
	}

	t.Run("plain loggers", func(t *testing.T) {
		loggers := []string{"log", "warn", "error", "debug", "info"}
		for _, logger := range loggers {
			t.Run(logger, func(t *testing.T) {
				script := `console.` + logger + `('hello', 'world', 123);`
				output := captureOutput(func() {
					_, err := vm.RunString(script)
					if err != nil {
						t.Fatalf("vm.RunString() failed: %v", err)
					}
				})
				expected := "hello world 123"
				if output != expected {
					t.Errorf("Expected '%s', but got '%s'", expected, output)
				}
			})
		}
	})

	t.Run("colored loggers", func(t *testing.T) {
		colors := []string{"red", "green", "yellow"}
		for _, c := range colors {
			t.Run(c, func(t *testing.T) {
				script := `console.` + c + `('hello', 'world');`
				output := captureOutput(func() {
					_, err := vm.RunString(script)
					if err != nil {
						t.Fatalf("vm.RunString() failed: %v", err)
					}
				})
				// Just check if it contains the text, not the color codes
				if !strings.Contains(output, "hello world") {
					t.Errorf("Expected output to contain 'hello world', but got '%s'", output)
				}
			})
		}
	})

	t.Run("console.color", func(t *testing.T) {
		t.Run("with valid color", func(t *testing.T) {
			script := `console.color('hello', 'world', 'red');`
			output := captureOutput(func() {
				_, err := vm.RunString(script)
				if err != nil {
					t.Fatalf("vm.RunString() failed: %v", err)
				}
			})
			if !strings.Contains(output, "hello world") {
				t.Errorf("Expected output to contain 'hello world', but got '%s'", output)
			}
		})

		t.Run("with default color", func(t *testing.T) {
			script := `console.color('hello world');`
			output := captureOutput(func() {
				_, err := vm.RunString(script)
				if err != nil {
					t.Fatalf("vm.RunString() failed: %v", err)
				}
			})
			if !strings.Contains(output, "hello world") {
				t.Errorf("Expected output to contain 'hello world', but got '%s'", output)
			}
		})

		t.Run("with invalid color", func(t *testing.T) {
			script := `console.color('hello', 'world', 'invalidcolor');`
			output := captureOutput(func() {
				_, err := vm.RunString(script)
				if err != nil {
					t.Fatalf("vm.RunString() failed: %v", err)
				}
			})
			if !strings.Contains(output, "hello world invalidcolor") {
				t.Errorf("Expected output to contain 'hello world invalidcolor', but got '%s'", output)
			}
		})

		t.Run("with no arguments", func(t *testing.T) {
			script := `console.color();`
			output := captureOutput(func() {
				_, err := vm.RunString(script)
				if err != nil {
					t.Fatalf("vm.RunString() failed: %v", err)
				}
			})
			if output != "" {
				t.Errorf("Expected empty output, but got '%s'", output)
			}
		})

		t.Run("with various colors", func(t *testing.T) {
			colors := []string{"blue", "magenta", "cyan", "black", "green", "yellow"}
			for _, c := range colors {
				t.Run(c, func(t *testing.T) {
					script := `console.color('test', '` + c + `');`
					output := captureOutput(func() {
						_, err := vm.RunString(script)
						if err != nil {
							t.Fatalf("vm.RunString() failed: %v", err)
						}
					})
					if !strings.Contains(output, "test") {
						t.Errorf("Expected output to contain 'test', but got '%s'", output)
					}
				})
			}
		})
	})
}

func TestStringifyJS(t *testing.T) {
	vm := NewRuntime()
	_, err := vm.RunString(`const console = require('console');`)
	if err != nil {
		t.Fatalf("Failed to declare console: %v", err)
	}

	// Helper to capture stdout
	captureOutput := func(f func()) string {
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		f()

		_ = w.Close()
		os.Stdout = old
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		return strings.TrimSpace(buf.String())
	}
	
	t.Run("stringify one data type array", func(t *testing.T) {
		script := `console.log(["p", "r", "a", "s"]);`
		output := captureOutput(func() {
			_, err := vm.RunString(script)
			if err != nil {
				t.Fatalf("vm.RunString() failed: %v", err)
			}
		})
		require.Contains(t, output, `["p", "r", "a", "s"]`)
	})
	
	t.Run("stringify multiple data type array", func(t *testing.T) {
		script := `console.log(["d", 2, 1, true]);`
		output := captureOutput(func() {
			_, err := vm.RunString(script)
			if err != nil {
				t.Fatalf("vm.RunString() failed: %v", err)
			}
		})
		require.Contains(t, output, `["d", 2, 1, true]`)
	})
	
	t.Run("stringify empty array", func(t *testing.T) {
		script := `console.log([]);`
		output := captureOutput(func() {
			_, err := vm.RunString(script)
			if err != nil {
				t.Fatalf("vm.RunString() failed: %v", err)
			}
		})
		require.Contains(t, output, `[]`)
	})
	
	t.Run("stringify empty object", func(t *testing.T) {
		script := `console.log({});`
		output := captureOutput(func() {
			_, err := vm.RunString(script)
			if err != nil {
				t.Fatalf("vm.RunString() failed: %v", err)
			}
		})
		require.Contains(t, output, `{  }`)
	})
	
	t.Run("stringify object", func(t *testing.T) {
		script := `console.log({a: 1, b: "c"});`
		output := captureOutput(func() {
			_, err := vm.RunString(script)
			if err != nil {
				t.Fatalf("vm.RunString() failed: %v", err)
			}
		})
		require.Contains(t, output, `{ a: 1, b: "c" }`)
	})
	
	t.Run("stringify functions", func(t *testing.T) {
		script := `console.log(function() { return 1; });`
		output := captureOutput(func() {
			_, err := vm.RunString(script)
			if err != nil {
				t.Fatalf("vm.RunString() failed: %v", err)
			}
		})
		require.Contains(t, output, "[Function]")
	})
	
	t.Run("stringify null", func(t *testing.T) {
		script := `console.log(null);`
		output := captureOutput(func() {
			_, err := vm.RunString(script)
			if err != nil {
				t.Fatalf("vm.RunString() failed: %v", err)
			}
		})
		require.Contains(t, output, "null")
	})
}
