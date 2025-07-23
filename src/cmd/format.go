/*
Copyright 2025 PRAS
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"github.com/PRASSamin/prasmoid/utils"
)

var watch bool
var dir string

func init() {
	FormatCmd.Flags().BoolVarP(&watch, "watch", "w", false, "watch for changes")
	FormatCmd.Flags().StringVarP(&dir, "dir", "d", "./contents", "directory to format")
	rootCmd.AddCommand(FormatCmd)
}

// FormatCmd represents the format command
var FormatCmd = &cobra.Command{
	Use:   "format",
	Short: "Prettify QML files",
	Long:  "Automatically format QML source files to ensure consistent style and readability.",
	Run: func(cmd *cobra.Command, args []string) {
		if !utils.IsValidPlasmoid() {
			color.Red("Current directory is not a valid plasmoid.")
			return
		}
		if !utils.IsPackageInstalled("qmlformat") {
			var confirm bool
			confirmPrompt := &survey.Confirm{
				Message: "qmlformat is not installed. Do you want to install it?",
				Default: true,
			}
			if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
				return
			}
			
			if confirm {
				_, err := utils.TryInstallPackage("qmlformat")
				if err != nil {
					color.Red("Failed to install qmlformat.")
					return
				} 
				color.Green("qmlformat installed successfully.")
			} else {
				color.Yellow("Operation cancelled.")
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
		color.Red("Failed to start watcher:", err)
		return
	}
	defer watcher.Close()

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
				if !isQmlFile(file) || event.Op&fsnotify.Write != fsnotify.Write {
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
					cmd := exec.Command("qmlformat", "-i", file)
					if err := cmd.Run(); err != nil {
						color.Red("Format failed: %v", err)
						return
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
				color.Red("Watcher error: %v", err)
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
		color.Red("Failed to watch directory:", err)
		return
	}

	fmt.Printf("Formatter running in watch mode ...\n")
	<-done
}


func prettify(path string) {
	var files []string

	filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if isQmlFile(path) {
			files = append(files, path)
		}
		return nil
	})

	format(files)
	color.Green("Formatted %d files.", len(files))
}

func format(files []string){
	formatter := exec.Command("qmlformat", "-i")
	formatter.Args = append(formatter.Args, files...)
	formatter.Stdout = os.Stdout
	formatter.Stderr = os.Stderr
	if err := formatter.Run(); err != nil {
		fmt.Println("Failed to format qml files:", err)
	}
}

func isQmlFile(filename string) bool {
	return strings.HasSuffix(filename, ".qml")
}