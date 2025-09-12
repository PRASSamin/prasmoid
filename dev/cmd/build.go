/*
Copyright 2025 PRAS
Development commands for Prasmoid
*/
package cmd

import (
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/PRASSamin/prasmoid/internal"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var DIST_DIR, _ = filepath.Abs("./dist/")
var ROOT_DIR, _ = filepath.Abs(".")

func init() {
	rootCmd.AddCommand(buildCmd)
}

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build prasmoid cli.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := os.RemoveAll(DIST_DIR); err != nil {
			color.Red("Warning: Failed to remove distribution directory: %v", err)
		}
		color.Blue("Building cli...")
		var builds = []bool{
			false,
			true,
		}
		for _, build := range builds {
			BuildCli(build)
		}

		// generate checksums
		_, err := generateChecksums()
		if err != nil {
			color.Red("Failed to generate checksums: %v", err)
			return
		}
	},
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
