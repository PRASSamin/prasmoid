/*
Copyright 2025 PRAS
*/
package extendcli

import (
	"io"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/PRASSamin/prasmoid/types"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsIgnored(t *testing.T) {
	root := t.TempDir()
	tests := []struct {
		name       string
		filename   string
		ignoreList []string
		want       bool
	}{
		{"no ignore list", "command.js", []string{}, false},
		{"exact match", "command.js", []string{"command.js"}, true},
		{"no match", "command.js", []string{"another.js"}, false},
		{"glob match star", "command.js", []string{"*.js"}, true},
		{"glob no match", "command.ts", []string{"*.js"}, false},
		{"glob match question mark", "command.js", []string{"command.j?"}, true},
		{"glob match brackets", "command.js", []string{"command.[jt]s"}, true},
		{"mixed list match", "command.js", []string{"another.js", "*.js"}, true},
		{"mixed list no match", "command.ts", []string{"another.js", "*.js"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isIgnored(tt.filename, tt.ignoreList, root); got != tt.want {
				t.Errorf("isIgnored() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRegisterJSCommand(t *testing.T) {
	t.Run("valid js command", func(t *testing.T) {
		// Arrange
		t.Cleanup(func() {
			osReadFile = os.ReadFile
		})
		validJS := `const prasmoid = require("prasmoid");
		prasmoid.Command({
			run: (ctx) => {},
			short: "A brief description of your command.",
		});`
		osReadFile = func(name string) ([]byte, error) {
			return []byte(validJS), nil
		}
		rootCmd := &cobra.Command{}

		// Act
		err := registerJSCommand(rootCmd, "test.js")

		// Assert
		assert.NoError(t, err)
		assert.Equal(t, 1, len(rootCmd.Commands()))
		assert.Equal(t, "test", rootCmd.Commands()[0].Use)
		assert.Equal(t, "A brief description of your command.", rootCmd.Commands()[0].Short)
	})

	t.Run("non-existent file", func(t *testing.T) {
		// Arrange
		t.Cleanup(func() { osReadFile = os.ReadFile })
		osReadFile = func(name string) ([]byte, error) {
			return nil, os.ErrNotExist
		}
		rootCmd := &cobra.Command{}

		// Act
		err := registerJSCommand(rootCmd, "non-existent-file.js")

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read JS script")
	})

	t.Run("invalid js syntax", func(t *testing.T) {
		// Arrange
		t.Cleanup(func() { osReadFile = os.ReadFile })
		osReadFile = func(name string) ([]byte, error) {
			return []byte("const a =;"), nil
		}
		rootCmd := &cobra.Command{}

		// Act
		err := registerJSCommand(rootCmd, "invalid.js")

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error running script")
	})

	t.Run("flag with no name", func(t *testing.T) {
		// Arrange
		t.Cleanup(func() { osReadFile = os.ReadFile })
		js := `prasmoid.Command({run:()=>{}, flags: [{ type: "string" }] })`
		osReadFile = func(name string) ([]byte, error) {
			return []byte(js), nil
		}
		rootCmd := &cobra.Command{}

		// Act
		err := registerJSCommand(rootCmd, "test.js")

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "flag name is required")
	})

	t.Run("flag with no type", func(t *testing.T) {
		// Arrange
		t.Cleanup(func() { osReadFile = os.ReadFile })
		js := `prasmoid.Command({run:()=>{}, flags: [{ name: "myflag" }] })`
		osReadFile = func(name string) ([]byte, error) {
			return []byte(js), nil
		}
		rootCmd := &cobra.Command{}

		// Act
		err := registerJSCommand(rootCmd, "test.js")

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "flag type is required")
	})

	t.Run("unsupported flag type", func(t *testing.T) {
		// Arrange
		t.Cleanup(func() { osReadFile = os.ReadFile })
		js := `prasmoid.Command({run:()=>{}, flags: [{ name: "myflag", type: "invalid" }] })`
		osReadFile = func(name string) ([]byte, error) {
			return []byte(js), nil
		}
		rootCmd := &cobra.Command{}

		// Act
		err := registerJSCommand(rootCmd, "test.js")

		// Assert
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported flag type")
	})

	t.Run("command execution with args and flags", func(t *testing.T) {
		// Arrange
		t.Cleanup(func() { osReadFile = os.ReadFile })
		jsContent := `
		const prasmoid = require("prasmoid");
		prasmoid.Command({
		    run: (ctx) => {
		        const args = ctx.Args();
		        const flags = ctx.Flags();
		        const name = flags.get("name");
		        const verbose = flags.get("verbose");
				const directName = flags.name;
		        console.log("Args:", args.join(","));
		        console.log("Flag 'name':", name);
		        console.log("Flag 'verbose':", verbose);
				console.log("Flag direct 'name':", directName);
		    },
		    short: "test run",
		    flags: [
		        { name: "name", type: "string", shorthand: "n", value: "default", description: "a string flag" },
		        { name: "verbose", type: "bool", value: false, description: "a bool flag" },
		    ],
		});`
		osReadFile = func(name string) ([]byte, error) {
			return []byte(jsContent), nil
		}
		rootCmd := &cobra.Command{Use: "root"}
		rootCmd.AddGroup(&cobra.Group{ID: "custom", Title: "Custom Commands"})
		rootCmd.SetOut(io.Discard) // Prevent cobra from printing
		rootCmd.SetErr(io.Discard)

		err := registerJSCommand(rootCmd, "test.js")
		require.NoError(t, err)

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Act
		rootCmd.SetArgs([]string{"test", "arg1", "arg2", "--name", "test-name", "--verbose"})
		executeErr := rootCmd.Execute()

		_ = w.Close()
		os.Stdout = oldStdout
		var buf strings.Builder
		_, _ = io.Copy(&buf, r)
		output := buf.String()

		// Assert
		require.NoError(t, executeErr)
		assert.Contains(t, output, "Args: arg1,arg2")
		assert.Contains(t, output, "Flag 'name': test-name")
		assert.Contains(t, output, "Flag 'verbose': true")
		assert.Contains(t, output, "Flag direct 'name': test-name")
	})

	t.Run("js command run error", func(t *testing.T) {
		// Arrange
		t.Cleanup(func() { osReadFile = os.ReadFile })
		jsContent := `
		const prasmoid = require("prasmoid");
		prasmoid.Command({
		    run: (ctx) => {
		        throw new Error("JS runtime error");
		    },
		    short: "test run error",
		});`
		osReadFile = func(name string) ([]byte, error) {
			return []byte(jsContent), nil
		}
		rootCmd := &cobra.Command{Use: "root"}
		rootCmd.AddGroup(&cobra.Group{ID: "custom", Title: "Custom Commands"})
		rootCmd.SetOut(io.Discard)
		rootCmd.SetErr(io.Discard)

		err := registerJSCommand(rootCmd, "test.js")
		require.NoError(t, err)

		// Capture stderr
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		// Act
		rootCmd.SetArgs([]string{"test"})
		executeErr := rootCmd.Execute()

		_ = w.Close()
		os.Stderr = oldStderr
		var buf strings.Builder
		_, _ = io.Copy(&buf, r)
		output := buf.String()

		// Assert
		require.NoError(t, executeErr)
		assert.Contains(t, output, "JS command error")
		assert.Contains(t, output, "JS runtime error")
	})

	t.Run("flag variations", func(t *testing.T) {
		// Arrange
		t.Cleanup(func() { osReadFile = os.ReadFile })
		js := `prasmoid.Command({
        run:()=>{},
        flags: [
            { name: "string-no-shorthand", type: "string", value: "default", description: "a string flag" },
            { name: "bool-with-shorthand", type: "bool", shorthand: "b", value: false, description: "a bool flag" },
        ],
    })`
		osReadFile = func(name string) ([]byte, error) {
			return []byte(js), nil
		}
		rootCmd := &cobra.Command{}

		// Act
		err := registerJSCommand(rootCmd, "test.js")

		// Assert
		assert.NoError(t, err)
	})
}

func setup(t *testing.T) {
	t.Helper()
	// Save original functions
	originalOsStat := osStat
	originalOsReadDir := osReadDir
	originalOsReadFile := osReadFile
	originalOsMkdirAll := osMkdirAll
	originalOsWriteFile := osWriteFile
	originalOsRemoveAll := osRemoveAll
	originalOsCreateTemp := osCreateTemp
	originalFilepathJoin := filepathJoin
	originalDoublestarPathMatch := doublestarPathMatch

	// Restore original functions after test
	t.Cleanup(func() {
		osStat = originalOsStat
		osReadDir = originalOsReadDir
		osReadFile = originalOsReadFile
		osMkdirAll = originalOsMkdirAll
		osWriteFile = originalOsWriteFile
		osRemoveAll = originalOsRemoveAll
		osCreateTemp = originalOsCreateTemp
		filepathJoin = originalFilepathJoin
		doublestarPathMatch = originalDoublestarPathMatch
	})
}

func TestDiscoverAndRegisterCustomCommands(t *testing.T) {
	rootCmd := &cobra.Command{Use: "root"}

	t.Run("discover and register", func(t *testing.T) {
		setup(t) // Call setup at the beginning of each test case
		commandsDir := t.TempDir()
		jsCmd1 := `if (typeof prasmoid !== 'undefined') { prasmoid.command = { Short: "cmd1" } }`
		jsCmd2 := `if (typeof prasmoid !== 'undefined') { prasmoid.command = { Short: "cmd2" } }`
		ignoredFile := `if (typeof prasmoid !== 'undefined') { prasmoid.command = { Short: "ignored" } }`

		_ = osWriteFile(filepathJoin(commandsDir, "cmd1.js"), []byte(jsCmd1), 0644)
		_ = osWriteFile(filepathJoin(commandsDir, "cmd2.js"), []byte(jsCmd2), 0644)
		_ = osWriteFile(filepathJoin(commandsDir, "ignored.js"), []byte(ignoredFile), 0644)
		_ = osMkdirAll(filepathJoin(commandsDir, "subdir"), 0755)

		config := types.Config{
			Commands: types.ConfigCommands{
				Dir:    commandsDir,
				Ignore: []string{"ignored.js"},
			},
		}

		DiscoverAndRegisterCustomCommands(rootCmd, config)

		if len(rootCmd.Commands()) != 2 {
			t.Errorf("expected 2 commands to be registered, got %d", len(rootCmd.Commands()))
		}

		var cmd1, cmd2 bool
		for _, cmd := range rootCmd.Commands() {
			if strings.HasPrefix(cmd.Use, "cmd1") {
				cmd1 = true
			}
			if strings.HasPrefix(cmd.Use, "cmd2") {
				cmd2 = true
			}
		}
		if !cmd1 || !cmd2 {
			t.Errorf("expected cmd1 and cmd2 to be registered")
		}
	})

	t.Run("commands dir does not exist", func(t *testing.T) {
		setup(t) // Call setup at the beginning of each test case
		rootCmd := &cobra.Command{Use: "root"}
		config := types.Config{
			Commands: types.ConfigCommands{
				Dir: "/non/existent/dir",
			},
		}
		DiscoverAndRegisterCustomCommands(rootCmd, config)
		if len(rootCmd.Commands()) != 0 {
			t.Errorf("expected 0 commands for non-existent dir, got %d", len(rootCmd.Commands()))
		}
	})

	t.Run("cannot read commands dir", func(t *testing.T) {
		setup(t) // Call setup at the beginning of each test case
		rootCmd := &cobra.Command{Use: "root"}
		commandsDir := t.TempDir()
		// On Unix-like systems, removing read permissions should cause ReadDir to fail.
		// Mock osReadDir to simulate permission denied error
		osReadDir = func(name string) ([]fs.DirEntry, error) {
			return nil, os.ErrPermission
		}

		config := types.Config{
			Commands: types.ConfigCommands{
				Dir: commandsDir,
			},
		}

		// Capture stderr to check for the error message
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w

		DiscoverAndRegisterCustomCommands(rootCmd, config)

		_ = w.Close()
		os.Stderr = oldStderr
		var buf strings.Builder
		_, _ = io.Copy(&buf, r)
		output := buf.String()

		if len(rootCmd.Commands()) != 0 {
			t.Errorf("expected 0 commands when dir is unreadable, got %d", len(rootCmd.Commands()))
		}
		if !strings.Contains(output, "Error reading commands directory") {
			t.Errorf("expected error message about reading directory, got: %s", output)
		}
	})
}
