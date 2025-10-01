/*
Copyright 2025 PRAS

This command implements the Prasmoid CLI update functionality using a remote update script. Instead of embedding complex update logic directly in the Go code, which would increase the binary size by approximately 2MB, this approach leverages a lightweight shell script hosted on GitHub. This design choice ensures that Prasmoid remains lightweight while maintaining robust update capabilities.
*/
package upgrade

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	root "github.com/PRASSamin/prasmoid/cmd"
)

func init() {
	if utilsIsPackageInstalled("curl") {
		upgradeCmd.Short = "Upgrade to latest version of Prasmoid CLI."
	} else {
		upgradeCmd.Short = fmt.Sprintf("Upgrade to latest version of Prasmoid CLI %s", color.RedString("(disabled)"))
	}
	upgradeCmd.GroupID = "cli"
	root.RootCmd.AddCommand(upgradeCmd)
}

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Run: func(cmd *cobra.Command, args []string) {
		if !utilsIsPackageInstalled("curl") {
			fmt.Println(color.RedString("upgrade command is disabled due to missing dependencies."))
			fmt.Println(color.BlueString("Please install curl and try again."))
			return
		}

		if err := checkRoot(); err != nil {
			fmt.Println(color.RedString(err.Error()))
			return
		}

		exePath, err := osExecutable()
		if err != nil {
			fmt.Println(color.RedString("Failed to get current executable path: %v", err))
			return
		}

		cmdStr := fmt.Sprintf("sudo curl -sSL %s | bash -s %s", scriptURL, exePath)

		command := execCommand("bash", "-c", cmdStr)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr

		if err := command.Run(); err != nil {
			fmt.Println(color.RedString("Update failed: %v", err))
		}

		if err := osRemove(rootGetCacheFilePath()); err != nil {
			// Log the error, but don't fail the upgrade process
			fmt.Println(color.YellowString("Warning: Failed to remove update cache file: %v\n", err))
		}
	},
}

var checkRoot = func() error {
	currentUser, err := userCurrent()
	if err != nil {
		return fmt.Errorf("failed to get current user: %v", err)
	}

	if currentUser.Uid != "0" {
		return fmt.Errorf("the requested operation requires superuser privileges. use `sudo %s`", strings.Join(os.Args[0:], " "))
	}
	return nil
}
