/*
Copyright © 2025 PRAS
*/
package i18n

import (
	"github.com/PRASSamin/prasmoid/cmd"
	"github.com/spf13/cobra"
)

// i18nCmd represents the i18n command
var I18nCmd = &cobra.Command{
	Use:   "i18n",
	Short: "Manage internationalization (i18n) for the plasmoid",
	Long:  `Provides tools to extract translatable strings and compile translation files.`,
}

func init() {
	cmd.RootCmd.AddCommand(I18nCmd)
}
