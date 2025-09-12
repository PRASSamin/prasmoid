/*
Copyright 2025 PRAS
*/
package uninstall

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/PRASSamin/prasmoid/cmd"
)

func init() {
	cmd.RootCmd.AddCommand(UninstallCmd)
}

// UninstallCmd represents the production command
var UninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall plasmoid system-wide",
	Long:  "Uninstall the plasmoid from the system directories.",
	Run: func(cmd *cobra.Command, args []string) {
		if !utilsIsValidPlasmoid() {
			color.Red("Current directory is not a valid plasmoid.")
			return
		}
		if err := UninstallPlasmoid(); err != nil {
			color.Red("Failed to uninstall plasmoid:", err)
			return
		}
		color.Green("Plasmoid uninstalled successfully")
	},
}

var UninstallPlasmoid = func() error {
	isInstalled, where, err := utilsIsInstalled()
	if err != nil {
		return err
	}
	if isInstalled {
		if err := osRemoveAll(where); err != nil {
			return fmt.Errorf("failed to remove installation directory %s: %w", where, err)
		}
	}
	return nil
}
