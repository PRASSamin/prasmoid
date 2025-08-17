/*
Copyright 2025 PRAS
*/
package extendcli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/PRASSamin/prasmoid/internal/runtime"
	"github.com/PRASSamin/prasmoid/types"
	"github.com/spf13/cobra"
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
	rootCmd := &cobra.Command{}
	rootCmd.AddGroup(&cobra.Group{
		ID:    "custom",
		Title: "Custom Commands:",
	})
	validJS := `const prasmoid = require("prasmoid");
prasmoid.Command({
	run: (ctx) => {
		const plasmoidId = prasmoid.getMetadata("Id");
		if (!plasmoidId) {
			console.red(
			"Could not get Plasmoid ID. Are you in a valid project directory?"
			);
			return;
		}
	},
	short: "A brief description of your command.",
	long: "A longer description that spans multiple lines and likely contains examples\nand usage of using your command. For example:\n\nPlasmoid CLI is a CLI tool for KDE Plasmoid development.\nIt's a all-in-one tool for plasmoid development.",
	alias: ["tc", "testcmd"],
	flags: [
		{ name: "name", type: "string", shorthand: "n", value: "default", description: "a string flag" },
		{ name: "verbose", type: "bool", value: false, description: "a bool flag" },
	],
});`

	t.Run("valid js command", func(t *testing.T) {
		tmpfile := createTempJSFile(t, "valid_cmd", validJS)
		defer func() { _ = os.Remove(tmpfile) }()

		runtime.CommandStorage = runtime.CommandConfig{}
		_ = registerJSCommand(rootCmd, tmpfile)

		cmdName := strings.TrimSuffix(filepath.Base(tmpfile), ".js")
		cmdName = strings.ReplaceAll(cmdName, " ", "")

		found := false
		for _, cmd := range rootCmd.Commands() {
			if strings.HasPrefix(cmd.Use, cmdName) {
				found = true
				if !strings.Contains(cmd.Short,"A brief description of your command.") {
					t.Errorf("expected short description 'A brief description of your command.', got '%s'", cmd.Short)
				}
				if !strings.Contains(cmd.Long, "A longer description that spans multiple lines and likely contains examples") {
					t.Errorf("expected long description 'A longer description that spans multiple lines and likely contains examples...', got '%s'", cmd.Long)
				}
				if len(cmd.Aliases) != 2 {
					t.Errorf("expected 2 aliases, got %d", len(cmd.Aliases))
				}

				break
			}
		}

		if !found {
			t.Errorf("command '%s' not registered", cmdName)
		}
	})

	t.Run("non-existent file", func(t *testing.T) {
		err := registerJSCommand(rootCmd, "non-existent-file.js")
		if err == nil {
			t.Errorf("expected error for non-existent file")
		}
		if !strings.Contains(err.Error(), "failed to read JS script") {
			t.Errorf("expected error message for non-existent file, got: %s", err.Error())
		}
	})

	t.Run("invalid js syntax", func(t *testing.T) {
		tmpfile := createTempJSFile(t, "invalid_syntax", "const a =;")
		defer func() { _ = os.Remove(tmpfile) }()
		err := registerJSCommand(rootCmd, tmpfile)
		if err == nil {
			t.Errorf("expected error for invalid JS syntax")
		}
		if !strings.Contains(err.Error(), "error running script") {
			t.Errorf("expected error message for invalid JS syntax, got: %s", err.Error())
		}
	})

	t.Run("flag with no name", func(t *testing.T) {
		js := `prasmoid.Command({ 
			run: (ctx) => {
				const plasmoidId = prasmoid.getMetadata("Id");
				if (!plasmoidId) {
					console.red(
						"Could not get Plasmoid ID. Are you in a valid project directory?"
					);
					return;
				}
			},
	flags: [{ type: "string" }] })
		`
		tmpfile := createTempJSFile(t, "flag_no_name", js)
		defer func() { _ = os.Remove(tmpfile) }()
		err := registerJSCommand(rootCmd, tmpfile)
		if err == nil {
			t.Errorf("expected error for flag with no name")
		}
		if !strings.Contains(err.Error(), "flag name is required") {
			t.Errorf("expected error message for flag with no name, got: %s", err.Error())
		}
	})

	t.Run("flag with no type", func(t *testing.T) {
		js := `prasmoid.Command({run:()=>{}, flags: [{ name: "myflag" }] })`
		tmpfile := createTempJSFile(t, "flag_no_type", js)
		defer func() { _ = os.Remove(tmpfile) }()
		err := registerJSCommand(rootCmd, tmpfile)
		if err == nil {
			t.Errorf("expected error for flag with no type")
		}
		if !strings.Contains(err.Error(), "flag type is required") {
			t.Errorf("expected error message for flag with no type, got: %s", err.Error())
		}
	})

	t.Run("unsupported flag type", func(t *testing.T) {
		js := `prasmoid.Command({run:()=>{}, flags: [{ name: "myflag", type: "invalid" }] })`
		tmpfile := createTempJSFile(t, "flag_unsupported_type", js)
		defer func() { _ = os.Remove(tmpfile) }()
		err := registerJSCommand(rootCmd, tmpfile)
		if err == nil {
			t.Errorf("expected error for unsupported flag type")
		}
		if !strings.Contains(err.Error(), "unsupported flag type") {
			t.Errorf("expected error message for unsupported flag type, got: %s", err.Error())
		}
	})

	t.Run("command execution with args and flags", func(t *testing.T) {
		// Reset command storage for this test
		runtime.CommandStorage = runtime.CommandConfig{}

		jsContent := `
const prasmoid = require("prasmoid");
prasmoid.Command({
    run: (ctx) => {
        const args = ctx.Args();
        const flags = ctx.Flags();
        const name = flags.get("name");
        const verboseByGet = flags.get("verbose");
		const verboseByDot = flags.verbose;
        console.log("Args:", args.join(","));
        console.log("Flag 'name':", name);
        console.log("Flag 'verbose':", verboseByGet);
		console.log("Flag 'verbose':", verboseByDot);
    },
    short: "test run",
    flags: [
        { name: "name", type: "string", shorthand: "n", value: "default", description: "a string flag" },
        { name: "verbose", type: "bool", value: false, description: "a bool flag" },
    ],
});`
		tmpfile := createTempJSFile(t, "run_cmd", jsContent)
		defer func() { _ = os.Remove(tmpfile) }()

		err := registerJSCommand(rootCmd, tmpfile)
		if err != nil {
			t.Fatalf("registerJSCommand failed: %v", err)
		}

		cmdName := strings.TrimSuffix(filepath.Base(tmpfile), ".js")
		var registeredCmd *cobra.Command
		for _, cmd := range rootCmd.Commands() {
			if strings.HasPrefix(cmd.Use, cmdName) {
				registeredCmd = cmd
				break
			}
		}

		if registeredCmd == nil {
			t.Fatalf("command '%s' not registered", cmdName)
		}

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		_ = registeredCmd.Flags().Set("name", "test-name")
		_ = registeredCmd.Flags().Set("verbose", "true")
		// Execute the command
		registeredCmd.Run(registeredCmd, []string{"arg1", "arg2"})

		// Restore stdout and read captured output
		_ = w.Close()
		os.Stdout = oldStdout
		var buf strings.Builder
		_,_ = io.Copy(&buf, r)
		output := buf.String()

		// Assertions
		expectedArgs := "Args: arg1,arg2"
		if !strings.Contains(output, expectedArgs) {
			t.Errorf("expected output to contain '%s', got '%s'", expectedArgs, output)
		}

		expectedNameFlag := "Flag 'name': test-name"
		if !strings.Contains(output, expectedNameFlag) {
			t.Errorf("expected output to contain '%s', got '%s'", expectedNameFlag, output)
		}

		expectedVerboseFlag := "Flag 'verbose': true"
		if !strings.Contains(output, expectedVerboseFlag) {
			t.Errorf("expected output to contain '%s', got '%s'", expectedVerboseFlag, output)
		}
	})

	t.Run("js command run error", func(t *testing.T) {
		runtime.CommandStorage = runtime.CommandConfig{}

		jsContent := `
const prasmoid = require("prasmoid");
prasmoid.Command({
    run: (ctx) => {
        throw new Error("JS runtime error");
    },
    short: "test run error",
});`
		tmpfile := createTempJSFile(t, "run_error_cmd", jsContent)
		defer func() { _ = os.Remove(tmpfile) }()

		err := registerJSCommand(rootCmd, tmpfile)
		if err != nil {
			t.Fatalf("registerJSCommand failed: %v", err)
		}

		cmdName := strings.TrimSuffix(filepath.Base(tmpfile), ".js")
		var registeredCmd *cobra.Command
		for _, cmd := range rootCmd.Commands() {
			if strings.HasPrefix(cmd.Use, cmdName) {
				registeredCmd = cmd
				break
			}
		}

		if registeredCmd == nil {
			t.Fatalf("command '%s' not registered", cmdName)
		}

		// Capture stdout
		oldStderr := os.Stderr
		r, w, _ := os.Pipe()
		os.Stderr = w
			
		// Execute the command
		registeredCmd.Run(registeredCmd, []string{})
		
		// Restore stdout and read captured output
		_ = w.Close()
		os.Stdout = oldStderr
		var buf strings.Builder
		_,_ = io.Copy(&buf, r)
		output := buf.String()

		if !strings.Contains(output, "JS command error") || !strings.Contains(output, "JS runtime error") {
			t.Errorf("expected error message in stderr, got: %s", output)
		}
	})

	t.Run("flag variations", func(t *testing.T) {
		rootCmd := &cobra.Command{}
		js := `prasmoid.Command({
        run:()=>{},
        flags: [
            { name: "string-no-shorthand", type: "string", value: "default", description: "a string flag" },
            { name: "bool-with-shorthand", type: "bool", shorthand: "b", value: false, description: "a bool flag" },
        ],
    })`
		tmpfile := createTempJSFile(t, "flag_variations", js)
		defer func() { _ = os.Remove(tmpfile) }()

		runtime.CommandStorage = runtime.CommandConfig{}
		err := registerJSCommand(rootCmd, tmpfile)
		if err != nil {
			t.Fatalf("registerJSCommand failed: %v", err)
		}

		cmdName := strings.TrimSuffix(filepath.Base(tmpfile), ".js")
		var registeredCmd *cobra.Command
		for _, cmd := range rootCmd.Commands() {
			if strings.HasPrefix(cmd.Use, cmdName) {
				registeredCmd = cmd
				break
			}
		}
		if registeredCmd == nil {
			t.Fatalf("command '%s' not registered", cmdName)
		}

		stringFlag := registeredCmd.Flags().Lookup("string-no-shorthand")
		if stringFlag == nil {
			t.Error("expected 'string-no-shorthand' flag to be registered")
		} else if stringFlag.Shorthand != "" {
			t.Errorf("expected 'string-no-shorthand' to have no shorthand, got '%s'", stringFlag.Shorthand)
		}

		boolFlag := registeredCmd.Flags().Lookup("bool-with-shorthand")
		if boolFlag == nil {
			t.Error("expected 'bool-with-shorthand' flag to be registered")
		} else if boolFlag.Shorthand != "b" {
			t.Errorf("expected 'bool-with-shorthand' to have shorthand 'b', got '%s'", boolFlag.Shorthand)
		}
	})
}

