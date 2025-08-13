/*
Copyright © 2025 PRAS
*/
package changeset

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/AlecAivazis/survey/v2"
	"github.com/PRASSamin/prasmoid/utils"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var ChangesFolder string = ".changes"

var validBumps = map[string]bool{
	"patch": true,
	"minor": true,
	"major": true,
}

func init() {
	changesetAddCmd.Flags().StringP("bump", "b", "", "Version bump type (patch|minor|major)")
	changesetAddCmd.Flags().StringP("summary", "s", "", "Changelog summary (optional)")
	changesetAddCmd.Flags().BoolP("apply", "a", false, "Apply changeset after creation")
	changesetCmd.AddCommand(changesetAddCmd)
}

var changesetAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Create a new changeset",
	Run: func(cmd *cobra.Command, args []string) {
		bump, _ := cmd.Flags().GetString("bump")
		summary, _ := cmd.Flags().GetString("summary")
		apply, _ := cmd.Flags().GetBool("apply")
		_ = AddChangeset(bump, summary, apply)
	},
}

func AddChangeset(bump string, summary string, apply bool) error {
	if !utils.IsValidPlasmoid() {
		return fmt.Errorf("current directory is not a valid plasmoid")
	}
	version, verr := utils.GetDataFromMetadata("Version")
	id, ierr := utils.GetDataFromMetadata("Id")
	if verr != nil || ierr != nil {
		return fmt.Errorf("invalid metadata: %v", fmt.Sprintf("%v or %v", ierr, verr))
	}
	var next string

	// Handle bump: flag or prompt
	if bump == "" || !validBumps[bump] {
		bumpLabels := make(map[string]string)
		for _, bumpType := range []string{"patch", "minor", "major"} {
			if nextVer, err := GetNextVersion(version.(string), bumpType); err == nil {
				bumpLabels[bumpType] = fmt.Sprintf("%s (%s)", bumpType, nextVer)
			} else {
				color.Red("Failed to compute next version for %s: %v", bumpType, err)
				return fmt.Errorf("failed to compute next version for %s: %v", bumpType, err)
			}
		}
	
		options := []string{bumpLabels["patch"], bumpLabels["minor"], bumpLabels["major"]}

		var selected string
		err := survey.AskOne(&survey.Select{
			Message: "Select version bump:",
			Options: options,
			Default: bumpLabels["patch"],
		}, &selected, survey.WithValidator(survey.Required))

		if err != nil {
			color.Red("Failed to prompt for version bump: %v", err)
			return fmt.Errorf("failed to prompt for version bump: %v", err)
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
			return fmt.Errorf("failed to compute next version for %s: %v", bump, err)
		}
	}

	// Handle summary: flag or editor prompt
	if summary == "" {
		var err error
		summary, err = OpenEditor()
		if err != nil || strings.TrimSpace(summary) == "" {
			// Fallback to inline prompt if editor fails or empty
			prompt := &survey.Input{
				Message: "Enter changelog summary:",
			}
			err = survey.AskOne(prompt, &summary, survey.WithValidator(survey.Required))

			if err != nil {
				color.Red("Failed to prompt for changelog summary: %v", err)
				return fmt.Errorf("failed to prompt for changelog summary: %v", err)
			}
		}
	}

	// Create changes dir if not exist
	if err := os.MkdirAll(ChangesFolder, 0755); err != nil {
		color.Red("Failed to create changes directory: %v", err)
		return fmt.Errorf("failed to create changes directory: %v", err)
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
		return fmt.Errorf("failed to write changeset: %v", err)
	}

	if apply {
		_ = ApplyChanges()
	}
	return nil
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

func OpenEditor() (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "nano"
	}

	tmpFile, err := os.CreateTemp("", "changeset-*.mdx")
	if err != nil {
		return "", err
	}
	defer func() {
		if err := os.Remove(tmpFile.Name()); err != nil {
			log.Printf("Error removing temporary file: %v", err)
		}
	}()

	if _, err := tmpFile.WriteString(`# Write your changelog entry below.
# Lines starting with # will be ignored.
# Example: Added system tray widget for memory usage.
`); err != nil {
		return "", err
	}
	if err := tmpFile.Close(); err != nil {
		return "", err
	}

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

