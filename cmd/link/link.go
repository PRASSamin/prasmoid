/*
Copyright 2025 PRAS
*/
package link

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/PRASSamin/prasmoid/cmd"
)

var linkWhere bool

func init() {
	LinkCmd.Flags().BoolVarP(&linkWhere, "where", "w", false, "show where the plasmoid is linked.")
	cmd.RootCmd.AddCommand(LinkCmd)
}

// devCmd represents the dev command
var LinkCmd = &cobra.Command{
	Use:   "link",
	Short: "Link plasmoid to local development directory",
	Long:  "Create a symlink to local development folder for easy testing.",
	Run: func(cmd *cobra.Command, args []string) {
		if !utilsIsValidPlasmoid() {
			color.Red("Current directory is not a valid plasmoid.")
			return
		}
		dest, err := utilsGetDevDest()
		if err != nil {
			color.Red(err.Error())
			return
		}
		if linkWhere {
			fmt.Println("Plasmoid linked to:\n", "- ", color.BlueString(dest))
			return
		}
		if err := LinkPlasmoid(dest); err != nil {
			color.Red("Failed to link plasmoid:", err)
			return
		}
		color.Green("Plasmoid linked successfully.")
	},
}

var LinkPlasmoid = func(dest string) error {
	// Remove if exists
	_ = osRemoveAll(dest)

	cwd, err := osGetwd()
	if err != nil {
		return err
	}

	return osSymlink(cwd, dest)
}
