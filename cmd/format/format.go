/*
Copyright 2025 PRAS
*/
package format

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/PRASSamin/prasmoid/cmd"
	"github.com/PRASSamin/prasmoid/consts"
	"github.com/PRASSamin/prasmoid/utils"
	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

var (
	utilsIsPackageInstalled = utils.IsPackageInstalled
	utilsDetectPackageManager = utils.DetectPackageManager
	surveyAskOne              = survey.AskOne
	utilsInstallQmlformat     = utils.InstallQmlformat
	execCommand               = exec.Command
)

var watch bool
var dir string

func init() {
	FormatCmd.Flags().BoolVarP(&watch, "watch", "w", false, "watch for changes")
	FormatCmd.Flags().StringVarP(&dir, "dir", "d", "./contents", "directory to format")
	cmd.RootCmd.AddCommand(FormatCmd)
}

// FormatCmd represents the format command
var FormatCmd = &cobra.Command{
	Use:   "format",
	Short: "Prettify QML files",
	Long:  "Automatically format QML source files to ensure consistent style and readability.",
	Run: func(cmd *cobra.Command, args []string) {
		if !utils.IsValidPlasmoid() {
			fmt.Println(color.RedString("Current directory is not a valid plasmoid."))
			return
		}
		if !utilsIsPackageInstalled(consts.QmlFormatPackageName["binary"]) {
			pm, _ := utilsDetectPackageManager()
			var confirm bool
			confirmPrompt := &survey.Confirm{
				Message: "qmlformat is not installed. Do you want to install it?",
				Default: true,
			}
			if err := surveyAskOne(confirmPrompt, &confirm); err != nil {
				return
			}

			if confirm {
				if err := utilsInstallQmlformat(pm); err != nil {
					fmt.Println(color.RedString("Failed to install qmlformat."))
					return
				}
				fmt.Println(color.GreenString("qmlformat installed successfully."))
			} else {
				fmt.Println(color.YellowString("Operation cancelled."))
				return
			}
		}

		crrPath, _ := os.Getwd()
		relPath := filepath.Join(crrPath, dir)
		if watch {
			prettifyOnWatch(relPath)
		} else {
			prettify(relPath)
		}
	},
}

func prettifyOnWatch(path string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Println(color.RedString("Failed to start watcher: %v", err))
		return
	}
	defer func() {
		if err := watcher.Close(); err != nil {
			log.Printf("Error closing watcher: %v", err)
		}
	}()

	done := make(chan bool)
	debounceTimers := make(map[string]*time.Timer)
	cooldownInProgress := make(map[string]bool)
	debounceDuration := 300 * time.Millisecond

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				file := event.Name
				if !utils.IsQmlFile(file) || event.Op&fsnotify.Write != fsnotify.Write {
					continue
				}

				// If cooldown already in progress, skip it
				if cooldownInProgress[file] {
					continue
				}

				// Mark as pending format
				cooldownInProgress[file] = true

				// Cancel any previous debounce for this file
				if timer, exists := debounceTimers[file]; exists {
					timer.Stop()
				}

				// Set new debounce
				debounceTimers[file] = time.AfterFunc(debounceDuration, func() {
					cmd := execCommand("qmlformat", "-i", file)
					if err := cmd.Run(); err != nil {
						fmt.Println(color.RedString("Format failed: %v.", err))
					} else {
						fmt.Printf("Formatted: %s\n", filepath.Base(file))
					}
					// Release the lock
					cooldownInProgress[file] = false
				})
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println(color.RedString("Watcher error: %v", err))
				return
			}
		}
	}()

	// Recursively watch all directories under the target path
	err = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return watcher.Add(p)
		}
		return nil
	})
	if err != nil {
		fmt.Println(color.RedString("Failed to watch directory: %v", err))
		return
	}

	fmt.Println("Formatter running in watch mode ...")
	<-done
}

func prettify(path string) {
	var files []string

	if err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if utils.IsQmlFile(path) {
			files = append(files, path)
		}
		return nil
	}); err != nil {
		fmt.Println(color.RedString("Error walking directory for prettify: %v", err))
		return
	}

	format(files)
	fmt.Println(color.GreenString("Formatted %d files.", len(files)))
}

func format(files []string) {
	formatter := execCommand("qmlformat", "-i")
	formatter.Args = append(formatter.Args, files...)
	formatter.Stdout = os.Stdout
	formatter.Stderr = os.Stderr
	if err := formatter.Run(); err != nil {
		fmt.Println(color.RedString("Failed to format qml files: %v", err))
	}
}
