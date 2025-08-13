/*
Copyright © 2025 PRAS
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/PRASSamin/prasmoid/utils"
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
		if !utils.IsValidPlasmoid() {
			color.Red("Current directory is not a valid plasmoid.")
			return
		}

		currentLocales := ConfigRC.I18n.Locales

		locales := utils.AskForLocales(currentLocales)

		if locales != nil {
			ConfigRC.I18n.Locales = locales
			content, _ := json.MarshalIndent(ConfigRC, "", "  ")
			fmt.Println(string(content))
			if err := os.WriteFile("prasmoid.config.js", []byte(`/// <reference path="prasmoid.d.ts" />
/** @type {PrasmoidConfig} */
const config = `+string(content)), 0644); err != nil {
				color.Red("Error writing prasmoid.config.js: %v", err)
				return
			}
		}
	},
}
