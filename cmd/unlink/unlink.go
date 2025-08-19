/*
Copyright © 2025 PRAS
*/
package unlink

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/PRASSamin/prasmoid/cmd"
)

func init() {
	cmd.RootCmd.AddCommand(UnlinkCmd)
}

// UnlinkCmd represents the unlink command
var UnlinkCmd = &cobra.Command{
	Use:   "unlink",
	Short: "Unlink development plasmoid from the system",
	Long:  "Remove the symlink linking the development plasmoid from the system directories.",
	Run: func(cmd *cobra.Command, args []string) {
		dest, err := utilsGetDevDest()

		if err != nil {
			fmt.Println(color.RedString(err.Error()))
			return
		}

		// Remove if exists
		_ = osRemoveAll(dest)
		fmt.Println(color.GreenString("Plasmoid unlinked successfully."))
	},
}
