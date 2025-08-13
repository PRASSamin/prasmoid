/*
Copyright © 2025 PRAS

[Documentation reference](https://develop.kde.org/docs/plasma/widget/translations-i18n/)
[Logic reference](https://github.com/Zren/plasma-applet-lib/blob/master/package/translate/build)
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/PRASSamin/prasmoid/types"
	"github.com/PRASSamin/prasmoid/utils"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var silent bool

func init() {
	I18nCompileCmd.Flags().Bool("restart", false, "Restart plasmashell after compiling")
	I18nCompileCmd.Flags().BoolVarP(&silent, "silent", "s", false, "Do not show progress messages")

	i18nCmd.AddCommand(I18nCompileCmd)
}

var I18nCompileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Compile .po files to binary .mo files",
	Run: func(cmd *cobra.Command, args []string) {
		if !utils.IsValidPlasmoid() {
			color.Red("Current directory is not a valid plasmoid.")
			return
		}

		if !utils.IsPackageInstalled("msgfmt") {
			color.Red("msgfmt command not found. Please install gettext.")
			return
		}

		if !silent {
			color.Cyan("Compiling translation files...")
		}

		if err := CompileI18n(ConfigRC, silent); err != nil {
			color.Red("Failed to compile messages: %v", err)
			return
		}

		if !silent {
			color.Green("Successfully compiled all translation files.")
		}

		if restart, _ := cmd.Flags().GetBool("restart"); restart {
			color.Cyan("Restarting plasmashell...")
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

	poFiles, err := filepath.Glob(filepath.Join(poDir, "*.po"))
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
		if err := os.MkdirAll(installDir, 0755); err != nil {
			return fmt.Errorf("could not create directory %s: %w", installDir, err)
		}

		// Run msgfmt
		cmd := exec.Command("msgfmt", "-o", moFile, poFile)
		if err := runCommand(cmd); err != nil {
			return fmt.Errorf("failed to compile %s: %w\nTry checking syntax with `msgfmt -c %s`", poFile, err, poFile)
		}

		compiledCount++
	}

	if compiledCount == 0 {
		color.Yellow("No translation files were compiled (check config.i18n.locales)")
	}

	return nil
}

func restartPlasmashell() error {
	if err := exec.Command("killall", "plasmashell").Run(); err != nil {
		color.Yellow("Could not stop plasmashell (it might not have been running).", err)
	}
	// Use kstart5 or kstart, depending on the system
	cmd := "kstart5"
	if _, err := exec.LookPath(cmd); err != nil {
		cmd = "kstart"
	}

	return exec.Command(cmd, "plasmashell").Start()
}
