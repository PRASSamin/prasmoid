/*
Copyright 2025 PRAS
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

var CliConfig struct {
    Version string `yaml:"version"`
    Name    string `yaml:"name"`
    Author  string `yaml:"author"`
    License string `yaml:"license"`
    Github  string `yaml:"github"`
}

func init() {
	if err := loadConfig("config.yml"); err != nil {
		return
	}
    rootCmd.AddCommand(VersionCmd)
}

var VersionCmd = &cobra.Command{
    Use:   "version",
    Short: "Show Prasmoid version",
    Long:  "Show Prasmoid version.",
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println(CliConfig.Version)
    },
}

func loadConfig(path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return fmt.Errorf("failed to read config file: %w", err)
    }

    if err := yaml.Unmarshal(data, &CliConfig); err != nil {
        return fmt.Errorf("failed to unmarshal config: %w", err)
    }

    return nil
}