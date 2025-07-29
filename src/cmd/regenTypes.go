package cmd

import (
	"github.com/PRASSamin/prasmoid/consts"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	regenCmd.AddCommand(regenTypesCmd)
}

var regenTypesCmd = &cobra.Command{
	Use:   "types",
	Short: "Regenerate prasmoid.d.ts",
	Run: func(cmd *cobra.Command, args []string) {
		if err := createFileFromTemplate("prasmoid.d.ts", consts.PRASMOID_DTS); err != nil {
			color.Red("Failed to regenerate prasmoid.d.ts:", err)
			return
		}
		color.Green("prasmoid.d.ts regenerated successfully.")
	},
}