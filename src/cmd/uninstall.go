/*
Copyright 2025 PRAS
*/
package cmd

import (
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/PRASSamin/prasmoid/utils"
)

func init() {
	rootCmd.AddCommand(UninstallCmd)
}

// UninstallCmd represents the production command
var UninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall plasmoid system-wide",
	Long:  "Uninstall the plasmoid from the system directories.",
	Run: func(cmd *cobra.Command, args []string) {
		if !utils.IsValidPlasmoid() {
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

func UninstallPlasmoid() error {
	isInstalled, where, err := utils.IsInstalled()
	if err != nil {
		return err
	}
	if isInstalled {
		os.RemoveAll(where) 
	}
	return nil
}