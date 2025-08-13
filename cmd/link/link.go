/*
Copyright 2025 PRAS
*/
package link

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/PRASSamin/prasmoid/cmd"
	"github.com/PRASSamin/prasmoid/utils"
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
		if !utils.IsValidPlasmoid() {
			color.Red("Current directory is not a valid plasmoid.")
			return
		}
		dest, err := utils.GetDevDest()
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

func LinkPlasmoid(dest string) error {
	// Remove if exists
	_ = os.Remove(dest)
	_ = os.RemoveAll(dest)

	// retrive current dir
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	// Link
	return os.Symlink(cwd, dest)
}
