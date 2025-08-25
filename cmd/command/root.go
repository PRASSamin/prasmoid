/*
Copyright © 2025 PRAS
*/
package command

import (
	"github.com/PRASSamin/prasmoid/cmd"
	"github.com/spf13/cobra"
)

// commandsCmd represents the commands command
var commandsCmd = &cobra.Command{
	Use:   "command",
	Short: "Manage project-specific custom commands",
}

func init() {
	cmd.RootCmd.AddCommand(commandsCmd)
}