func TestDiscoverAndRegisterCustomCommands(t *testing.T) {
	rootCmd := &cobra.Command{Use: "root"}

	t.Run("discover and register", func(t *testing.T) {
		commandsDir := t.TempDir()
		jsCmd1 := `if (typeof prasmoid !== 'undefined') { prasmoid.command = { Short: "cmd1" } }`
		jsCmd2 := `if (typeof prasmoid !== 'undefined') { prasmoid.command = { Short: "cmd2" } }`
		ignoredFile := `if (typeof prasmoid !== 'undefined') { prasmoid.command = { Short: "ignored" } }`

		_ = os.WriteFile(filepath.Join(commandsDir, "cmd1.js"), []byte(jsCmd1), 0644)
		_ = os.WriteFile(filepath.Join(commandsDir, "cmd2.js"), []byte(jsCmd2), 0644)
		_ = os.WriteFile(filepath.Join(commandsDir, "ignored.js"), []byte(ignoredFile), 0644)
		_ = os.Mkdir(filepath.Join(commandsDir, "subdir"), 0755)

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
		rootCmd := &cobra.Command{Use: "root"}
		commandsDir := t.TempDir()
		// On Unix-like systems, removing read permissions should cause ReadDir to fail.
		if err := os.Chmod(commandsDir, 0300); err != nil {
			t.Skipf("Skipping test: could not chmod: %v", err)
		}
		defer func() { _ = os.Chmod(commandsDir, 0755) }() // clean up

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
		_,_ = io.Copy(&buf, r)
		output := buf.String()

		if len(rootCmd.Commands()) != 0 {
			t.Errorf("expected 0 commands when dir is unreadable, got %d", len(rootCmd.Commands()))
		}
		if !strings.Contains(output, "Error reading commands directory") {
			t.Errorf("expected error message about reading directory, got: %s", output)
		}
	})
}

func createTempJSFile(t *testing.T, name, content string) string {
	t.Helper()
	tmpfile, err := os.CreateTemp("", fmt.Sprintf("%s.*.js", name))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}
	return tmpfile.Name()
}
