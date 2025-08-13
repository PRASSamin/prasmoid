package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/PRASSamin/prasmoid/internal"
	"github.com/fatih/color"
	"github.com/fsnotify/fsnotify"
)

func Watcher() {
	const root = "."
	const debounceDuration = 500 * time.Millisecond

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("Error creating watcher: %v\n", err)
		return
	}
	defer func() {
		if err := watcher.Close(); err != nil {
			fmt.Printf("Error closing watcher: %v\n", err)
		}
	}()

	shouldWatch := func(path string, info os.FileInfo) bool {
		// ignore .git, .DS_Store, and build artifacts
		if strings.Contains(path, ".git") ||
			strings.Contains(path, ".DS_Store") {
			return false
		}
		// watch go files and directories
		return strings.HasSuffix(path, ".go") || info.IsDir()
	}

	// walk through the project directory and add watchers
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

				// create, write, remove
				if event.Op&fsnotify.Write == fsnotify.Write ||
					event.Op&fsnotify.Create == fsnotify.Create ||
					event.Op&fsnotify.Remove == fsnotify.Remove {

					if strings.HasSuffix(event.Name, ".go") {
						// get the base filename
						baseName := filepath.Base(event.Name)
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
	command := exec.Command("go", "build", "-ldflags=-s -w", "-o", strings.ToLower(internal.AppMetaData.Name), ".")

	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	if err := command.Run(); err != nil {
		color.Red("Build failed! " + err.Error())
		return
	}

	color.Green("Development build successful!")
}

func BuildCli(portable bool) {
	filename := filepath.Join(DIST_DIR, strings.ToLower(internal.AppMetaData.Name))
	version := internal.AppMetaData.Version
	var cgo = 1

	if portable {
		filename += "-portable"
		version += "-portable"
		cgo = 0
	}

	// inject version into the binary
	ldflags := fmt.Sprintf(`-s -w -X 'github.com/PRASSamin/prasmoid/internal.Version=%s'`, version)

	command := exec.Command("go", "build", "-ldflags", ldflags, "-o", filename, ".")

	// set cgo enabled
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

func generateChecksums() (map[string]string, error) {
	files, err := os.ReadDir(DIST_DIR)
	if err != nil {
		return nil, fmt.Errorf("failed to read dist directory: %v", err)
	}

	checksums := []string{}
	checksumMap := make(map[string]string)

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(DIST_DIR, file.Name())
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %v", file.Name(), err)
		}

		hash := sha256.Sum256(data)
		checksum := fmt.Sprintf("%x", hash)
		checksumMap[file.Name()] = checksum
		checksums = append(checksums, fmt.Sprintf("%s  %s", checksum, file.Name()))
	}

	checksumFile := filepath.Join(DIST_DIR, "SHA256SUMS")
	if err := os.WriteFile(checksumFile, []byte(strings.Join(checksums, "\n")), 0644); err != nil {
		return nil, fmt.Errorf("failed to write checksum file: %v", err)
	}

	color.Green("Generated checksums in %s", checksumFile)
	return checksumMap, nil
}
