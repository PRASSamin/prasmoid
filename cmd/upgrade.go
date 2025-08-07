/*
Copyright 2025 PRAS

This command implements the Prasmoid CLI update functionality using a remote update script. Instead of embedding complex update logic directly in the Go code, which would increase the binary size by approximately 2MB, this approach leverages a lightweight shell script hosted on GitHub. This design choice ensures that Prasmoid remains lightweight while maintaining robust update capabilities.
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/PRASSamin/prasmoid/consts"
	"github.com/PRASSamin/prasmoid/utils"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(upgradeCmd)
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade to latest version of Prasmoid CLI.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := checkRoot(); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}

		if !utils.IsPackageInstalled(consts.CurlPackageName["binary"]) {
			pm, _ := utils.DetectPackageManager()
			var confirm bool
			confirmPrompt := &survey.Confirm{
				Message: "curl is not installed. Do you want to install it first?",
				Default: true,
			}
			if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
				return
			}
			
			if confirm {
				if err := utils.InstallPackage(pm, consts.CurlPackageName["binary"], consts.CurlPackageName); err != nil {
					color.Red("Failed to install curl:", err)
					return
				}
			} else {
				fmt.Println("Operation cancelled.")
				return
			}
		}

		scriptURL := "https://raw.githubusercontent.com/PRASSamin/prasmoid/main/update"
		
		exePath, err := os.Executable()
		if err != nil {
			color.Red("Failed to get current executable path: %v", err)
			return
		}

		cmdStr := fmt.Sprintf("curl -sSL %s | bash -s %s", scriptURL, exePath)

		command := exec.Command("bash", "-c", cmdStr)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr

		if err := command.Run(); err != nil {
			color.Red("Update failed: %v", err)
		}

		if err := os.Remove(GetCacheFilePath()); err != nil {}
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
