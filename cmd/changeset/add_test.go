package changeset

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	time "time"

	"github.com/PRASSamin/prasmoid/tests"
	"github.com/PRASSamin/prasmoid/utils"
)

func TestGetNextVersion(t *testing.T) {
	testCases := []struct {
		name           string
		currentVersion string
		bump           string
		expected       string
		expectError    bool
	}{
		{
			name:           "patch bump",
			currentVersion: "1.2.3",
			bump:           "patch",
			expected:       "1.2.4",
			expectError:    false,
		},
		{
			name:           "minor bump",
			currentVersion: "1.2.3",
			bump:           "minor",
			expected:       "1.3.0",
			expectError:    false,
		},
		{
			name:           "major bump",
			currentVersion: "1.2.3",
			bump:           "major",
			expected:       "2.0.0",
			expectError:    false,
		},
		{
			name:           "invalid version format",
			currentVersion: "1.2",
			bump:           "patch",
			expected:       "",
			expectError:    true,
		},
		{
			name:           "invalid bump type",
			currentVersion: "1.2.3",
			bump:           "invalid",
			expected:       "",
			expectError:    true,
		},
		{
			name:           "patch from zero",
			currentVersion: "0.0.0",
			bump:           "patch",
			expected:       "0.0.1",
			expectError:    false,
		},
		{
			name:           "minor with high patch",
			currentVersion: "1.9.15",
			bump:           "minor",
			expected:       "1.10.0",
			expectError:    false,
		},
		{
			name:           "invalid char in version",
			currentVersion: "1.9.x",
			bump:           "minor",
			expected:       "",
			expectError:    true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nextVersion, err := GetNextVersion(tc.currentVersion, tc.bump)

			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error but got: %v", err)
				}
				if nextVersion != tc.expected {
					t.Errorf("Expected version %s, but got %s", tc.expected, nextVersion)
				}
			}
		})
	}
}

func TestOpenEditor(t *testing.T) {
	createEditorScript := func(t *testing.T, content string) string {
		t.Helper()
		scriptPath := filepath.Join(t.TempDir(), "editor.sh")
		err := os.WriteFile(scriptPath, []byte(content), 0755)
		if err != nil {
			t.Fatalf("Failed to create temp script: %v", err)
		}
		return scriptPath
	}

	t.Run("successful edit with content", func(t *testing.T) {
		editorScript := createEditorScript(t, `#!/bin/sh
echo "This is a test changelog entry." > "$1"
`)
		t.Setenv("EDITOR", editorScript)

		content, err := OpenEditor()
		if err != nil {
			t.Fatalf("OpenEditor() failed: %v", err)
		}

		expected := "This is a test changelog entry."
		if content != expected {
			t.Errorf("Expected content %q, got %q", expected, content)
		}
	})

	t.Run("content with comments and empty lines is cleaned", func(t *testing.T) {
		scriptContent := `#!/bin/sh
cat > "$1" <<'EOF'
# This is a comment.
This is the first line.

This is the second line.

# Another comment.
EOF
`
		editorScript := createEditorScript(t, scriptContent)
		t.Setenv("EDITOR", editorScript)

		content, err := OpenEditor()
		if err != nil {
			t.Fatalf("OpenEditor() failed: %v", err)
		}

		expected := "This is the first line.\nThis is the second line."
		if content != expected {
			t.Errorf("Expected content %q, got %q", expected, content)
		}
	})

	t.Run("editor returns an error", func(t *testing.T) {
		editorScript := createEditorScript(t, `#!/bin/sh
exit 1
`)
		t.Setenv("EDITOR", editorScript)

		_, err := OpenEditor()
		if err == nil {
			t.Fatal("OpenEditor() should have failed but did not")
		}
	})

	t.Run("empty content after cleaning", func(t *testing.T) {
		scriptContent := `#!/bin/sh
cat > "$1" <<'EOF'
# A comment
# Another comment

EOF
`
		editorScript := createEditorScript(t, scriptContent)
		t.Setenv("EDITOR", editorScript)

		content, err := OpenEditor()
		if err != nil {
			t.Fatalf("OpenEditor() failed: %v", err)
		}

		if content != "" {
			t.Errorf("Expected empty content, got %q", content)
		}
	})

	t.Run("default editor is used when EDITOR is not set", func(t *testing.T) {
		t.Setenv("EDITOR", "") // Unset EDITOR for this test

		tmpDir := t.TempDir()
		fakeNanoPath := filepath.Join(tmpDir, "nano")
		scriptContent := `#!/bin/sh
echo "content from fake nano" > "$1"
`
		if err := os.WriteFile(fakeNanoPath, []byte(scriptContent), 0755); err != nil {
			t.Fatalf("Failed to create fake nano script: %v", err)
		}

		t.Setenv("PATH", tmpDir+string(filepath.ListSeparator)+os.Getenv("PATH"))

		content, err := OpenEditor()
		if err != nil {
			t.Fatalf("OpenEditor() failed: %v", err)
		}

		expected := "content from fake nano"
		if content != expected {
			t.Errorf("Expected content %q, got %q", expected, content)
		}
	})

	t.Run("temp file creation fails", func(t *testing.T) {
		// Set TMPDIR to a non-existent directory to cause os.CreateTemp to fail.
		t.Setenv("TMPDIR", filepath.Join(t.TempDir(), "nonexistent"))

		_, err := OpenEditor()
		if err == nil {
			t.Fatal("OpenEditor() should have failed but did not")
		}
	})
}


