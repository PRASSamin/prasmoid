/*
Copyright 2025 PRAS
*/
package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/PRASSamin/prasmoid/consts"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var commandTemplates = map[string]string{
	"js": consts.JS_COMMAND_TEMPLATE,
}

var commandName string

func init() {
	CommandsAddCmd.Flags().StringVarP(&commandName, "name", "n", "", "Command name")
	CommandsRootCmd.AddCommand(CommandsAddCmd)
}

var CommandsAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a custom command",
	Long:  "Add a custom command to the project.",
	Run: func(cmd *cobra.Command, args []string) {
		invalidChars := regexp.MustCompile(`[\\/:*?"<>|\s@]`)

		// Ask for command name
		if strings.TrimSpace(commandName) == "" || invalidChars.MatchString(commandName) {
			namePrompt := &survey.Input{
				Message: "Command name:",
			}
			if err := survey.AskOne(namePrompt, &commandName, survey.WithValidator(func(ans interface{}) error {
				name := ans.(string)
				if name == "" {
					return errors.New("command name cannot be empty")
				}

				if invalidChars.MatchString(name) {
					return errors.New("invalid characters in command name")
				}

				baseName := filepath.Join(ConfigRC.Commands.Dir, name)
				if _, err := os.Stat(baseName + ".js"); err == nil {
					return errors.New("command name already exists with extension .js")
				}
				return nil
			})); err != nil {
				return
			}
		}

		template := commandTemplates["js"]

		if _, err := os.Stat(ConfigRC.Commands.Dir); os.IsNotExist(err) {
			if err := os.MkdirAll(ConfigRC.Commands.Dir, 0755); err != nil {
				color.Red("Failed to create commands directory: %v", err)
				return
			}
		}

		commandFile := commandName + ".js"
		filePath := filepath.Join(ConfigRC.Commands.Dir, commandFile)

		// Create the new command file
		file, err := os.Create(filePath)
		if err != nil {
			color.Red("Error creating file: %v", err)
			return
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
			return
		}

		templateFilled := fmt.Sprintf(template, relPath, commandName)

		_, err = file.WriteString(templateFilled)
		if err != nil {
			color.Red("Error writing to file: %v", err)
			return
		}

		color.Green("Successfully created %s command at %s", commandName, color.BlueString(filePath))
	},
}
