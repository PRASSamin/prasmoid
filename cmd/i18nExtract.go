/*
Copyright © 2025 PRAS

[Documentation reference](https://develop.kde.org/docs/plasma/widget/translations-i18n/)
[Logic reference](https://github.com/Zren/plasma-applet-lib/blob/master/package/translate/merge)
*/
package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/PRASSamin/prasmoid/consts"
	"github.com/PRASSamin/prasmoid/utils"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	I18nExtractCmd.Flags().Bool("no-po", false, "Skip .po file generation")
	i18nCmd.AddCommand(I18nExtractCmd)
}

var I18nExtractCmd = &cobra.Command{
	Use:   "extract",
	Short: "Extract translatable strings from source files",
	Run: func(cmd *cobra.Command, args []string) {		
		if !utils.IsValidPlasmoid() {
			color.Red("Current directory is not a valid plasmoid.")
			return
		}

		if !utils.IsPackageInstalled(consts.GettextPackageName["binary"]) {
			pm, _ := utils.DetectPackageManager()
			var confirm bool
			confirmPrompt := &survey.Confirm{
				Message: "gettext is not installed. Do you want to install it first?",
				Default: true,
			}
			if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
				return
			}
			
			if confirm {
				if err := utils.InstallPackage(pm, consts.GettextPackageName["binary"], consts.GettextPackageName); err != nil {
					color.Red("Failed to install gettext:", err)
					return
				}
			} else {
				fmt.Println("Operation cancelled.")
				return
			}
		}

		color.Cyan("Extracting translatable strings...")

		// Create translations directory if it doesn't exist
		translationsDir := ConfigRC.I18n.Dir
		_ = os.MkdirAll(translationsDir, 0755)

		// Run xgettext to extract strings
		if err := runXGettext(translationsDir); err != nil {
			color.Red("Failed to extract strings: %v", err)
			return
		}

		color.Green("Successfully extracted strings to %s/template.pot", translationsDir)

		// Generate .po files for configured locales
		isPoGen, _ := cmd.Flags().GetBool("no-po")
		if !isPoGen {
			if err := generatePoFiles(translationsDir); err != nil {
				color.Red("Failed to generate .po files: %v", err)
				return
			}
		}
	},
}

func generatePoFiles(poDir string) error {
	if len(ConfigRC.I18n.Locales) == 0 {
		color.Yellow("No locales configured in prasmoid.config.js. Skipping .po file generation.")
		return nil
	}

	potFile := filepath.Join(poDir, "template.pot")

	for _, lang := range ConfigRC.I18n.Locales {
		poFile := filepath.Join(poDir, lang+".po")

		if _, err := os.Stat(poFile); os.IsNotExist(err) {
			// .po file doesn't exist, create it from the template
			color.Cyan("Creating %s...", poFile)
			cmd := exec.Command("msginit", "--no-translator", "-i", potFile, "-o", poFile, "-l", lang)
			if err := runCommand(cmd); err != nil {
				return fmt.Errorf("failed to create %s: %w", poFile, err)
			}

			content, _ := os.ReadFile(poFile)
			content = bytes.Replace(content, []byte("$__LANGUAGE__$"), []byte(lang), 1)
			_ = os.WriteFile(poFile, content, 0644)
		} else {
			// .po file exists, update it
			color.Cyan("Updating %s...", poFile)
			cmd := exec.Command("msgmerge", "--update", "--no-fuzzy-matching", poFile, potFile)
			if err := runCommand(cmd); err != nil {
				return fmt.Errorf("failed to merge %s: %w", poFile, err)
			}
		}
	}

	return cleanupBackupFiles(poDir)
}

func cleanupBackupFiles(poDir string) error {
	poDir, _ = filepath.Abs(poDir)
	backupFiles, err := doublestar.Glob(os.DirFS(poDir), "*.{po~,pot~,bak}")
	if err != nil {
		return fmt.Errorf("failed to find backup files: %w", err)
	}

	for _, file := range backupFiles {
		if err := os.Remove(filepath.Join(poDir, file)); err != nil {
			// Don't fail the whole process, just log a warning
			color.Yellow("Warning: failed to remove backup file %s: %v", file, err)
		}
	}
	return nil
}

