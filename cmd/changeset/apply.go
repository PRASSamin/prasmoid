/*
Copyright 2025 PRAS
*/
package changeset

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/frontmatter"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	changesetCmd.AddCommand(changesetApplyCmd)
}

// ChangesetMeta represents the metadata for a changeset
type ChangesetMeta struct {
	ID   string `yaml:"id"`
	Bump string `yaml:"bump"`
	Next string `yaml:"next"`
	Date string `yaml:"date"`
}

// changesetApplyCmd represents the changesetApply command
var changesetApplyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply all .mdx changesets from the .changes directory",
	Run: func(cmd *cobra.Command, args []string) {
		ApplyChanges()
	},
}

var ApplyChanges = func() {
	changesetFiles := []string{}
	err := filepathWalk(".changes", func(path string, info os.FileInfo, err error) error {
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
		fmt.Println(color.RedString("Failed to walk changes directory: %v", err))
		return
	}

	if len(changesetFiles) == 0 {
		fmt.Println(color.YellowString("No changeset files found."))
		fmt.Println(color.CyanString("run `prasmoid changeset add` to create a changeset."))
		return
	}

	for _, file := range changesetFiles {
		data, err := osReadFile(file)
		if err != nil {
			fmt.Println(color.RedString("Failed to read changeset file %s: %v", file, err))
			continue
		}

		meta, body, err := matterParse(data)
		if err != nil {
			fmt.Println(color.RedString("Failed to parse %s: %v", file, err))
			continue
		}

		if err := utilsUpdateMetadata("Version", meta.Next); err != nil {
			fmt.Println(color.RedString("Metadata update failed in %s: %v", file, err))
			continue
		}

		if err := UpdateChangelog(meta.Next, meta.Date, body); err != nil {
			fmt.Println(color.RedString("Changelog update failed in %s: %v", file, err))
			continue
		}

		if err := osRemove(file); err != nil {
			fmt.Println(color.RedString("Failed to remove changeset file %s: %v", file, err))
			continue
		}
	}

	fmt.Println(color.GreenString("All changesets applied successfully!"))
}

var matterParse = func(data []byte) (ChangesetMeta, string, error) {
	var meta ChangesetMeta
	body, err := frontmatter.Parse(bytes.NewReader(data), &meta)
	if err != nil {
		return ChangesetMeta{}, "", err
	}

	return meta, string(body), nil
}

var UpdateChangelog = func(version, date, body string) error {
	const changelogPath = "CHANGELOG.md"

	newEntry := fmt.Sprintf("## [v%s] - %s\n\n%s\n\n", version, date, strings.TrimSpace(body))

	// Check if file exists
	if _, err := osStat(changelogPath); os.IsNotExist(err) {
		initial := "# CHANGELOG\n\n" + newEntry
		return osWriteFile(changelogPath, []byte(initial), 0644)
	}

	// Read the existing changelog
	data, err := osReadFile(changelogPath)
	if err != nil {
		return fmt.Errorf("failed to read changelog: %w", err)
	}

	content := string(data)

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
	if err := osWriteFile(changelogPath, []byte(updated), 0644); err != nil {
		return fmt.Errorf("failed to write changelog: %w", err)
	}

	return nil
}
