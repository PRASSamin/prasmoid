/*
Copyright 2025 PRAS
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	updateCmd.AddCommand(meCmd)
}

var meCmd = &cobra.Command{
	Use:   "me",
	Short: "Update Prasmoid CLI",
	Long:  "Update Prasmoid CLI to the latest version from GitHub releases.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := checkRoot(); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		scriptURL := "https://raw.githubusercontent.com/PRASSamin/prasmoid/main/update"
		
		exePath, err := os.Executable()
		if err != nil {
			color.Red("Failed to get current executable path: %v", err)
			return
		}

		cmdStr := fmt.Sprintf("curl -sSL %s | bash -s %s %s", scriptURL, exePath, strings.Join(args, " "))

		command := exec.Command("bash", "-c", cmdStr)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr

		if err := command.Run(); err != nil {
			color.Red("Update failed: %v", err)
		}
	},
}

func checkRoot() error {
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current user: %v", err)
	}

	if currentUser.Uid != "0" {
		return fmt.Errorf("the requested operation requires superuser privileges. use `sudo %s`", strings.Join(os.Args[0:], " "))
	}
	return nil
}

