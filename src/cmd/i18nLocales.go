/*
Copyright © 2025 PRAS
*/
package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	i18nLocalesCmd.Aliases = []string{"l"}
	i18nCmd.AddCommand(i18nLocalesCmd)
}

var i18nLocalesCmd = &cobra.Command{
	Use:   "locales",
	Short: "Manages supported locales for your plasmoid.",
}