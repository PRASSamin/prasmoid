/*
Copyright © 2025 PRAS
*/
package changeset

import (
	"github.com/PRASSamin/prasmoid/cmd"
	"github.com/spf13/cobra"
)

// changesetCmd represents the changeset command
var changesetCmd = &cobra.Command{
	Use:   "changeset",
	Short: "Manage release lifecycle commands",
	Long:  "Handle creating, applying, and managing changesets and version bumps for the plasmoid.",
}

func init() {
	cmd.RootCmd.AddCommand(changesetCmd)
}
