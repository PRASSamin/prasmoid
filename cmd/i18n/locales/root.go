/*
Copyright © 2025 PRAS
*/
package locales

import (
	"github.com/spf13/cobra"

	"github.com/PRASSamin/prasmoid/cmd/i18n"
)

var i18nLocalesCmd = &cobra.Command{
	Use:   "locales",
	Short: "Manages supported locales for your plasmoid.",
}

func init() {
	i18n.I18nCmd.AddCommand(i18nLocalesCmd)
}
