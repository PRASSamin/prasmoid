/*
Copyright 2025 PRAS
*/
package setup

import (
	"github.com/PRASSamin/prasmoid/cmd"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	cmd.RootCmd.AddCommand(SetupCmd)
}

// SetupCmd represents the setup command
var SetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup development environment",
	Long:  "Install plasmoidviewer and other development dependencies.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := utilsInstallDependencies(); err != nil {
			color.Red("Failed to install dependencies: %v", err)
			return
		}
	},
}
