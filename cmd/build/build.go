/*
Copyright 2025 PRAS
*/
package build

import (
	"archive/zip"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	root "github.com/PRASSamin/prasmoid/cmd"
)

var buildOutputDir string

func init() {
	BuildCmd.Flags().StringVarP(&buildOutputDir, "output", "o", "./build", "Output folder")
	root.RootCmd.AddCommand(BuildCmd)
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
	if !utilsIsValidPlasmoid() {
		return fmt.Errorf("current directory is not a valid plasmoid")
	}

	// compile translations
	color.Cyan("→ Compiling translations...")
	if err := i18nCompileI18n(root.ConfigRC, false); err != nil {
		color.Red("Failed to compile translations: %v", err)
		// We don't necessarily want to stop the build if translations fail
		color.Yellow("Continuing build without translations...")
	} else {
		color.Green("Translations compiled successfully.")
	}

	color.Cyan("→ Starting plasmoid build...")
	plasmoidID, ierr := utilsGetDataFromMetadata("Id")
	version, verr := utilsGetDataFromMetadata("Version")
	if ierr != nil || verr != nil {
		return fmt.Errorf("invalid metadata: %v", fmt.Sprintf("%v or %v", ierr, verr))
	}
	zipFileName := plasmoidID.(string) + "-" + version.(string) + ".plasmoid"

	if err := osRemoveAll(buildOutputDir); err != nil {
		return fmt.Errorf("failed to clean build dir: %v", err)
	}
	if err := osMkdirAll(buildOutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create build dir: %v", err)
	}

	outFile, err := osCreate(filepath.Join(buildOutputDir, zipFileName))
	if err != nil {
		return fmt.Errorf("failed to create zip file: %v", err)
	}
	defer func() {
		if err := outFile.Close(); err != nil {
			log.Printf("error closing output file: %v", err)
		}
	}()

	zipWriter := zipNewWriter(outFile)
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

var AddFileToZip = func(zipWriter *zip.Writer, filename string) error {
	file, err := osOpen(filename)
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
	_, err = ioCopy(w, file)
	return err
}

var AddDirToZip = func(zipWriter *zip.Writer, baseDir string) error {
	return filepathWalk(baseDir, func(path string, info os.FileInfo, err error) error {
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

		file, err := osOpen(path)
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

		_, err = ioCopy(w, file)
		return err
	})
}
