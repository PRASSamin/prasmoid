/*
Copyright © 2025 PRAS
*/
package cmd

import (
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/PRASSamin/prasmoid/utils"
)

func init() {
	rootCmd.AddCommand(UnlinkCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// UnlinkCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// UnlinkCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// UnlinkCmd represents the unlink command
var UnlinkCmd = &cobra.Command{
	Use:   "unlink",
	Short: "Unlink development plasmoid from the system",
	Long:  "Remove the symlink linking the development plasmoid from the system directories.",
	Run: func(cmd *cobra.Command, args []string) {
		dest, err := utils.GetDevDest()

		if err != nil {
			color.Red(err.Error())
			return
		}

		// Remove if exists
		_ = os.Remove(dest)
		_ = os.RemoveAll(dest)
		color.Green("Plasmoid unlinked successfully.")
	},
}
