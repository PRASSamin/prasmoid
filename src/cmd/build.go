/*
Copyright 2025 PRAS
*/
package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/PRASSamin/prasmoid/utils"
)

var buildOutputDir string

func init() {
	BuildCmd.Flags().StringVarP(&buildOutputDir, "output", "o", "./build", "Output folder")
	rootCmd.AddCommand(BuildCmd)
}

// buildCmd represents the build command
var BuildCmd = &cobra.Command{
		Use:   "build",
		Short: "Build the project",
		Long:  "Package the project files and generate the deployable .plasmoid archive.",
	Run: func(cmd *cobra.Command, args []string) {
		if !utils.IsValidPlasmoid() {
			color.Red("Current directory is not a valid plasmoid.")
			return
		}
		color.Cyan(" Starting plasmoid build...")
		plasmoidID, ierr := utils.GetDataFromMetadata("Id")
		version, verr := utils.GetDataFromMetadata("Version")
		if ierr != nil || verr != nil {
			color.Red("Failed to get plasmoid version: %v", fmt.Errorf("%v or %v", ierr, verr))
			return
		}
		zipFileName := plasmoidID + "-" + version + ".plasmoid"

		if err := os.RemoveAll(buildOutputDir); err != nil {
			color.Red("Failed to clean build dir: %v", err)
			return
		}
		if err := os.MkdirAll(buildOutputDir, 0755); err != nil {
			color.Red("Failed to create build dir: %v", err)
			return
		}

		outFile, err := os.Create(filepath.Join(buildOutputDir, zipFileName))
		if err != nil {
			color.Red("Failed to create zip file: %v", err)
			return
		}
		defer outFile.Close()

		zipWriter := zip.NewWriter(outFile)
		defer zipWriter.Close()

		// Copy metadata.json
		if err := addFileToZip(zipWriter, "metadata.json"); err != nil {
			color.Red("Error adding metadata.json: %v", err)
			return
		}

		// Copy contents directory recursively
		if err := addDirToZip(zipWriter, "contents"); err != nil {
			color.Red("Error adding contents/: %v", err)
			return
		}

		color.Green(" Build complete: %s", filepath.Join(buildOutputDir, zipFileName))
	},
}

func addFileToZip(zipWriter *zip.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	w, err := zipWriter.Create(filename)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, file)
	return err
}

func addDirToZip(zipWriter *zip.Writer, baseDir string) error {
	return filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath := path
		if strings.HasPrefix(path, "./") {
			relPath = path[2:]
		}

		if info.IsDir() {
			return nil // folders are implicit in zip
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		w, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(w, file)
		return err
	})
}
