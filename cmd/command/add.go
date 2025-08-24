/*
Copyright 2025 PRAS
*/
package command

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	root "github.com/PRASSamin/prasmoid/cmd"
	"github.com/PRASSamin/prasmoid/consts"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var commandTemplates = map[string]string{
	"js": consts.JS_COMMAND_TEMPLATE,
}

func commandNameValidator(ans interface{}) error {
	name := ans.(string)
	if name == "" {
		return fmt.Errorf("command name cannot be empty")
	}

	if invalidChars.MatchString(name) {
		return fmt.Errorf("invalid characters in command name")
	}

	baseName := filepath.Join(root.ConfigRC.Commands.Dir, name)
	if _, err := osStat(baseName + ".js"); err == nil {
		return fmt.Errorf("command already exists")
	}
	return nil
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
		AddCommand(commandName)
	},
}

func AddCommand(commandName string) {
	// Ask for command name
	if strings.TrimSpace(commandName) == "" || invalidChars.MatchString(commandName) {
		namePrompt := &survey.Input{
			Message: "Command name:",
		}
		if err := surveyAskOne(namePrompt, &commandName, survey.WithValidator(commandNameValidator)); err != nil {
			fmt.Println(color.RedString("Error asking for command name: %v", err))
			return
		}
	}

	if !invalidChars.MatchString(commandName) {
		baseName := filepath.Join(root.ConfigRC.Commands.Dir, commandName)
		if _, err := osStat(baseName + ".js"); err == nil {
			fmt.Println(color.RedString("Command already exists"))
			return
		}
	}

	if strings.TrimSpace(root.ConfigRC.Commands.Dir) != "" {
		// Ensure the commands directory exists
		if err := osMkdirAll(root.ConfigRC.Commands.Dir, 0755); err != nil {
			fmt.Println(color.RedString("Failed to create commands directory: %v", err))
			return
		}
	}

	commandFile := commandName + ".js"
	filePath := filepath.Join(root.ConfigRC.Commands.Dir, commandFile)

	// Absolute path to command file
	absCommandFilePath, _ := filepathAbs(filePath)

	cwd, _ := osGetwd()
	rootDir, _ := filepathAbs(cwd)
	prasmoidDef := filepath.Join(rootDir, "prasmoid.d.ts")

	// Calculate relative path from command file to prasmoid.d.ts
	relPath, err := filepathRel(filepath.Dir(absCommandFilePath), prasmoidDef)
	if err != nil {
		fmt.Println(color.RedString("Error calculating relative path: %v", err))
		return
	}

	templateFilled := fmt.Sprintf(commandTemplates["js"], relPath, commandName)

	err = osWriteFile(filePath, []byte(templateFilled), 0644)
	if err != nil {
		fmt.Println(color.RedString("Error writing to file: %v", err))
		return
	}

	fmt.Println(color.GreenString("Successfully created %s command at %s", commandName, color.BlueString(filePath)))
}
