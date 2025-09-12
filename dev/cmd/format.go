/*
Copyright 2025 PRAS
Development commands for Prasmoid
*/
package cmd

import (
	"os"
	"os/exec"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(formatCmd)
}

var formatCmd = &cobra.Command{
	Use:   "format",
	Short: "Prettify the project.",
	Run: func(cmd *cobra.Command, args []string) {
		command := exec.Command("gofmt", "-s", "-w", ".")
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		if err := command.Run(); err != nil {
			color.Red("Failed to format project: %v", err)
			return
		}
	},
}
