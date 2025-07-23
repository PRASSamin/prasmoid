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
	"github.com/PRASSamin/prasmoid/utils"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var commandTemplates = map[string]string{
	"js": `/// <reference path="%s" />
const prasmoid = require("prasmoid");

prasmoid.Command({
    run: (ctx) => {
		const plasmoidId = prasmoid.getMetadata("Id");
		if (!plasmoidId) {
			console.red(
			"Could not get Plasmoid ID. Are you in a valid project directory?"
			);
			return;
		}

		console.color('%s Called', "blue");
	},
	short: "A brief description of your command.",
	long: "A longer description that spans multiple lines and likely contains examples\nand usage of using your command. For example:\n\nPlasmoid CLI is a CLI tool for KDE Plasmoid development.\nIt's a all-in-one tool for plasmoid development.",
	flags: [],
});`,
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
		config, _ := utils.LoadConfigRC()

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
				
				baseName := filepath.Join(config.Commands.Dir, name)
				if _, err := os.Stat(baseName + ".js"); err == nil {
					return errors.New("command name already exists with extension .js")
				}
				return nil
			})); err != nil {
				return
			}
		}

		template := commandTemplates["js"]

		if _, err := os.Stat(config.Commands.Dir); os.IsNotExist(err) {
			os.MkdirAll(config.Commands.Dir, 0755)
		}

		commandFile := commandName + ".js"
		filePath := filepath.Join(config.Commands.Dir, commandFile)

		// Create the new command file
		file, err := os.Create(filePath)
		if err != nil {
			color.Red("Error creating file: %v", err)
			return
		}
		defer file.Close()

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
