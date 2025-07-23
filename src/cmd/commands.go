/*
Copyright © 2025 PRAS
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// changeset/rootCmd represents the changeset/root command
var CommandsRootCmd = &cobra.Command{
	Use:   "commands",
	Short: "Manage project-specific custom commands",
}

func init() {
	rootCmd.AddCommand(CommandsRootCmd)
}
