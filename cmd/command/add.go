/*
Copyright 2025 PRAS
*/
package command

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/PRASSamin/prasmoid/consts"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	root "github.com/PRASSamin/prasmoid/cmd"
)

var commandTemplates = map[string]string{
	"js": consts.JS_COMMAND_TEMPLATE,
}


func init() {
	commandsAddCmd.Flags().StringP("name", "n", "", "Command name")
	commandsCmd.AddCommand(commandsAddCmd)
}

var commandsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a custom command",
	Long:  "Add a custom command to the project.",
	Run: func(cmd *cobra.Command, args []string) {
		commandName, _ := cmd.Flags().GetString("name")
		_ = AddCommand(commandName)
	},
}

func AddCommand(commandName string) error {
	invalidChars := regexp.MustCompile(`[\\/:*?"<>|\s@]`)

	// Ask for command name
	if strings.TrimSpace(commandName) == "" || invalidChars.MatchString(commandName) {
		namePrompt := &survey.Input{
			Message: "Command name:",
		}
		if err := survey.AskOne(namePrompt, &commandName, survey.WithValidator(func(ans interface{}) error {
			name := ans.(string)
			if name == "" {
				return fmt.Errorf("command name cannot be empty")
			}

			if invalidChars.MatchString(name) {
				return fmt.Errorf("invalid characters in command name")
			}

			baseName := filepath.Join(root.ConfigRC.Commands.Dir, name)
			if _, err := os.Stat(baseName + ".js"); err == nil {
				return fmt.Errorf("command already exists")
			}
			return nil
		})); err != nil {
			return fmt.Errorf("error asking for command name: %v", err)
		}
	}

	if !invalidChars.MatchString(commandName) {
		baseName := filepath.Join(root.ConfigRC.Commands.Dir, commandName)
		if _, err := os.Stat(baseName + ".js"); err == nil {
			color.Red("Command name already exists")
			return fmt.Errorf("command already exists")
		}
	}

	
	if (strings.TrimSpace(root.ConfigRC.Commands.Dir) != "") {
		// Ensure the commands directory exists
		if err := os.MkdirAll(root.ConfigRC.Commands.Dir, 0755); err != nil {
			color.Red("Failed to create commands directory: %v", err)
			return fmt.Errorf("failed to create commands directory: %v", err)
		}
	}

	commandFile := commandName + ".js"
	filePath := filepath.Join(root.ConfigRC.Commands.Dir, commandFile)

	// Create the new command file
	file, err := os.Create(filePath)
	if err != nil {
		color.Red("Error creating file: %v", err)
		return fmt.Errorf("failed to create command file: %v", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			color.Red("Error closing file: %v", err)
		}
	}()

	// Absolute path to command file
	absCommandFilePath, _ := filepath.Abs(filePath)
	
	cwd, _ := os.Getwd()
	rootDir, _ := filepath.Abs(cwd)
	prasmoidDef := filepath.Join(rootDir, "prasmoid.d.ts")

	// Calculate relative path from command file to prasmoid.d.ts
	relPath, err := filepath.Rel(filepath.Dir(absCommandFilePath), prasmoidDef)
	if err != nil {
		color.Red("Error calculating relative path: %v", err)
		return fmt.Errorf("error calculating relative path: %v", err)
	}
	
	templateFilled := fmt.Sprintf(commandTemplates["js"], relPath, commandName)

	_, err = file.WriteString(templateFilled)
	if err != nil {
		color.Red("Error writing to file: %v", err)
		return fmt.Errorf("error writing to file: %v", err)
	}

	color.Green("Successfully created %s command at %s", commandName, color.BlueString(filePath))
	return nil
}