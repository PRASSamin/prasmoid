/*
Copyright 2025 PRAS
*/
package command

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	root "github.com/PRASSamin/prasmoid/cmd"
)

func init() {
	commandsRemoveCmd.Flags().StringP("name", "n", "", "Command name")
	commandsRemoveCmd.Flags().BoolP("force", "f", false, "Force removal without confirmation")
	commandsCmd.AddCommand(commandsRemoveCmd)
}

var commandsRemoveCmd = &cobra.Command{
	Use:   "remove",
	Short: "Remove a custom command",
	Long:  "Remove a custom command to the project.",
	Run: func(cmd *cobra.Command, args []string) {
		commandName, _ := cmd.Flags().GetString("name")
		force, _ := cmd.Flags().GetBool("force")
		RemoveCommand(commandName, force)
	},
}

func RemoveCommand(commandName string, force bool) {
	availableCmds := []string{}
	if err := filepathWalk(root.ConfigRC.Commands.Dir, func(path string, info os.FileInfo, err error) error {
		if info == nil || info.IsDir() {
			return nil
		}
		base := info.Name()
		nameWithoutExt := strings.TrimSuffix(base, filepath.Ext(base))
		availableCmds = append(availableCmds, fmt.Sprintf("%s (%s)", nameWithoutExt, base))
		return nil
	}); err != nil {
		fmt.Println(color.RedString("Error walking commands directory: %v", err))
		return
	}

	if len(availableCmds) == 0 {
		fmt.Println(color.RedString("No commands found in the commands directory."))
		return
	}

	// If not provided, prompt
	if strings.TrimSpace(commandName) == "" {
		commandNamePrompt := &survey.Select{
			Message: "Select a command to remove:",
			Options: availableCmds,
		}
		if err := surveyAskOne(commandNamePrompt, &commandName); err != nil {
			fmt.Println(color.RedString("Error asking for command name: %v", err))
			return
		}
	}

	// Normalize input
	commandName = strings.TrimSpace(commandName)

	var fileName string
	if strings.Contains(commandName, "(") && strings.Contains(commandName, ")") {
		// If user passed the select-menu style value
		re := regexpMustCompile(`\(([^)]+)\)$`)
		matches := re.FindStringSubmatch(commandName)
		if len(matches) == 2 {
			fileName = matches[1]
		}
	} else {
		// If user passed plain name or file name
		if !strings.Contains(filepath.Ext(commandName), ".") {
			// No extension -> add .js (or your default ext)
			fileName = commandName + ".js"
		} else {
			fileName = commandName
		}
	}

	// Verify it exists
	filePath := filepath.Join(root.ConfigRC.Commands.Dir, fileName)
	if _, err := osStat(filePath); os.IsNotExist(err) {
		fmt.Println(color.RedString("Command file does not exist: %s", fileName))
		return
	}

	// Confirmation
	var confirm = force
	if !force {
		confirmPrompt := &survey.Confirm{
			Message: fmt.Sprintf("Are you sure you want to remove %s?", fileName),
			Default: true,
		}
		if err := surveyAskOne(confirmPrompt, &confirm); err != nil {
			fmt.Println(color.RedString("Error asking for confirmation: %v", err))
			return
		}
	}
	if !confirm {
		fmt.Println(color.YellowString("Operation cancelled."))
		return
	}

	// Remove
	if err := osRemove(filePath); err != nil {
		fmt.Println(color.RedString("Error removing file: %v", err))
		return
	}
	fmt.Println(color.GreenString("Successfully removed command: %s", fileName))
}
