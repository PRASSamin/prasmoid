package command

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	root "github.com/PRASSamin/prasmoid/cmd"
	initCmd "github.com/PRASSamin/prasmoid/cmd/init"
)

func TestRemoveCommand(t *testing.T) {
	t.Run("successfully removes a command with force flag", func(t *testing.T) {
		_, cleanup := initCmd.SetupTestProject(t)
		defer cleanup()

		_ = AddCommand("test-cmd")
		_ = commandsRemoveCmd.Flags().Set("name", "test-cmd (test-cmd.js)")
		_ = commandsRemoveCmd.Flags().Set("force", "true")

		commandsRemoveCmd.Run(commandsRemoveCmd, []string{})

		if _, err := os.Stat(filepath.Join(root.ConfigRC.Commands.Dir, "test-cmd.js")); !os.IsNotExist(err) {
			t.Errorf("Command file was not removed")
		}
	})
	
	t.Run("successfully removes a command with command name", func(t *testing.T) {
		_, cleanup := initCmd.SetupTestProject(t)
		defer cleanup()

		_ = AddCommand("test-cmd")
		_ = commandsRemoveCmd.Flags().Set("name", "test-cmd")
		_ = commandsRemoveCmd.Flags().Set("force", "true")

		commandsRemoveCmd.Run(commandsRemoveCmd, []string{})

		if _, err := os.Stat(filepath.Join(root.ConfigRC.Commands.Dir, "test-cmd.js")); !os.IsNotExist(err) {
			t.Errorf("Command file was not removed")
		}
	})
	
	t.Run("successfully removes a command with command file name", func(t *testing.T) {
		_, cleanup := initCmd.SetupTestProject(t)
		defer cleanup()

		_ = AddCommand("test-cmd")
		_ = commandsRemoveCmd.Flags().Set("name", "test-cmd.js")
		_ = commandsRemoveCmd.Flags().Set("force", "true")

		commandsRemoveCmd.Run(commandsRemoveCmd, []string{})

		if _, err := os.Stat(filepath.Join(root.ConfigRC.Commands.Dir, "test-cmd.js")); !os.IsNotExist(err) {
			t.Errorf("Command file was not removed")
		}
	})

	t.Run("fails to remove command", func(t *testing.T) {
		_, cleanup := initCmd.SetupTestProject(t)
		defer cleanup()
	
		_ = AddCommand("test-cmd")
		cmdFile := filepath.Join(root.ConfigRC.Commands.Dir, "test-cmd.js")
		
		// Make the directory read-only instead of the file
		dir := filepath.Dir(cmdFile)
		_ = os.Chmod(dir, 0555)
		defer func() { _ = os.Chmod(dir, 0755) }() 
		
		err := RemoveCommand("test-cmd (test-cmd.js)", true)
	
		if err == nil {
			t.Fatal("RemoveCommand() should have failed but did not")
		}
		if !strings.Contains(err.Error(), "error removing file") {
			t.Errorf("Expected 'error removing file' error, but got: %v", err)
		}
		
		// Verify the file still exists
		if _, err := os.Stat(cmdFile); os.IsNotExist(err) {
			t.Error("File was removed when it shouldn't have been")
		}
	})

	t.Run("fails to remove non-existent command with force flag", func(t *testing.T) {
		_, cleanup := initCmd.SetupTestProject(t)
		defer cleanup()

		err := RemoveCommand("non-existent (non-existent.js)", true)
		if err == nil {
			t.Fatal("RemoveCommand() should have failed but did not")
		}
		if !strings.Contains(err.Error(), "no commands found in the commands directory") {
			t.Errorf("Expected 'no commands found in the commands directory' error, but got: %v", err)
		}
	})

	t.Run("filepath.Walk fails for non-existent dir", func(t *testing.T) {
		_, cleanup := initCmd.SetupTestProject(t)
		defer cleanup()

		// Point to a non-existent directory
		root.ConfigRC.Commands.Dir = filepath.Join(root.ConfigRC.Commands.Dir, "nonexistent")

		err := RemoveCommand("test-cmd (test-cmd.js)", true)

		if err == nil {
			t.Fatal("RemoveCommand() should have failed but did not")
		}
		if !strings.Contains(err.Error(), "no commands found in the commands directory") {
			t.Errorf("Expected 'no commands found in the commands directory' error, but got: %v", err)
		}
	})

	t.Run("select prompt fails in non-interactive", func(t *testing.T) {
		_, cleanup := initCmd.SetupTestProject(t)
		defer cleanup()

		_ = AddCommand("test-cmd")

		err := RemoveCommand("", false)

		if err == nil {
			t.Fatal("RemoveCommand() should have failed but did not")
		}
		if !strings.Contains(err.Error(), "error asking for command name") {
			t.Errorf("Expected 'error asking for command name' error, but got: %v", err)
		}
	})

	t.Run("confirm prompt fails in non-interactive", func(t *testing.T) {
		_, cleanup := initCmd.SetupTestProject(t)
		defer cleanup()

		_ = AddCommand("test-cmd")
		err := RemoveCommand("test-cmd (test-cmd.js)", false)

		if err == nil {
			t.Fatal("RemoveCommand() should have failed but did not")
		}
		if !strings.Contains(err.Error(), "error asking for confirmation") {
			t.Errorf("Expected 'error asking for confirmation' error, but got: %v", err)
		}
	})
}
