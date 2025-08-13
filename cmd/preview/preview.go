/*
Copyright 2025 PRAS
*/
package preview

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"

	"github.com/PRASSamin/prasmoid/cmd"
	"github.com/PRASSamin/prasmoid/cmd/link"
	"github.com/PRASSamin/prasmoid/consts"
	"github.com/PRASSamin/prasmoid/utils"
)

func init() {
	PreviewCmd.Flags().BoolP("watch", "w", false, "Watch for changes and automatically restart the preview. Note: This uses hot restart instead of hot reload, which may be slower.")
	cmd.RootCmd.AddCommand(PreviewCmd)
}

var (
	currentViewer *exec.Cmd
)

// PreviewCmd represents the preview command
var PreviewCmd = &cobra.Command{
	Use:   "preview",
	Short: "Enter plasmoid preview mode",
	Long:  "Launch the plasmoid in preview mode for testing and development.",
	Run: func(cmd *cobra.Command, args []string) {
		if !utils.IsValidPlasmoid() {
			color.Red("Current directory is not a valid plasmoid.")
			return
		}
		watch, _ := cmd.Flags().GetBool("watch")

		if !utils.IsLinked() {
			var confirm bool
			confirmPrompt := &survey.Confirm{
				Message: "Plasmoid is not linked. Do you want to link it first?",
				Default: true,
			}
			if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
				return
			}

			if confirm {
				dest, err := utils.GetDevDest()
				if err != nil {
					color.Red(err.Error())
					return
				}
				if err := link.LinkPlasmoid(dest); err != nil {
					color.Red("Failed to link plasmoid:", err)
					return
				}
			} else {
				fmt.Println("Operation cancelled.")
				return
			}
		}

		if !utils.IsPackageInstalled(consts.PlasmoidPreviewPackageName["binary"]) {
			pm, _ := utils.DetectPackageManager()
			var confirm bool
			confirmPrompt := &survey.Confirm{
				Message: "plasmoidviewer is not installed. Do you want to install it first?",
				Default: true,
			}
			if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
				return
			}

			if confirm {
				if err := utils.InstallPackage(pm, consts.PlasmoidPreviewPackageName["binary"], consts.PlasmoidPreviewPackageName); err != nil {
					color.Red("Failed to install plasmoidviewer:", err)
					return
				}
			} else {
				fmt.Println("Operation cancelled.")
				return
			}
		}

		if err := previewPlasmoid(watch); err != nil {
			color.Red("Failed to preview plasmoid:", err)
			return
		}
	},
}

func previewPlasmoid(watch bool) error {
	id, err := utils.GetDataFromMetadata("Id")
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	if watch {
		watchOnChange("./contents", id.(string))
		return nil
	}

	plasmoidViewer := exec.CommandContext(ctx, "plasmoidviewer", "-a", id.(string))
	plasmoidViewer.Stdout = os.Stdout
	plasmoidViewer.Stderr = os.Stderr
	return plasmoidViewer.Run()
}

func watchOnChange(path string, id string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		color.Red("Failed to start watcher:", err)
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

	// Set up signal handling
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		if currentViewer != nil {
			if err := currentViewer.Process.Kill(); err != nil {
				log.Printf("Error killing current viewer process: %v", err)
			}
			if err := currentViewer.Wait(); err != nil {
				log.Printf("Error waiting for current viewer process: %v", err)
			}
		}
		close(done)
	}()

	plasmoidViewer := exec.Command("plasmoidviewer", "-a", id)
	plasmoidViewer.Stdout = os.Stdout
	plasmoidViewer.Stderr = os.Stderr

	if err := plasmoidViewer.Start(); err != nil {
		color.Red("Error starting plasmoidviewer: %v", err)
		return
	}
	currentViewer = plasmoidViewer

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

				if cooldownInProgress[file] {
					continue
				}

				cooldownInProgress[file] = true

				if timer, exists := debounceTimers[file]; exists {
					timer.Stop()
				}

				debounceTimers[file] = time.AfterFunc(debounceDuration, func() {
					if currentViewer != nil {
						if err := currentViewer.Process.Kill(); err != nil {
							log.Printf("Error killing current viewer process: %v", err)
						}
						if err := currentViewer.Wait(); err != nil {
							log.Printf("Error waiting for current viewer process: %v", err)
						}
					}

					plasmoidViewer := exec.Command("plasmoidviewer", "-a", id)
					plasmoidViewer.Stdout = os.Stdout
					plasmoidViewer.Stderr = os.Stderr

					if err := plasmoidViewer.Start(); err != nil {
						color.Red("Error starting plasmoidviewer: %v", err)
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
				color.Red("Watcher error: %v", err)
				return
			}
		}
	}()

	if err := filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return watcher.Add(p)
		}
		return nil
	}); err != nil {
		color.Red("Failed to watch directory:", err)
		return
	}

	fmt.Printf("Previewer running in watch mode ... Press Ctrl+C to exit\n")
	<-done
}
