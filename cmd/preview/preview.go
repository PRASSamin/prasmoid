/*
Copyright 2025 PRAS
*/
package preview

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/PRASSamin/prasmoid/cmd"
	"github.com/PRASSamin/prasmoid/cmd/link"
	"github.com/PRASSamin/prasmoid/utils"
	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

type iWatcher interface {
	Add(string) error
	Close() error
	Events() chan fsnotify.Event
	Errors() chan error
}

type watcherWrapper struct {
	*fsnotify.Watcher
}

func (w *watcherWrapper) Events() chan fsnotify.Event {
	return w.Watcher.Events
}

func (w *watcherWrapper) Errors() chan error {
	return w.Watcher.Errors
}

// To enable mocking
var (
	// os/exec
	execCommand = exec.Command

	// utils
	utilsIsValidPlasmoid      = utils.IsValidPlasmoid
	utilsIsLinked             = utils.IsLinked
	utilsGetDevDest           = utils.GetDevDest
	utilsGetDataFromMetadata  = utils.GetDataFromMetadata
	utilsIsQmlFile            = utils.IsQmlFile
	utilsIsPackageInstalled   = utils.IsPackageInstalled

	// link
	linkLinkPlasmoid = link.LinkPlasmoid

	// survey
	surveyAskOne = survey.AskOne

	// fsnotify
	fsnotifyNewWatcher = func() (iWatcher, error) {
		w, err := fsnotify.NewWatcher()
		if err != nil {
			return nil, err
		}
		return &watcherWrapper{w}, nil
	}
	currentViewer *exec.Cmd
	viewerMutex   sync.Mutex

	// filepath
	filepathWalk = filepath.Walk

	// os/signal
	signalNotify = signal.Notify

	// time
	timeAfterFunc = time.AfterFunc

	// confirmation
	confirmLink         bool
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
		if !utilsIsPackageInstalled("plasmoidviewer") {
			fmt.Println(color.RedString("plasmoidviewer is not installed. Please install it and try again"))
			return
		}
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
					fmt.Println(color.RedString("Failed to link plasmoid: %v", err))
					return
				}
			} else {
				fmt.Println("Operation cancelled.")
				return
			}
		}

		

		if err := previewPlasmoid(watch); err != nil {
			fmt.Println(color.RedString("Failed to preview plasmoid: %v", err))
			return
		}
	},
}

var previewPlasmoid = func(watch bool) error {
	id, err := utilsGetDataFromMetadata("Id")
	if err != nil {
		return err
	}

	if watch {
		watchOnChange("./contents", id.(string))
		return nil
	}

	plasmoidViewer := execCommand("plasmoidviewer", "-a", id.(string))
	plasmoidViewer.Stdout = os.Stdout
	plasmoidViewer.Stderr = os.Stderr
	return plasmoidViewer.Run()
}

var watchOnChange = func(path string, id string) {
	watcher, err := fsnotifyNewWatcher()
	if err != nil {
		fmt.Println(color.RedString("Failed to start watcher: %v", err))
		return
	}
	defer func() {
		if err := watcher.Close(); err != nil {
			fmt.Println(color.RedString("Error closing watcher: %v", err))
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
		viewerMutex.Lock()
		if currentViewer != nil && currentViewer.Process != nil {
			if err := currentViewer.Process.Kill(); err != nil {
				log.Printf("Error killing current viewer process: %v", err)
			}
		}
		viewerMutex.Unlock()
		close(done)
	}()

	plasmoidViewer := execCommand("plasmoidviewer", "-a", id)
	plasmoidViewer.Stdout = os.Stdout
	plasmoidViewer.Stderr = os.Stderr
	viewerMutex.Lock()
	currentViewer = plasmoidViewer
	if err := currentViewer.Start(); err != nil {
		viewerMutex.Unlock()
		fmt.Println(color.RedString("Error starting plasmoidviewer: %v", err))
		return
	}
	viewerMutex.Unlock()

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events():
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
					viewerMutex.Lock()
					if currentViewer != nil {
						if err := currentViewer.Process.Kill(); err != nil {
							log.Printf("Error killing current viewer process: %v", err)
						}
					}
					viewerMutex.Unlock()

					plasmoidViewer := execCommand("plasmoidviewer", "-a", id)
					plasmoidViewer.Stdout = os.Stdout
					plasmoidViewer.Stderr = os.Stderr

					viewerMutex.Lock()
					currentViewer = plasmoidViewer
					if err := currentViewer.Start(); err != nil {
						viewerMutex.Unlock()
						fmt.Println(color.RedString("Error starting plasmoidviewer: %v", err))
						return
					}
					viewerMutex.Unlock()

					go func() {
						if err := plasmoidViewer.Wait(); err != nil {
							log.Printf("Error waiting for plasmoid viewer process: %v", err)
						}
						viewerMutex.Lock()
						if currentViewer == plasmoidViewer {
							currentViewer = nil
						}
						viewerMutex.Unlock()
					}()

					cooldownInProgress[file] = false
				})
			case err, ok := <-watcher.Errors():
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
		fmt.Println(color.RedString("Failed to watch directory: %v", err))
		return
	}

	fmt.Printf("Previewer running in watch mode ... Press Ctrl+C to exit\n")
	<-done
}