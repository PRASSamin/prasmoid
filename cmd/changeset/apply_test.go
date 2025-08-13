package changeset

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUpdateChangelog(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "changelog-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("Failed to remove temporary directory: %v", err)
		}
	}()

	changelogPath := filepath.Join(tmpDir, "CHANGELOG.md")

	// Test case 1: Create a new changelog
	t.Run("create new changelog", func(t *testing.T) {
		err := updateChangelogInPath(t, changelogPath, "1.0.0", "2025-01-01", "Initial release")
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
		err := updateChangelogInPath(t, changelogPath, "1.1.0", "2025-01-02", "Added new feature")
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
func updateChangelogInPath(t *testing.T, path, version, date, body string) error {
	// Temporarily change the working directory for the test
	originalWd, _ := os.Getwd()
	defer func(t *testing.T) {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("Failed to restore original directory: %v", err)
		}
	}(t)

	if err := os.Chdir(filepath.Dir(path)); err != nil {
		return err
	}

	return UpdateChangelog(version, date, body)
}
