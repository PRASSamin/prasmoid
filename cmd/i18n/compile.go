/*
Copyright © 2025 PRAS

[Documentation reference](https://develop.kde.org/docs/plasma/widget/translations-i18n/)
[Logic reference](https://github.com/Zren/plasma-applet-lib/blob/master/package/translate/build)
*/
package i18n

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	root "github.com/PRASSamin/prasmoid/cmd"
	"github.com/PRASSamin/prasmoid/consts"
	"github.com/PRASSamin/prasmoid/types"
	"github.com/PRASSamin/prasmoid/utils"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var silent bool
var confirm bool

func init() {
	I18nCompileCmd.Flags().Bool("restart", false, "Restart plasmashell after compiling")
	I18nCompileCmd.Flags().BoolVarP(&silent, "silent", "s", false, "Do not show progress messages")

	I18nCmd.AddCommand(I18nCompileCmd)
}

var I18nCompileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Compile .po files to binary .mo files",
	Run: func(cmd *cobra.Command, args []string) {
		if !utils.IsValidPlasmoid() {
			fmt.Println(color.RedString("Current directory is not a valid plasmoid."))
			return
		}

		if !utils.IsPackageInstalled(consts.GettextPackageName["binary"]) {
			pm, _ := utils.DetectPackageManager()
			confirmPrompt := &survey.Confirm{
				Message: "gettext is not installed. Do you want to install it first?",
				Default: true,
			}
			if !confirm {
				if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
					return
				}
			}

			if confirm {
				if err := utils.InstallPackage(pm, consts.GettextPackageName["binary"], consts.GettextPackageName); err != nil {
					fmt.Println(color.RedString("Failed to install gettext:", err))
					return
				}
			} else {
				fmt.Println("Operation cancelled.")
				return
			}
		}

		if !silent {
			fmt.Println(color.CyanString("Compiling translation files..."))
		}

		if err := CompileI18n(root.ConfigRC, silent); err != nil {
			fmt.Println(color.RedString("Failed to compile messages: %v", err))
			return
		}

		if !silent {
			fmt.Println(color.GreenString("Successfully compiled all translation files."))
		}

		if restart, _ := cmd.Flags().GetBool("restart"); restart {
			fmt.Println(color.CyanString("Restarting plasmashell..."))
			if err := restartPlasmashell(); err != nil {
				color.Red("Failed to restart plasmashell: %v", err)
			}
		}
	},
}

func CompileI18n(config types.Config, silent bool) error {
	poDir := config.I18n.Dir

	plasmoidId, err := utils.GetDataFromMetadata("Id")
	plasmoidIdStr, err := utils.EnsureStringAndValid("Id", plasmoidId, err)
	if err != nil {
		return err
	}
	projectName := "plasma_applet_" + plasmoidIdStr

	poFiles, err := filepathGlob(filepath.Join(poDir, "*.po"))
	if err != nil {
		return fmt.Errorf("could not find .po files: %w", err)
	}

	if len(poFiles) == 0 {
		color.Yellow("No .po files found to compile.")
		return nil
	}

	// Create a set of desired locales from config
	desiredLocales := make(map[string]bool)
	for _, loc := range config.I18n.Locales {
		desiredLocales[loc] = true
	}

	compiledCount := 0

	for _, poFile := range poFiles {
		lang := strings.TrimSuffix(filepath.Base(poFile), ".po")

		if !desiredLocales[lang] {
			color.Yellow("Skipping %s (not listed in prasmoid.config.js)", poFile)
			continue
		}

		// Define the output path
		installDir := filepath.Join("contents", "locale", lang, "LC_MESSAGES")
		moFile := filepath.Join(installDir, projectName+".mo")

		if !silent {
			color.Cyan("Compiling %s → %s", poFile, moFile)
		}

		// Create the destination directory
		if err := osMkdirAll(installDir, 0755); err != nil {
			return fmt.Errorf("could not create directory %s: %w", installDir, err)
		}

		// Run msgfmt
		cmd := execCommand("msgfmt", "-o", moFile, poFile)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to compile %s: %w Output: %s Try checking syntax with `msgfmt -c %s`", poFile, err, string(output), poFile)
		}

		compiledCount++
	}

	if compiledCount == 0 {
		color.Yellow("No translation files were compiled (check config.i18n.locales)")
	}

	return nil
}

func restartPlasmashell() error {
	if err := execCommand("killall", "plasmashell").Run(); err != nil {
		color.Yellow("Could not stop plasmashell (it might not have been running).", err)
	}
	// Use kstart5 or kstart, depending on the system
	cmdName := "kstart5"
	if _, err := execLookPath(cmdName); err != nil {
		cmdName = "kstart"
	}

	return execCommand(cmdName, "plasmashell").Start()
}
