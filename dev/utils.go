package main

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"
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
	defer watcher.Close()

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

func getFileSha256(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", sha256.Sum256(data)), nil
}

func updatePackageBuilds(prasmoidChecksum string) error {
	color.Blue("Checking for PKGBUILD updates...")
	pkgbuildFiles, err := filepath.Glob("pkg/*/PKGBUILD")
	if err != nil {
		return fmt.Errorf("error finding PKGBUILD files: %w", err)
	}

	newVersion := internal.AppMetaData.Version
	licenseSha, err := getFileSha256(filepath.Join(ROOT_DIR, "LICENSE.md"))
	if err != nil {
		return fmt.Errorf("failed to get checksum for LICENSE.md: %w", err)
	}
	readmeSha, err := getFileSha256(filepath.Join(ROOT_DIR, "README.md"))
	if err != nil {
		return fmt.Errorf("failed to get checksum for README.md: %w", err)
	}

	for _, pkgbuildPath := range pkgbuildFiles {
		content, err := os.ReadFile(pkgbuildPath)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", pkgbuildPath, err)
		}
		originalContent := string(content)

		// Check version
		verRegex := regexp.MustCompile(`pkgver=([0-9.]+)`)
		matches := verRegex.FindStringSubmatch(originalContent)
		if len(matches) < 2 {
			color.Yellow("Could not find pkgver in %s, skipping update.", pkgbuildPath)
			continue
		}
		currentVersion := matches[1]

		if currentVersion == newVersion {
			color.Green("PKGBUILD %s is already up to date.", pkgbuildPath)
			continue
		}

		color.Yellow("Updating %s from version %s to %s...", pkgbuildPath, currentVersion, newVersion)

		// Update pkgver
		newContent := strings.Replace(originalContent, "pkgver="+currentVersion, "pkgver="+newVersion, 1)

		// Update source URLs
		newContent = strings.ReplaceAll(newContent, "v"+currentVersion, "v"+newVersion)

		// Update sha256sums
		shaRegex := regexp.MustCompile(`sha256sums=\((?s)(.*?)\)`)
		newShaBlock := fmt.Sprintf("sha256sums=(\n\t\"%s\"\n\t\"%s\"\n\t\"%s\"\n)", prasmoidChecksum, licenseSha, readmeSha)
		newContent = shaRegex.ReplaceAllString(newContent, newShaBlock)

		if err := os.WriteFile(pkgbuildPath, []byte(newContent), 0644); err != nil {
			return fmt.Errorf("failed to write updated content to %s: %w", pkgbuildPath, err)
		}
		color.Green("Successfully updated %s", pkgbuildPath)
	}
	return nil
}

func updateSpecFile(
	// prasmoidChecksum string
	) error {
	color.Blue("Checking for spec file updates...")
	specPath := "pkg/rpm/prasmoid.spec"

	newVersion := internal.AppMetaData.Version
	// licenseSha, err := getFileSha256(filepath.Join(ROOT_DIR, "LICENSE.md"))
	// if err != nil {
	// 	return fmt.Errorf("failed to get checksum for LICENSE.md: %w", err)
	// }
	// readmeSha, err := getFileSha256(filepath.Join(ROOT_DIR, "README.md"))
	// if err != nil {
	// 	return fmt.Errorf("failed to get checksum for README.md: %w", err)
	// }

	content, err := os.ReadFile(specPath)
	if err != nil {
		return fmt.Errorf("failed to read %s: %w", specPath, err)
	}
	originalContent := string(content)

	// Check version
	verRegex := regexp.MustCompile(`Version:\s+([0-9.]+)`)
	matches := verRegex.FindStringSubmatch(originalContent)
	if len(matches) < 2 {
		color.Yellow("Could not find Version in %s, skipping update.", specPath)
		return nil
	}
	currentVersion := matches[1]

	if currentVersion == newVersion {
		color.Green("Spec file %s is already up to date.", specPath)
		return nil
	}

	color.Yellow("Updating %s from version %s to %s...", specPath, currentVersion, newVersion)

	// Update Version
	newContent := strings.Replace(originalContent, "Version:        "+currentVersion, "Version:        "+newVersion, 1)

	// Update source URLs
	newContent = strings.ReplaceAll(newContent, "v"+currentVersion, "v"+newVersion)

	if err := os.WriteFile(specPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write updated content to %s: %w", specPath, err)
	}
	color.Green("Successfully updated %s", specPath)

	return nil
}