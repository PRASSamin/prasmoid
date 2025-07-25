package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"

	"github.com/PRASSamin/prasmoid/internal"
	"github.com/PRASSamin/prasmoid/utils"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("System Monitor CLI Handler")
		fmt.Println("Usage:")
		fmt.Println("  go run main.go [command]")
		fmt.Println("Available Commands:")
		fmt.Println("  build       Build the cli")
		fmt.Println("  watch       Watch the cli for auto development build")
		return
	}


	switch os.Args[1] {
	case "build":
        var builds = []bool{
            false,
            true,
        }
        for _, build := range builds {
            BuildCli(build)
        }
	case "watch":
		Watcher()
	default:
		fmt.Println("System Monitor CLI Handler")
		fmt.Println("Usage:")
		fmt.Println("  go run main.go [command]")
		fmt.Println("Available Commands:")
		fmt.Println("  build       Build the cli")
		fmt.Println("  watch       Watch the cli for auto development build")
	}
}

func Watcher() {
    const root = "."
    const debounceDuration = 500 * time.Millisecond

    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        fmt.Printf("Error creating watcher: %v\n", err)
        return
    }
    defer watcher.Close()

    // Function to determine if we should watch a file/directory
    shouldWatch := func(path string, info os.FileInfo) bool {
        // Ignore .git, .DS_Store, and build artifacts
        if strings.Contains(path, ".git") || 
           strings.Contains(path, ".DS_Store") {
            return false
        }
        // Watch Go files and directories
        return strings.HasSuffix(path, ".go") || info.IsDir()
    }

    // Walk through the project directory and add watchers
    err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        if shouldWatch(path, info) {
            if err := watcher.Add(path); err != nil { 
                fmt.Printf("Warning: Failed to watch %s: %v\n", path, err)
            }
        }
        return nil
    })
    if err != nil {
        fmt.Printf("Error walking directory: %v\n", err)
        return
    }

    fmt.Println("Watching for changes in project directory...")
    fmt.Println("Press Ctrl+C to stop watching")
    BuildDevelopment()

    // Debounce function
    var buildPending bool
    var lastChangedFile string

    rebuild := func() {
        if buildPending {
            return
        }
        buildPending = true
        time.AfterFunc(debounceDuration, func() {
            buildPending = false
            BuildDevelopment()
        })
    } 

    go func() {
        for {
            select {
            case event, ok := <-watcher.Events:
                if !ok {
                    return
                }

                // Only rebuild on create, write, or remove events for Go files
                if event.Op&fsnotify.Write == fsnotify.Write || 
                   event.Op&fsnotify.Create == fsnotify.Create || 
                   event.Op&fsnotify.Remove == fsnotify.Remove {
                    
                    if strings.HasSuffix(event.Name, ".go") {
                        // Get the base filename
                        baseName := filepath.Base(event.Name)
                        // Only print if it's a different file than last time
                        if baseName != lastChangedFile {
                            fmt.Println("\nFile changed:", baseName)
                            lastChangedFile = baseName
                        }
                        rebuild()
                    }
                }

            case err, ok := <-watcher.Errors:
                if !ok {
                    return
                }
                fmt.Printf("Error: %v\n", err)
            }
        }
    }()

    // Keep the program running
    sig := make(chan os.Signal, 1)
    signal.Notify(sig, os.Interrupt)
    <-sig
}

func BuildDevelopment() {
    command := exec.Command("go", "build", "-ldflags=-s -w", "-o", strings.ToLower(internal.AppMeta.Name), "./src")
    
    command.Stdout = os.Stdout
    command.Stderr = os.Stderr
    command.Run()

    color.Green("Development build successful!")
}

func BuildCli(compress bool) {
	filename := "./dist/" + strings.ToLower(internal.AppMeta.Name)
    os.RemoveAll("./dist")
	version := internal.AppMeta.Version
    
	if compress {
		filename += "-compressed"
		version += "-compressed"
	}

	// Inject version into the binary
	ldflags := fmt.Sprintf(`-s -w -X 'github.com/PRASSamin/prasmoid/internal.Version=%s'`, version)
    
	command := exec.Command("go", "build", "-ldflags", ldflags, "-o", filename, "./src")

	command.Stdout = os.Stdout
    command.Stderr = os.Stderr
    command.Run()

	color.Green("Build successful!")

	if compress {
		color.Cyan("Starting executable compression...")

		if !utils.IsPackageInstalled("upx") {
			var confirm bool
			confirmPrompt := &survey.Confirm{
				Message: "UPX is not installed. Do you want to install it now?",
				Default: true,
			}
			if err := survey.AskOne(confirmPrompt, &confirm); err != nil {
				return
			}
			if confirm {
				if _, err := utils.TryInstallPackage("upx"); err != nil {
					color.Red("Failed to install UPX.")
					return
				}
				color.Green("UPX installed successfully.")
			} else {
				color.Yellow("Compression cancelled.")
				return
			}
		}

		command = exec.Command("upx", "--best", "--lzma", filename)
		command.Stdout = os.Stdout
        command.Stderr = os.Stderr
        command.Run()
		color.Green("Compression successful!")
	}
}