func runXGettext(poDir string) error {
	pName, err := utils.GetDataFromMetadata("Name")
	plasmoidName, err := utils.EnsureStringAndValid("Name", pName, err)
	if err != nil {
		return err
	}

	v, err := utils.GetDataFromMetadata("Version")
	version, err := utils.EnsureStringAndValid("Version", v, err)
	if err != nil {
		return err
	}

	bAddress, err := utils.GetDataFromMetadata("BugReportUrl")
	bugAddress, err := utils.EnsureStringAndValid("BugReportUrl", bAddress, err)
	if err != nil {
		bugAddress = ""
	}

	authors, _ := utils.GetDataFromMetadata("Authors")

	potFileNew := filepath.Join(poDir, "template.pot.new")
	potFile := filepath.Join(poDir, "template.pot")

	// Find all translatable source files
	var srcFiles []string
	err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if strings.Contains(path, "node_modules") || strings.Contains(path, ".git") {
			return nil
		}		
		if !info.IsDir() && (strings.HasSuffix(path, ".qml") || strings.HasSuffix(path, ".js")) {
			srcFiles = append(srcFiles, path)
		}
		return nil
	})
	if err != nil {
		return err
	}

	if len(srcFiles) == 0 {
		color.Yellow("No translatable source files (.qml, .js) found.")
		return nil
	}

	// Extract from source files
	xgettextSrcCmd := exec.Command("xgettext",
		"--from-code=UTF-8", "--width=200", "--add-location=file",
		"-C", "-kde", "-ci18n",
		"-ki18n:1", "-ki18nc:1c,2", "-ki18np:1,2", "-ki18ncp:1c,2,3",
		"-kki18n:1", "-kki18nc:1c,2", "-kki18np:1,2", "-kki18ncp:1c,2,3",
		"-kxi18n:1", "-kxi18nc:1c,2", "-kxi18np:1,2", "-kxi18ncp:1c,2,3",
		"-kkxi18n:1", "-kkxi18nc:1c,2", "-kkxi18np:1,2", "-kkxi18ncp:1c,2,3",
		"-kI18N_NOOP:1", "-kI18NC_NOOP:1c,2",
		"-kI18N_NOOP2:1c,2", "-kI18N_NOOP2_NOSTRIP:1c,2",
		"-ktr2i18n:1", "-ktr2xi18n:1",
		"-kN_:1",
		"--package-name="+plasmoidName,
		"--msgid-bugs-address="+bugAddress,
		"--package-version="+version,
		// No --join-existing, this is the first and only run
		"-o", potFileNew,
	)
	xgettextSrcCmd.Args = append(xgettextSrcCmd.Args, srcFiles...)
	if err := runCommand(xgettextSrcCmd); err != nil {
		return fmt.Errorf("xgettext for source files failed: %w", err)
	}

	if _, err := os.Stat(potFileNew); os.IsNotExist(err) {
		return fmt.Errorf("no translatable strings found in source files")
	}

	// Post-process the new pot file
	postProcessPotFile(potFileNew, plasmoidName, authors)

	// Compare and replace the old pot file if necessary
	return handlePotFileUpdate(potFile, potFileNew)
}

func postProcessPotFile(path string, name string, authors interface{}) {
	input, err := os.ReadFile(path)
	if err != nil { return }

	var (
		email string = "EMAIL@ADDRESS"
		author string = "FIRST AUTHOR"
	)

	if authors != nil {
		if authorList, ok := authors.([]interface{}); ok && len(authorList) > 0 {
			if firstAuthor, ok := authorList[0].(map[string]interface{}); ok {
				if e, exists := firstAuthor["Email"]; exists && e != nil {
					if eStr, ok := e.(string); ok {
						email = eStr
					}
				}
				if n, exists := firstAuthor["Name"]; exists && n != nil {
					if nStr, ok := n.(string); ok {
						author = nStr
					}
				}
			}
		}
	}

	// Replace charset
	output := bytes.Replace(input, []byte("charset=CHARSET"), []byte("charset=UTF-8"), 1)
	// Replace title
	output = bytes.Replace(output, []byte("SOME DESCRIPTIVE TITLE."), []byte(fmt.Sprintf("Translation of %s in $__LANGUAGE__$", name)), 1)
	// Replace copyright year
	output = bytes.Replace(output, []byte("Copyright (C) YEAR THE PACKAGE'S COPYRIGHT HOLDER"), []byte(fmt.Sprintf("Copyright (C) %d %s", time.Now().Year(), author)), 1)
	// Replace author
    output = bytes.Replace(output, []byte("FIRST AUTHOR <EMAIL@ADDRESS>, YEAR."), []byte(fmt.Sprintf("%s <%s>, %d.", author, email, time.Now().Year())), 1)


	os.WriteFile(path, output, 0644)
}

func handlePotFileUpdate(oldPath, newPath string) error {
	// If old file doesn't exist, just rename the new one
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return os.Rename(newPath, oldPath)
	}

	// Compare files
	oldData, _ := os.ReadFile(oldPath)
	newData, _ := os.ReadFile(newPath)

	// Simple byte comparison, ignoring POT-Creation-Date for now
	if bytes.Equal(oldData, newData) {
		fmt.Println("No changes in translatable strings.")
		return os.Remove(newPath)
	} else {
		fmt.Println("Translatable strings updated.")
		return os.Rename(newPath, oldPath)
	}
}

func runCommand(cmd *exec.Cmd) error {
    var stderr bytes.Buffer
    cmd.Stderr = &stderr
    err := cmd.Run()
    if err != nil {
        return fmt.Errorf("command `%s` failed: %v\n%s", cmd.String(), err, stderr.String())
    }
    return nil
}
