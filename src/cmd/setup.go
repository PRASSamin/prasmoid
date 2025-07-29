/*
Copyright 2025 PRAS
*/
package cmd

import (
	"github.com/PRASSamin/prasmoid/consts"
	"github.com/PRASSamin/prasmoid/utils"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)


func init() {
	rootCmd.AddCommand(SetupCmd)
}

// SetupCmd represents the setup command
var SetupCmd = &cobra.Command{
		Use:   "setup",
		Short: "Setup development environment",
		Long:  "Install plasmoidviewer and other development dependencies.",
	Run: func(cmd *cobra.Command, args []string) {
		pm, err := utils.DetectPackageManager()
	if err != nil {
		color.Red("Failed to detect package manager.")
		return
	}

	if !utils.IsPackageInstalled(consts.QmlFormatPackageName["binary"]) {
		color.Yellow("Installing qmlformat...")
		if err := utils.InstallQmlformat(pm); err != nil {
			color.Red("Failed to install qmlformat.")
			return
		}
	}

	if !utils.IsPackageInstalled(consts.PlasmoidPreviewPackageName["binary"]) {
		color.Yellow("Installing plasmoidviewer...")
		if err := utils.InstallPlasmoidPreview(pm); err != nil {
			color.Red("Failed to install plasmoidviewer.")
			return
		}
	}

	if !utils.IsPackageInstalled(consts.CurlPackageName["binary"]) {
		color.Yellow("Installing curl...")
		if err := utils.InstallPackage(pm, consts.CurlPackageName["binary"], consts.CurlPackageName); err != nil {
			color.Red("Failed to install curl.")
			return
		}
	}
	},
}

