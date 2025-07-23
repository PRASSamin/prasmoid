/*
Copyright 2025 PRAS
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "prasmoid",
	Short: "Manage plasmoid projects",
	Long:  "CLI for building, packaging, and managing KDE plasmoid projects efficiently.",
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	DiscoverAndRegisterCustomCommands(rootCmd)

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddGroup(&cobra.Group{
		ID:     "custom",
		Title:  "Custom Commands:",
	})
}