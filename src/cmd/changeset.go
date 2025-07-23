/*
Copyright © 2025 PRAS
*/
package cmd

import (
	"github.com/spf13/cobra"
)

// changeset/rootCmd represents the changeset/root command
var ChangesetRootCmd = &cobra.Command{
	Use:   "changeset",
	Short: "Manage release lifecycle commands",
	Long:  "Handle creating, applying, and managing changesets and version bumps for the plasmoid.",
}

func init() {
	rootCmd.AddCommand(ChangesetRootCmd)
}
