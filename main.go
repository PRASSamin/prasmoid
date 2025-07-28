package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"

	"github.com/PRASSamin/prasmoid/internal"
)

var DIST_DIR, _ = filepath.Abs("./dist/")

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
        os.RemoveAll(DIST_DIR)
        color.Blue("Building cli...")
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
    if err := command.Run(); err != nil {
        color.Red("Build failed! " + err.Error())
        return
    }

    color.Green("Development build successful!")
}

func BuildCli(portable bool) {
	filename := filepath.Join(DIST_DIR, strings.ToLower(internal.AppMeta.Name))
	version := internal.AppMeta.Version
    var cgo = 1
    
	if portable {
		filename += "-portable"
		version += "-portable"
        cgo = 0
	}

	// Inject version into the binary
	ldflags := fmt.Sprintf(`-s -w -X 'github.com/PRASSamin/prasmoid/internal.Version=%s'`, version)
    
	command := exec.Command("go", "build", "-ldflags", ldflags, "-o", filename, "./src")
    
    // Set CGO_ENABLED in the command's environment
    command.Env = append(os.Environ(), fmt.Sprintf("CGO_ENABLED=%d", cgo))

	command.Stdout = os.Stdout
    command.Stderr = os.Stderr
    if err := command.Run(); err != nil {
        color.Red("Build failed! " + err.Error())
        return
    }

    filenameSize, _ := os.Stat(filename)
	color.Green("Build successful! %s (%d mb)", filename, filenameSize.Size()/1024/1024)
}
