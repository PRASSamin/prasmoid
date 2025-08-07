/*
Copyright © 2025 PRAS
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// i18nCmd represents the i18n command
var i18nCmd = &cobra.Command{
	Use:   "i18n",
	Short: "Manage internationalization (i18n) for the plasmoid",
	Long:  `Provides tools to extract translatable strings and compile translation files.`,
}

func init() {
	rootCmd.AddCommand(i18nCmd)
}
