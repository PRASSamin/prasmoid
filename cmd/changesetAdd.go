/*
Copyright © 2025 PRAS
*/
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/PRASSamin/prasmoid/utils"
)

var bump string
var summary string
var ChangesFolder string = ".changes"

var validBumps = map[string]bool{
	"patch": true,
	"minor": true,
	"major": true,
}


func init() {
	ChangesetAddCmd.Flags().StringVarP(&bump, "bump", "b", "", "Version bump type (patch|minor|major)")
	ChangesetAddCmd.Flags().StringVarP(&summary, "summary", "s", "", "Changelog summary (optional)")
	ChangesetRootCmd.AddCommand(ChangesetAddCmd)
}


var ChangesetAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Create a new changeset",
	Run: func(cmd *cobra.Command, args []string) {
		bump = strings.ToLower(bump)
		version, _ := utils.GetDataFromMetadata("Version")
		id, _ := utils.GetDataFromMetadata("Id")
		var next string;

		bumpLabels := make(map[string]string)
		for _, bumpType := range []string{"patch", "minor", "major"} {
			if nextVer, err := GetNextVersion(version.(string), bumpType); err == nil {
				bumpLabels[bumpType] = fmt.Sprintf("%s (%s)", bumpType, nextVer)
			} else {
				color.Red("Failed to compute next version for %s: %v", bumpType, err)
				return
			}
		}

		options := []string{bumpLabels["patch"], bumpLabels["minor"], bumpLabels["major"]}
		
		// Handle bump: flag or prompt
		if !validBumps[bump] {
			var selected string
			err := survey.AskOne(&survey.Select{
				Message: "Select version bump:",
				Options: options,
				Default: bumpLabels["patch"],
			}, &selected)

			if err != nil {
				color.Red("Failed to prompt for version bump: %v", err)
				return
			}
		
			// Split once and trim parens cleanly
			parts := strings.SplitN(selected, " ", 2)
			bump = parts[0]
		}

		// Handle next version: flag or compute
		if next == "" {
			var err error
			next, err = GetNextVersion(version.(string), bump)
			if err != nil {
				color.Red("Failed to compute next version for %s: %v", bump, err)
				return
			}
		}
		
				

		// Handle summary: flag or editor prompt
		if summary == "" {
			var err error
			summary, err = openEditor()
			if err != nil || strings.TrimSpace(summary) == "" {
				// Fallback to inline prompt if editor fails or empty
				prompt := &survey.Input{
					Message: "Enter changelog summary:",
				}
				err = survey.AskOne(prompt, &summary)

				if err != nil {
					color.Red("Failed to prompt for changelog summary: %v", err)
					return
				}
			}
		}

		// Validate summary
		if strings.TrimSpace(summary) == "" {
			color.Red("Empty changelog. Aborting.")
			return
		}

		// Create changes dir if not exist
		if err := os.MkdirAll(ChangesFolder, 0755); err != nil {
			color.Red("Failed to create changes directory: %v", err)
			return
		}

		// Generate filename & content
		now := time.Now()
		timestamp := now.Format("2006-01-02-15-04")
		dateOnly := now.Format("2006-01-02")
		filename := filepath.Join(ChangesFolder, fmt.Sprintf("%s-%s.mdx", timestamp, bump))

		content := fmt.Sprintf(`---
id: %s
bump: %s
next: %s
date: %s
---		

%s
`, id, bump, next, dateOnly, summary)

		// Write file
		if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
			color.Red("Failed to write changeset: %v", err)
			return
		}

		var applyChanges bool
		applyConfirmPrompt := &survey.Confirm{
			Message: color.YellowString("Changeset created. Would you like to apply it now?"),
			Default: false,
		}
		if err := survey.AskOne(applyConfirmPrompt, &applyChanges); err != nil {
			return
		}
		if applyChanges {
			ApplyChanges()
		}
	},
}

func GetNextVersion(version string, bump string) (string, error) {
	versionParts := strings.Split(version, ".")
	if len(versionParts) != 3 {
		return "", fmt.Errorf("invalid version format: %s, expected format: major.minor.patch", version)
	}

	major, err1 := strconv.Atoi(versionParts[0])
	minor, err2 := strconv.Atoi(versionParts[1])
	patch, err3 := strconv.Atoi(versionParts[2])
	if err1 != nil || err2 != nil || err3 != nil {
		return "", fmt.Errorf("version parts must be integers: %v %v %v", err1, err2, err3)
	}

	switch bump {
	case "patch":
		patch++
	case "minor":
		minor++
		patch = 0
	case "major":
		major++
		minor = 0
		patch = 0
	default:
		return "", fmt.Errorf("invalid bump type: %s, must be 'major', 'minor' or 'patch'", bump)
	}

	return fmt.Sprintf("%d.%d.%d", major, minor, patch), nil
}


func openEditor() (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano"
	}

	tmpFile, err := os.CreateTemp("", "changeset-*.mdx")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFile.Name())

	tmpFile.WriteString(`# Write your changelog entry below.
# Lines starting with # will be ignored.
# Example: Added system tray widget for memory usage.
`)
	tmpFile.Close()

	cmd := exec.Command(editor, tmpFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", err
	}

	data, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(data), "\n")
	var cleaned []string
	for _, line := range lines {
		if !strings.HasPrefix(line, "#") && strings.TrimSpace(line) != "" {
			cleaned = append(cleaned, line)
		}
	}

	return strings.Join(cleaned, "\n"), nil
}
