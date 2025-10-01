/*
Copyright 2025 PRAS

This command implements the Prasmoid CLI fix functionality using a remote fix script. Instead of embedding complex fix logic directly in the Go code, which would increase the binary size by approximately 2MB, this approach leverages a lightweight shell script hosted on GitHub. This design choice ensures that Prasmoid remains lightweight while maintaining robust fix capabilities.
*/
package fix

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	root "github.com/PRASSamin/prasmoid/cmd"
)

func init() {
	if utilsIsPackageInstalled("curl") {
		cliFixCmd.Short = "Install missing dependencies."
	} else {
		cliFixCmd.Short = fmt.Sprintf("Install missing dependencies %s", color.RedString("(disabled)"))
	}
	cliFixCmd.GroupID = "cli"
	root.RootCmd.AddCommand(cliFixCmd)
}

var cliFixCmd = &cobra.Command{
	Use:   "fix",
	Run: func(cmd *cobra.Command, args []string) {
		if !utilsIsPackageInstalled("curl") {
			fmt.Println(color.YellowString("fix command is disabled due to missing curl dependency."))
			fmt.Println(color.BlueString("Please install curl and try again."))
			return
		}

		if err := utilsCheckRoot(); err != nil {
			fmt.Println(color.RedString(err.Error()))
			return
		}

		cmdStr := fmt.Sprintf("sudo curl -sSL %s | bash", scriptURL)

		command := execCommand("bash", "-c", cmdStr)
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr

		if err := command.Run(); err != nil {
			fmt.Println(color.RedString("Fix failed: %v", err))
		}
	},
}
