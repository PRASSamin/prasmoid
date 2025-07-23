/*
Copyright 2025 PRAS
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/PRASSamin/prasmoid/utils"
)

func init() {
	rootCmd.AddCommand(InstallCmd)
}

// InstallCmd represents the production command
var InstallCmd = &cobra.Command{
	Use:   "install",
	Short: "Install plasmoid system-wide",
	Long:  "Install the plasmoid to the system directories for production use.",
	Run: func(cmd *cobra.Command, args []string) {
		if !utils.IsValidPlasmoid() {
			color.Red("Current directory is not a valid plasmoid.")
			return
		}
		if err := InstallPlasmoid(); err != nil {
			color.Red("Failed to install plasmoid:", err)
			return
		}
		color.Green("Plasmoid installed successfully in %s", color.BlueString(utils.GetDevDest()))
		color.Cyan("\n- Please restart plasmashell to apply changes.")
	},
}

// Helper function to copy directories
func copyDir(src, dest string) error {
	err := os.MkdirAll(dest, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dest, err)
	}

	entries, err := os.ReadDir(src)
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
			data, err := os.ReadFile(srcPath)
			if err != nil {
				return fmt.Errorf("failed to read file %s: %v", srcPath, err)
			}

			err = os.WriteFile(destPath, data, 0644)
			if err != nil {
				return fmt.Errorf("failed to write file %s: %v", destPath, err)
			}
		}
	}

	return nil
}

func InstallPlasmoid() error {
	isInstalled, where, err := utils.IsInstalled()
	if err != nil {
		return err
	}
	if isInstalled {
		os.RemoveAll(where) 
	}
	
	os.MkdirAll(where, 0755)

	// Copy metadata.json
	srcMeta := "metadata.json"
	destMeta := filepath.Join(where, "metadata.json")
	metaData, err := os.ReadFile(srcMeta)
	if err != nil {
		return fmt.Errorf("failed to read metadata.json: %v", err)
	}
	err = os.WriteFile(destMeta, metaData, 0644)
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