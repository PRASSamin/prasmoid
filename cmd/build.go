/*
Copyright 2025 PRAS
*/
package cmd

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
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
		if err := BuildPlasmoid(); err != nil {
			color.Red(err.Error())
		}
	},
}

func BuildPlasmoid() error {
	if !utils.IsValidPlasmoid() {
		return fmt.Errorf("current directory is not a valid plasmoid")
	}

	// compile translations
	color.Cyan("→ Compiling translations...")
	if err := CompileI18n(ConfigRC, false); err != nil {
		color.Red("Failed to compile translations: %v", err)
		// We don't necessarily want to stop the build if translations fail
		color.Yellow("Continuing build without translations...")
	} else {
		color.Green("Translations compiled successfully.")
	}

	color.Cyan("→ Starting plasmoid build...")
	plasmoidID, ierr := utils.GetDataFromMetadata("Id")
	version, verr := utils.GetDataFromMetadata("Version")
	if ierr != nil || verr != nil {
		return fmt.Errorf("failed to get plasmoid version: %v", fmt.Sprintf("%v or %v", ierr, verr))
	}
	zipFileName := plasmoidID.(string) + "-" + version.(string) + ".plasmoid"

	if err := os.RemoveAll(buildOutputDir); err != nil {
		return fmt.Errorf("failed to clean build dir: %v", err)
	}
	if err := os.MkdirAll(buildOutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create build dir: %v", err)
	}

	outFile, err := os.Create(filepath.Join(buildOutputDir, zipFileName))
	if err != nil {
		return fmt.Errorf("failed to create zip file: %v", err)
	}
	defer func() {
		if err := outFile.Close(); err != nil {
			log.Printf("error closing output file: %v", err)
		}
	}()

	zipWriter := zip.NewWriter(outFile)
	defer func() {
		if err := zipWriter.Close(); err != nil {
			log.Printf("Error closing zip writer: %v", err)
		}
	}()

	// Copy metadata.json
	if err := AddFileToZip(zipWriter, "metadata.json"); err != nil {
		return fmt.Errorf("error adding metadata.json: %v", err)
	}

	// Copy contents directory recursively
	if err := AddDirToZip(zipWriter, "contents"); err != nil {
		return fmt.Errorf("error adding contents/: %v", err)
	}

	color.Green("Build complete: %s", color.YellowString(filepath.Join(buildOutputDir, zipFileName)))
	return nil
}

func AddFileToZip(zipWriter *zip.Writer, filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Error closing file: %v", err)
		}
	}()

	w, err := zipWriter.Create(filename)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, file)
	return err
}

func AddDirToZip(zipWriter *zip.Writer, baseDir string) error {
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
		defer func() {
			if err := file.Close(); err != nil {
				log.Printf("Error closing file: %v", err)
			}
		}()

		w, err := zipWriter.Create(relPath)
		if err != nil {
			return err
		}

		_, err = io.Copy(w, file)
		return err
	})
}
