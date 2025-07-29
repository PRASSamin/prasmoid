package cmd

import (
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(regenCmd)
}

var regenCmd = &cobra.Command{
	Use:   "regen",
	Short: "Regenerate prasmoid files",
}