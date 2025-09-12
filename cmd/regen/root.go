/*
Copyright 2025 PRAS
*/
package regen

import (
	root "github.com/PRASSamin/prasmoid/cmd"
	"github.com/spf13/cobra"
)

func init() {
	root.RootCmd.AddCommand(regenCmd)
}

var regenCmd = &cobra.Command{
	Use:   "regen",
	Short: "Regenerate prasmoid files",
}
