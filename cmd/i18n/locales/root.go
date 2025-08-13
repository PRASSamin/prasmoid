/*
Copyright © 2025 PRAS
*/
package locales

import (
	"github.com/spf13/cobra"
	
	"github.com/PRASSamin/prasmoid/cmd/i18n"
)

func init() {
	i18nLocalesCmd.Aliases = []string{"l"}
	i18n.I18nCmd.AddCommand(i18nLocalesCmd)
}

var i18nLocalesCmd = &cobra.Command{
	Use:   "locales",
	Short: "Manages supported locales for your plasmoid.",
}
