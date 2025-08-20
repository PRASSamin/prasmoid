/*
Copyright 2025 PRAS
*/
package cmd

import (
	"fmt"

	"github.com/PRASSamin/prasmoid/cmd"
	"github.com/PRASSamin/prasmoid/internal"
	"github.com/spf13/cobra"
)

var CliConfig struct {
	Version string `yaml:"version"`
	Name    string `yaml:"name"`
	Author  string `yaml:"author"`
	License string `yaml:"license"`
	Github  string `yaml:"github"`
}

func init() {
	cmd.RootCmd.AddCommand(VersionCmd)
}

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show Prasmoid version",
	Long:  "Show Prasmoid version.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(internal.AppMetaData.Version)
	},
}
