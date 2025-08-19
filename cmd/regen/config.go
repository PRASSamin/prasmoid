package regen

import (
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	regenCmd.AddCommand(regenConfigCmd)
}

var regenConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Regenerate prasmoid.config.js",
	Run: func(cmd *cobra.Command, args []string) {
		locales := utilsAskForLocales()

		if err := initCmdCreateConfigFile(locales); err != nil {
			color.Red("Failed to regenerate config file:", err)
			return
		}
		color.Green("Config file regenerated successfully.")
	},
}
