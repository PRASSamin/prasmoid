package cmd_tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/PRASSamin/prasmoid/cmd"
)

func TestUpdateChangelog(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "changelog-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	changelogPath := filepath.Join(tmpDir, "CHANGELOG.md")

	// Test case 1: Create a new changelog
	t.Run("create new changelog", func(t *testing.T) {
		err := updateChangelogInPath(changelogPath, "1.0.0", "2025-01-01", "Initial release")
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}

		data, err := os.ReadFile(changelogPath)
		if err != nil {
			t.Fatalf("Failed to read changelog: %v", err)
		}

		expected := "# CHANGELOG\n\n## [v1.0.0] - 2025-01-01\n\nInitial release\n\n"
		if string(data) != expected {
			t.Errorf("Expected changelog content %q, but got %q", expected, string(data))
		}
	})

	// Test case 2: Append to an existing changelog
	t.Run("append to existing changelog", func(t *testing.T) {
		err := updateChangelogInPath(changelogPath, "1.1.0", "2025-01-02", "Added new feature")
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}

		data, err := os.ReadFile(changelogPath)
		if err != nil {
			t.Fatalf("Failed to read changelog: %v", err)
		}

		expected := "# CHANGELOG\n\n## [v1.1.0] - 2025-01-02\n\nAdded new feature\n\n## [v1.0.0] - 2025-01-01\n\nInitial release\n\n"
		if string(data) != expected {
			t.Errorf("Expected changelog content %q, but got %q", expected, string(data))
		}
	})
}

// Wrapper to allow testing with a specific path
func updateChangelogInPath(path, version, date, body string) error {
	// Temporarily change the working directory for the test
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(filepath.Dir(path))

	return cmd.UpdateChangelog(version, date, body)
}
