/*
Copyright 2025 PRAS
*/
package preview

import (
	"fmt"
	"log"
	"os"
	"syscall"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"

	"github.com/PRASSamin/prasmoid/cmd"
	"github.com/PRASSamin/prasmoid/consts"
)

func init() {
	PreviewCmd.Flags().BoolP("watch", "w", false, "Watch for changes and automatically restart the preview. Note: This uses hot restart instead of hot reload, which may be slower.")
	cmd.RootCmd.AddCommand(PreviewCmd)
}

// PreviewCmd represents the preview command
var PreviewCmd = &cobra.Command{
	Use:   "preview",
	Short: "Enter plasmoid preview mode",
	Long:  "Launch the plasmoid in preview mode for testing and development.",
	Run: func(cmd *cobra.Command, args []string) {
		if !utilsIsValidPlasmoid() {
			fmt.Println(color.RedString("Current directory is not a valid plasmoid."))
			return
		}
		watch, _ := cmd.Flags().GetBool("watch")

		if !utilsIsLinked() {
			confirmPrompt := &survey.Confirm{
				Message: "Plasmoid is not linked. Do you want to link it first?",
				Default: true,
			}
			if err := surveyAskOne(confirmPrompt, &confirmLink); err != nil {
				return
			}

			if confirmLink {
				dest, err := utilsGetDevDest()
				if err != nil {
					fmt.Println(color.RedString(err.Error()))
					return
				}
				if err := linkLinkPlasmoid(dest); err != nil {
					fmt.Println(color.RedString("Failed to link plasmoid:", err))
					return
				}
			} else {
				fmt.Println("Operation cancelled.")
				return
			}
		}

		if !utilsIsPackageInstalled(consts.PlasmoidPreviewPackageName["binary"]) {
			pm, _ := utilsDetectPackageManager()
			confirmPrompt := &survey.Confirm{
				Message: "plasmoidviewer is not installed. Do you want to install it first?",
				Default: true,
			}
			if err := surveyAskOne(confirmPrompt, &confirmInstallation); err != nil {
				return
			}

			if confirmInstallation {
				if err := utilsInstallPackage(pm, consts.PlasmoidPreviewPackageName["binary"], consts.PlasmoidPreviewPackageName); err != nil {
					fmt.Println(color.RedString("Failed to install plasmoidviewer:", err))
					return
				}
			} else {
				fmt.Println("Operation cancelled.")
				return
			}
		}

		if err := previewPlasmoid(watch); err != nil {
			fmt.Println(color.RedString("Failed to preview plasmoid:", err))
			return
		}
	},
}

var previewPlasmoid = func(watch bool) error {
	id, err := utilsGetDataFromMetadata("Id")
	if err != nil {
		return err
	}

	quit := make(chan os.Signal, 1)
	signalNotify(quit, os.Interrupt, syscall.SIGTERM)

	if watch {
		watchOnChange("./contents", id.(string))
		return nil
	}

	plasmoidViewer := execCommand("plasmoidviewer", "-a", id.(string))
	plasmoidViewer.Stdout = os.Stdout
	plasmoidViewer.Stderr = os.Stderr
	return nil
}

var watchOnChange = func(path string, id string) {
	watcher, err := fsnotifyNewWatcher()
	if err != nil {
		fmt.Println(color.RedString("Failed to start watcher:", err))
		return
	}
	defer func() {
		if err := watcher.Close(); err != nil {
			fmt.Println(color.RedString("Error closing watcher:", err))
		}
	}()

	done := make(chan bool)
	debounceTimers := make(map[string]*time.Timer)
	cooldownInProgress := make(map[string]bool)
	debounceDuration := 300 * time.Millisecond

	// Set up signal handling
	quit := make(chan os.Signal, 1)
	signalNotify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		if currentViewer != nil && currentViewer.Process != nil {
			if err := currentViewer.Process.Kill(); err != nil {
				log.Printf("Error killing current viewer process: %v", err)
			}
			if err := currentViewer.Wait(); err != nil {
				log.Printf("Error waiting for current viewer process: %v", err)
			}
		}
		close(done)
	}()

	plasmoidViewer := execCommand("plasmoidviewer", "-a", id)
	plasmoidViewer.Stdout = os.Stdout
	plasmoidViewer.Stderr = os.Stderr
	currentViewer = plasmoidViewer

	if err := plasmoidViewer.Start(); err != nil {
		fmt.Println(color.RedString("Error starting plasmoidviewer: %v", err))
		return
	}

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				file := event.Name
				if !utilsIsQmlFile(file) || event.Op&fsnotify.Write != fsnotify.Write {
					continue
				}

				if cooldownInProgress[file] {
					continue
				}

				cooldownInProgress[file] = true

				if timer, exists := debounceTimers[file]; exists {
					timer.Stop()
				}

				debounceTimers[file] = timeAfterFunc(debounceDuration, func() {
					if currentViewer != nil {
						if err := currentViewer.Process.Kill(); err != nil {
							log.Printf("Error killing current viewer process: %v", err)
						}
						if err := currentViewer.Wait(); err != nil {
							log.Printf("Error waiting for current viewer process: %v", err)
						}
					}

					plasmoidViewer := execCommand("plasmoidviewer", "-a", id)
					plasmoidViewer.Stdout = os.Stdout
					plasmoidViewer.Stderr = os.Stderr

					if err := plasmoidViewer.Start(); err != nil {
						fmt.Println(color.RedString("Error starting plasmoidviewer: %v", err))
						return
					}
					currentViewer = plasmoidViewer

					go func() {
						if err := plasmoidViewer.Wait(); err != nil {
							log.Printf("Error waiting for plasmoid viewer process: %v", err)
						}
						if currentViewer == plasmoidViewer {
							currentViewer = nil
						}
					}()

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

	if err := filepathWalk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return watcher.Add(p)
		}
		return nil
	}); err != nil {
		fmt.Println(color.RedString("Failed to watch directory:", err))
		return
	}

	fmt.Printf("Previewer running in watch mode ... Press Ctrl+C to exit\n")
	<-done
}