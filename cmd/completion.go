/*
Copyright 2025 PRAS
*/
package cmd

import (
	"github.com/spf13/cobra"
)

var completionCommand = &cobra.Command{
	Use:   "completion",
	Short: "Generate the autocompletion script for the specified shell",
	Hidden: true,
}

func init() {
	RootCmd.AddCommand(completionCommand)
}
