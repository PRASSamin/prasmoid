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
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	root "github.com/PRASSamin/prasmoid/cmd"
)

var rmCommandName string

func init() {
	commandsRemoveCmd.Flags().StringVarP(&rmCommandName, "name", "n", "", "Command name")
	commandsCmd.AddCommand(commandsRemoveCmd)
}

var commandsRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a custom command",
	Long:  "Remove a custom command to the project.",
	Run: func(cmd *cobra.Command, args []string) {
		availableCmds := []string{}
		if err := filepath.Walk(root.ConfigRC.Commands.Dir, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}
			availableCmds = append(availableCmds, fmt.Sprintf("%s (%s)", strings.TrimSuffix(info.Name(), filepath.Ext(info.Name())), info.Name()))
			return nil
		}); err != nil {
			color.Red("Error walking commands directory: %v", err)
			return
		}

		// Ask for runtime
		if strings.TrimSpace(rmCommandName) == "" {

			runtimePrompt := &survey.Select{
				Message: "Select a command to remove:",
				Options: availableCmds,
			}
			if err := survey.AskOne(runtimePrompt, &rmCommandName); err != nil {
				return
			}

		}

		// Extract filename using regex
		filenameRegex := regexp.MustCompile(`\(([^)]+)\)$`)
		matches := filenameRegex.FindStringSubmatch(rmCommandName)
		if len(matches) != 2 {
			color.Red("Invalid command format")
			return
		}
		fileName := matches[1]

		// Remove the file
		filePath := filepath.Join(root.ConfigRC.Commands.Dir, fileName)

		var confirm bool
		confirmPrompt := &survey.Confirm{
			Message: "Are you sure you want to remove this command?",
			Default: true,
		}
		if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
			return
		}
		if !confirm {
			return
		}

		err := os.Remove(filePath)
		if err != nil {
			color.Red("Error removing file: %v", err)
			return
		}
		color.Green("Successfully removed command: %s", fileName)
	},
}
