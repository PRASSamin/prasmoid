/*
Copyright 2025 PRAS
*/
package regen

import (
	"github.com/spf13/cobra"
	root "github.com/PRASSamin/prasmoid/cmd"
)

func init() {
	root.RootCmd.AddCommand(regenCmd)
}

var regenCmd = &cobra.Command{
	Use:   "regen",
	Short: "Regenerate prasmoid files",
}
