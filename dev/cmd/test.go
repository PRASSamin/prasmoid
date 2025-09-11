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
	rootCmd.AddCommand(testCmd)
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Run test units.",
	Run: func(cmd *cobra.Command, args []string) {
		if contains(args, "help") {
			cmd := exec.Command("go", "help", "test")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				color.Red("Failed to run tests: %v", err)
				return
			}
			// This return is intentionally placed here to prevent executing the test command
			// when the help flag is used. The first return handles the error case, while
			// this one handles the successful help command case.
			return
		}

		command := []string{"go", "test", "./cmd/...", "./internal/...", "./utils/...", "-v", "-race", "-coverprofile=coverage.out", "-covermode=atomic"}
		command = append(command, args...)
		c := exec.Command(command[0], command[1:]...)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			color.Red("Failed to run tests: %v", err)
			return
		}
	},
}

func contains[T comparable](slice []T, item T) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
