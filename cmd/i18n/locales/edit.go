/*
Copyright © 2025 PRAS
*/
package locales

import (
	"encoding/json"

	root "github.com/PRASSamin/prasmoid/cmd"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	i18nLocalesCmd.AddCommand(I18nLocalesEditCmd)
}

var I18nLocalesEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit locales for your plasmoid.",
	Run: func(cmd *cobra.Command, args []string) {
		if !utilsIsValidPlasmoid() {
			color.Red("Current directory is not a valid plasmoid.")
			return
		}

		currentLocales := root.ConfigRC.I18n.Locales

		locales := utilsAskForLocales(currentLocales)

		if locales != nil {
			root.ConfigRC.I18n.Locales = locales
			content, _ := json.MarshalIndent(root.ConfigRC, "", "  ")
			if err := osWriteFile("prasmoid.config.js", []byte(`/// <reference path="prasmoid.d.ts" />
/** @type {PrasmoidConfig} */
const config = `+string(content)), 0644); err != nil {
				color.Red("Error writing prasmoid.config.js: %v", err)
				return
			}
		}
	},
}
