/*
Copyright 2025 PRAS
*/
package install

import (
	"fmt"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/PRASSamin/prasmoid/cmd"
)

func init() {
	cmd.RootCmd.AddCommand(InstallCmd)
}

// InstallCmd represents the production command
var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install plasmoid system-wide",
	Long:  "Install the plasmoid to the system directories for production use.",
	Run: func(cmd *cobra.Command, args []string) {
		if !utilsIsValidPlasmoid() {
			color.Red("Current directory is not a valid plasmoid.")
			return
		}
		if err := InstallPlasmoid(); err != nil {
			color.Red("Failed to install plasmoid:", err)
			return
		}

		dest, _ := utilsGetDevDest()
		color.Green("Plasmoid installed successfully in %s", color.BlueString(dest))
		color.Cyan("\n- Please restart plasmashell to apply changes.")
	},
}

// Helper function to copy directories
func copyDir(src, dest string) error {
	err := osMkdirAll(dest, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dest, err)
	}

	entries, err := osReadDir(src)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %v", src, err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		if entry.IsDir() {
			err = copyDir(srcPath, destPath)
			if err != nil {
				return err
			}
		} else {
			data, err := osReadFile(srcPath)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %v", srcPath, err)
			}

			err = osWriteFile(destPath, data, 0644)
			if err != nil {
				return fmt.Errorf("failed to write file %s: %v", destPath, err)
			}
		}
	}

	return nil
}

var InstallPlasmoid = func() error {
	isInstalled, where, err := utilsIsInstalled()
	if err != nil {
		return err
	}
	if isInstalled {
		if err := osRemoveAll(where); err != nil {
			fmt.Printf("Warning: Failed to remove existing installation at %s: %v\n", where, err)
		}
	}

	if err := osMkdirAll(where, 0755); err != nil {
		return fmt.Errorf("failed to create installation directory %s: %w", where, err)
	}

	// Copy metadata.json
	srcMeta := "metadata.json"
	destMeta := filepath.Join(where, "metadata.json")
	metaData, err := osReadFile(srcMeta)
	if err != nil {
		return fmt.Errorf("failed to read metadata.json: %v", err)
	}
	err = osWriteFile(destMeta, metaData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write metadata.json: %v", err)
	}

	// Copy contents directory
	srcContents := "contents"
	destContents := filepath.Join(where, "contents")
	err = copyDir(srcContents, destContents)
	if err != nil {
		return fmt.Errorf("failed to copy contents directory: %v", err)
	}
	return nil
}