/*
Copyright 2025 PRAS
*/
package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/PRASSamin/prasmoid/utils"
	"github.com/adrg/frontmatter"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	ChangesetRootCmd.AddCommand(ChangesetApplyCmd)
}

// ChangesetMeta represents the metadata for a changeset
type ChangesetMeta struct {
	ID   string `yaml:"id"`
	Bump string `yaml:"bump"`
	Next string `yaml:"next"`
	Date string `yaml:"date"`
}

// changesetApplyCmd represents the changesetApply command
var ChangesetApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply all .mdx changesets from the .changes directory",
	Run: func(cmd *cobra.Command, args []string) {
		ApplyChanges()
	},
}

func ApplyChanges() {
	changesetFiles := []string{}
	err := filepath.Walk(".changes", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".mdx" {
			changesetFiles = append(changesetFiles, path)
		}
		return nil
	})
	if err != nil {
		color.Red("Failed to walk changes directory: %v", err)
		return
	}

	if len(changesetFiles) == 0 {
		color.Yellow("No changeset files found.")
		color.Cyan("run `prasmoid changeset add` to create a changeset.")
		return
	}

	for _, file := range changesetFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			color.Red("Failed to read changeset file %s: %v", file, err)
			continue
		}

		meta, body, err := matterParse(data)
		if err != nil {
			color.Red("Failed to parse %s: %v", file, err)
			continue
		}

		if err := utils.UpdateMetadata("Version", meta.Next); err != nil {
			color.Red("Metadata update failed in %s: %v", file, err)
			continue
		}

		if err := UpdateChangelog(meta.Next, meta.Date, body); err != nil {
			color.Red("Changelog update failed in %s: %v", file, err)
			continue
		}

		if err := os.Remove(file); err != nil {
			color.Red("Failed to remove changeset file %s: %v", file, err)
			continue
		}
	}

	color.Green("All changesets applied successfully!")
}

func matterParse(data []byte) (ChangesetMeta, string, error) {
	var meta ChangesetMeta
	body, err := frontmatter.Parse(bytes.NewReader(data), &meta)
	if err != nil {
		return ChangesetMeta{}, "", err
	}

	return meta, string(body), nil
}

func UpdateChangelog(version, date, body string) error {
	const changelogPath = "CHANGELOG.md"

	newEntry := fmt.Sprintf("## [v%s] - %s\n\n%s\n\n", version, date, strings.TrimSpace(body))

	// Check if file exists
	if _, err := os.Stat(changelogPath); os.IsNotExist(err) {
		initial := "# CHANGELOG\n\n" + newEntry
		return os.WriteFile(changelogPath, []byte(initial), 0644)
	}

	// Read the existing changelog
	data, err := os.ReadFile(changelogPath)
	if err != nil {
		return fmt.Errorf("failed to read changelog: %w", err)
	}

	content := string(data)

	// Ensure there's a top-level header
	if !strings.HasPrefix(content, "# CHANGELOG") {
		content = "# CHANGELOG\n\n" + content
	}

	// Split the content: keep the header, and append the new entry below
	parts := strings.SplitN(content, "\n", 2)
	header := parts[0]
	rest := ""
	if len(parts) > 1 {
		rest = parts[1]
	}

	// Rebuild
	updated := header + "\n\n" + newEntry + strings.TrimPrefix(rest, "\n")

	// Save it back
	if err := os.WriteFile(changelogPath, []byte(updated), 0644); err != nil {
		return fmt.Errorf("failed to write changelog: %w", err)
	}

	return nil
}