func TestAddChangeset(t *testing.T) {
	normalize := func(s string) string {
		return strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", "\n"), "\n", ""))
	}

	expected := func(summary string) string {
		return normalize(`---
id: org.kde.testplasmoid
bump: patch
next: 1.0.1
date: ` + time.Now().Format("2006-01-02") + `
---

		` + summary)
}

	t.Run("successful changeset creation with prompts and apply", func(t *testing.T) {
		tmpDir, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		summary := "This is a test summary."
		err := AddChangeset("patch", summary, true)
		if err != nil {
			t.Errorf("AddChangeset() returned an unexpected error: %v", err)
		}

		// Verify apply changes
		version, err := utils.GetDataFromMetadata("Version")
		if err != nil {
			t.Fatal("Failed to get metadata")
		}
		if version != "1.0.1" {
			t.Errorf("Expected version 1.0.1, got %s", version)
		}
		
		// verify summary in changelog
		changelogPath := filepath.Join(tmpDir, "CHANGELOG.md")
		content, err := os.ReadFile(changelogPath)
		if err != nil {
			t.Fatalf("Failed to read changelog: %v", err)
		}
		actual := normalize(string(content))
		if !strings.Contains(actual, summary) {
			t.Errorf("Summary not found in changelog.\nExpected:\n%q\n\nGot:\n%q", summary, actual)
		}
	})

	t.Run("successful changeset creation with flags", func(t *testing.T) {
		tmpDir, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		_ = changesetAddCmd.Flags().Set("bump", "patch")
		_ = changesetAddCmd.Flags().Set("summary", "Summary from flag")
		changesetAddCmd.Run(changesetAddCmd, []string{})

		// Verify changeset file was created
		changesDir := filepath.Join(tmpDir, ChangesFolder)
		files, err := os.ReadDir(changesDir)
		if err != nil {
			t.Fatalf("Failed to read changes directory: %v", err)
		}
		if len(files) != 1 {
			t.Errorf("Expected 1 changeset file, got %d", len(files))
		}

		// Verify content of the changeset file
		changesetFilePath := filepath.Join(changesDir, files[0].Name())
		content, err := os.ReadFile(changesetFilePath)
		if err != nil {
			t.Fatalf("Failed to read changeset file: %v", err)
		}

		if !strings.Contains(normalize(string(content)), expected("Summary from flag")) {
			t.Errorf("Changeset content mismatch.\nExpected part: %s\nActual content:\n%s", expected("Summary from flag"), normalize(string(content)))
		}
	})
	
	t.Run("invalid plasmoid", func(t *testing.T) {
		tmpDir, cleanup := tests.SetupTestProject(t)
		_ = os.Remove(filepath.Join(tmpDir, "metadata.json"))
		defer cleanup()

		summary := "This is a test summary."
		err := AddChangeset("patch", summary, false)
		if err == nil {
			t.Error("AddChangeset() expected an error for invalid plasmoid, but got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "current directory is not a valid plasmoid") {
			t.Errorf("AddChangeset() expected 'current directory is not a valid plasmoid', but got %v", err)
		}
	})

	t.Run("empty changelog after editor and flag", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		err := AddChangeset("patch", "", false)
		if err == nil {
			t.Error("AddChangeset() expected an error for empty changelog, but got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "failed to prompt for changelog summary") {
			t.Errorf("AddChangeset() expected 'failed to prompt for changelog summary' error, got: %v", err)
		}
	})

	t.Run("failed to create changes directory", func(t *testing.T) {
		tmpDir, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		// Make the temp directory read-only to simulate MkdirAll failure
		if err := os.Chmod(tmpDir, 0555); err != nil {
			t.Fatalf("Failed to make temp dir read-only: %v", err)
		}
		defer func() { _ = os.Chmod(tmpDir, 0755) }() // Restore permissions for cleanup

		err := AddChangeset("patch", "Summary for failing mkdir", false)
		if err == nil {
			t.Error("AddChangeset() expected an error for failed directory creation, but got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "failed to create changes directory") {
			t.Errorf("AddChangeset() expected 'failed to create changes directory' error, got: %v", err)
		}
	})

	t.Run("failed to write changeset file", func(t *testing.T) {
		tmpDir, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		// Create .changes folder and make it read-only
		changesFolder := filepath.Join(tmpDir, ChangesFolder)
		if err := os.MkdirAll(changesFolder, 0755); err != nil {
			t.Fatalf("Failed to create changes folder: %v", err)
		}
		if err := os.Chmod(changesFolder, 0555); err != nil {
			t.Fatalf("Failed to make changes folder read-only: %v", err)
		}
		defer func() { _ = os.Chmod(changesFolder, 0755) }() // Restore permissions for cleanup

		err := AddChangeset("patch", "Summary for failing write", false)
		if err == nil {
			t.Error("AddChangeset() expected an error for failed file write, but got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "failed to write changeset") {
			t.Errorf("AddChangeset() expected 'failed to write changeset' error, got: %v", err)
		}
	})

	t.Run("invalid metadata file", func(t *testing.T) {
		tmpDir, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		// Corrupt metadata.json to fail
		metadataContent := `{
			"KPlugin": {
				nothing: {}
			}
		}`
		metadataPath := filepath.Join(tmpDir, "metadata.json")
		if err := os.WriteFile(metadataPath, []byte(metadataContent), 0644); err != nil {
			t.Fatalf("Failed to create dummy metadata.json: %v", err)
		}

		err := AddChangeset("patch", "Summary for invalid metadata", false)
		if err == nil {
			t.Error("AddChangeset() expected an error for invalid metadata, but got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "invalid metadata") {
			t.Errorf("AddChangeset() expected 'invalid metadata' error, got: %v", err)
		}
	})
	
	t.Run("failed to compute next version with empty dump", func(t *testing.T) {
		tmpDir, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		// Corrupt metadata.json to make GetNextVersion fail
		metadataContent := `{
			"KPlugin": {
				"Version": "invalid",
				"Id": "invalid"
			}
		}`
		metadataPath := filepath.Join(tmpDir, "metadata.json")
		if err := os.WriteFile(metadataPath, []byte(metadataContent), 0644); err != nil {
			t.Fatalf("Failed to create dummy metadata.json: %v", err)
		}

		err := AddChangeset("", "Summary for invalid version format", false)
		if err == nil {
			t.Error("AddChangeset() expected an error for invalid metadata, but got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "failed to compute next version") {
			t.Errorf("AddChangeset() expected 'failed to compute next version' error, got: %v", err)
		}
	})
	
	t.Run("failed to compute next version with dump", func(t *testing.T) {
		tmpDir, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		// Corrupt metadata.json to make GetNextVersion fail
		metadataContent := `{
			"KPlugin": {
				"Version": "invalid",
				"Id": "invalid"
			}
		}`
		metadataPath := filepath.Join(tmpDir, "metadata.json")
		if err := os.WriteFile(metadataPath, []byte(metadataContent), 0644); err != nil {
			t.Fatalf("Failed to create dummy metadata.json: %v", err)
		}

		err := AddChangeset("patch", "Summary for invalid version format", false)
		if err == nil {
			t.Error("AddChangeset() expected an error for invalid metadata, but got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "failed to compute next version") {
			t.Errorf("AddChangeset() expected 'failed to compute next version' error, got: %v", err)
		}
	})

	t.Run("failed to prompt for version bump", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		err := AddChangeset("", "Summary for failed prompt", false)
		if err == nil {
			t.Error("AddChangeset() expected an error for failed prompt, but got nil")
		}
		if err != nil && !strings.Contains(err.Error(), "failed to prompt for version bump") {
			t.Errorf("AddChangeset() expected 'failed to prompt for version bump' error, got: %v", err)
		}
	})
}