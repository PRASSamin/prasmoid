package command

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	root "github.com/PRASSamin/prasmoid/cmd"
	"github.com/PRASSamin/prasmoid/consts"
	"github.com/PRASSamin/prasmoid/tests"
)

func TestAddCommand(t *testing.T) {
	t.Run("successful command creation", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		commandName := "my-test-command"
		err := AddCommand(commandName)
		if err != nil {
			t.Fatalf("AddCommand() failed: %v", err)
		}

		commandFile := filepath.Join(root.ConfigRC.Commands.Dir, commandName+".js")
		if _, err := os.Stat(commandFile); os.IsNotExist(err) {
			t.Errorf("Command file was not created at %s", commandFile)
		}

		content, err := os.ReadFile(commandFile)
		if err != nil {
			t.Fatalf("Failed to read command file: %v", err)
		}

		absCommandFilePath, _ := filepath.Abs(commandFile)
		cwd, _ := os.Getwd()
		prasmoidDef := filepath.Join(cwd, "prasmoid.d.ts")
		relPath, _ := filepath.Rel(filepath.Dir(absCommandFilePath), prasmoidDef)

		expectedContent := fmt.Sprintf(consts.JS_COMMAND_TEMPLATE, relPath, commandName)
		if string(content) != expectedContent {
			t.Errorf("Generated file content does not match expected content.\nExpected:\n%s\n\nGot:\n%s", expectedContent, string(content))
		}
	})

	t.Run("command already exists", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		commandName := "existing-command"

		err := AddCommand(commandName)
		if err != nil {
			t.Fatalf("Initial AddCommand() failed: %v", err)
		}

		err = AddCommand(commandName)
		if err == nil {
			t.Error("AddCommand() should have failed for existing command, but it didn't")
		}
		if err != nil && !strings.Contains(err.Error(), "command already exists") {
			t.Errorf("Expected 'command already exists' error, but got: %v", err)
		}
	})

	t.Run("empty command name fails in non-interactive", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		err := AddCommand("")
		if err == nil {
			t.Fatal("AddCommand() with empty name should have failed")
		}
		if !strings.Contains(err.Error(), "error asking for command name") {
			t.Errorf("Expected prompt error, but got: %v", err)
		}
	})

	t.Run("invalid command name fails in non-interactive", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		err := AddCommand("invalid@name")
		if err == nil {
			t.Fatal("AddCommand() with invalid name should have failed")
		}
		if !strings.Contains(err.Error(), "error asking for command name") {
			t.Errorf("Expected prompt error, but got: %v", err)
		}
	})

	t.Run("failed to create commands directory", func(t *testing.T) {
		tmpDir, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		if err := os.Chmod(tmpDir, 0555); err != nil {
			t.Fatalf("Failed to make temp dir read-only: %v", err)
		}
		defer func() { _ = os.Chmod(tmpDir, 0755) }()

		err := AddCommand("test-command")
		if err == nil {
			t.Error("AddCommand() expected an error for failed directory creation, but got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "failed to create commands directory") {
			t.Errorf("Expected 'failed to create commands directory' error, got: %v", err)
		}
	})

	t.Run("failed to create command file", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		commandsDir := root.ConfigRC.Commands.Dir
		if err := os.MkdirAll(commandsDir, 0755); err != nil {
			t.Fatalf("Failed to create commands folder: %v", err)
		}
		if err := os.Chmod(commandsDir, 0555); err != nil {
			t.Fatalf("Failed to make commands folder read-only: %v", err)
		}
		defer func() { _ = os.Chmod(commandsDir, 0755) }()

		err := AddCommand("test-command")
		if err == nil {
			t.Error("AddCommand() expected an error for failed file creation, but got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "failed to create command file") {
			t.Errorf("Expected 'failed to create command file' error, got: %v", err)
		}
	})

	t.Run("cobra command execution", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		commandName := "cobra-command"
		_ = commandsAddCmd.Flags().Set("name", commandName)

		commandsAddCmd.Run(commandsAddCmd, []string{})

		commandFile := filepath.Join(root.ConfigRC.Commands.Dir, commandName+".js")
		if _, err := os.Stat(commandFile); os.IsNotExist(err) {
			t.Errorf("Command file was not created at %s by cobra command", commandFile)
		}

		_ = commandsAddCmd.Flags().Set("name", "")
	})

	t.Run("empty commands dir in config", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		originalDir := root.ConfigRC.Commands.Dir
		root.ConfigRC.Commands.Dir = ""
		defer func() { root.ConfigRC.Commands.Dir = originalDir }()

		commandName := "test-in-root"
		err := AddCommand(commandName)
		if err != nil {
			t.Fatalf("AddCommand() failed with empty commands dir: %v", err)
		}

		cwd, _ := os.Getwd()
		commandFile := filepath.Join(cwd, commandName+".js")
		if _, err := os.Stat(commandFile); os.IsNotExist(err) {
			t.Errorf("Command file was not created in current dir at %s", commandFile)
		}
	})
}
