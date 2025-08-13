/*
Copyright 2025 PRAS
*/
package cmd

import (
	"github.com/spf13/cobra"
)

func completionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "completion",
		Short: "Generate the autocompletion script for the specified shell",
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
}

func init() {
	completion := completionCommand()

	completion.Hidden = true
	rootCmd.AddCommand(completion)

}
