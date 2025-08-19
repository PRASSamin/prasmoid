package changeset

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/PRASSamin/prasmoid/tests"
	"github.com/PRASSamin/prasmoid/utils"
)

func TestUpdateChangelog(t *testing.T) {
	tmpDir := t.TempDir()
	changelogPath := filepath.Join(tmpDir, "CHANGELOG.md")

	// Wrapper to allow testing with a specific path
	updateChangelogInPath := func(t *testing.T, path, version, date, body string) error {
		// Temporarily change the working directory for the test
		originalWd, _ := os.Getwd()
		defer func() {
			if err := os.Chdir(originalWd); err != nil {
				t.Errorf("Failed to restore original directory: %v", err)
			}
		}()

		if err := os.Chdir(filepath.Dir(path)); err != nil {
			return err
		}

		return UpdateChangelog(version, date, body)
	}

	t.Run("create new changelog", func(t *testing.T) {
		// ensure file doesn't exist
		_ = os.Remove(changelogPath)
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

	t.Run("append to existing changelog without header", func(t *testing.T) {
		// First, create a file without the header
		_ = os.WriteFile(changelogPath, []byte("Some initial content.\n"), 0644)

		err := updateChangelogInPath(t, changelogPath, "1.2.0", "2025-01-03", "Added another feature")
		if err != nil {
			t.Errorf("Expected no error, but got %v", err)
		}

		data, err := os.ReadFile(changelogPath)
		if err != nil {
			t.Fatalf("Failed to read changelog: %v", err)
		}

		expected := "# CHANGELOG\n\n## [v1.2.0] - 2025-01-03\n\nAdded another feature\n\nSome initial content.\n"
		if string(data) != expected {
			t.Errorf("Expected changelog content %q, but got %q", expected, string(data))
		}
	})

	t.Run("fail to update changelog", func(t *testing.T) {
		// Create the file as read-only for all users
		if err := os.WriteFile(changelogPath, []byte(""), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		if err := os.Chmod(changelogPath, 0444); err != nil {
			t.Fatalf("Failed to set file as read-only: %v", err)
		}
		defer func() { _ = os.Chmod(changelogPath, 0644) }()

		err := updateChangelogInPath(t, changelogPath, "1.2.0", "2025-01-03", "Added another feature")
		if err == nil {
			t.Fatal("UpdateChangelog() expected error, but got none")
		}
		if !strings.Contains(err.Error(), "permission denied") {
			t.Errorf("UpdateChangelog() expected permission error, but got %q", err.Error())
		}
	})
}

func TestApplyChanges(t *testing.T) {
	t.Run("successful apply", func(t *testing.T) {
		// Setup
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		changesDir := ".changes"
		_ = os.Mkdir(changesDir, 0755)

		// Create changeset files
		cs1 := "---\nid: org.kde.testplasmoid\nbump: patch\nnext: 1.0.1\ndate: 2025-01-01\n---\n- Added a new feature."

		cs2 := "---\nid: org.kde.testplasmoid\nbump: minor\nnext: 1.1.0\ndate: 2025-01-02\n---\n- Fixed a bug."

		// write files in reverse alphabetical order to test sorting
		_ = os.WriteFile(filepath.Join(changesDir, "z_change2.mdx"), []byte(cs2), 0644)
		_ = os.WriteFile(filepath.Join(changesDir, "a_change1.mdx"), []byte(cs1), 0644)

		// Run Apply
		changesetApplyCmd.Run(nil, []string{})

		// Verify version in metadata.json
		version, err := utils.GetDataFromMetadata("Version")
		fmt.Println("version", version)
		if err != nil {
			t.Fatal("Failed to get metadata")
		}
		if version != "1.1.0" {
			t.Errorf("expected version 1.1.0, got %s", version)
		}

		// Verify CHANGELOG.md
		changelog, _ := os.ReadFile("CHANGELOG.md")
		if !strings.Contains(string(changelog), "Added a new feature") {
			t.Error("changelog does not contain first change")
		}
		if !strings.Contains(string(changelog), "Fixed a bug") {
			t.Error("changelog does not contain second change")
		}

		// Verify changeset files are removed
		files, _ := os.ReadDir(changesDir)
		if len(files) != 0 {
			t.Errorf("expected .changes dir to be empty, but it has %d files", len(files))
		}
	})

	t.Run("no changeset files", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()
		_ = os.Mkdir(".changes", 0755)

		err := ApplyChanges()
		if err == nil {
			t.Fatal("expected an error but got none")
		}
		if !strings.Contains(err.Error(), "no changeset files found") {
			t.Errorf("expected 'no changeset files found' error, got: %v", err)
		}
	})

	t.Run("changeset dir not found", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		err := ApplyChanges()
		if err == nil {
			t.Fatal("expected an error but got none")
		}
		if !strings.Contains(err.Error(), "failed to walk changes directory") {
			t.Errorf("expected 'failed to walk changes directory' error, got: %v", err)
		}
	})

	t.Run("continue on parsing error", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		defer cleanup()

		changesDir := ".changes"
		_ = os.Mkdir(changesDir, 0755)

		cs_invalid := "---\ninvalid-yaml: :\n---\n- This one is invalid."
		cs_valid := "---\nid: org.kde.testplasmoid\nbump: minor\nnext: 1.1.0\ndate: 2025-01-02\n---\n- This one is valid."
		_ = os.WriteFile(filepath.Join(changesDir, "invalid.mdx"), []byte(cs_invalid), 0644)
		_ = os.WriteFile(filepath.Join(changesDir, "valid.mdx"), []byte(cs_valid), 0644)

		err := ApplyChanges()
		if err != nil {
			t.Fatalf("ApplyChanges() failed: %v", err)
		}

		files, _ := os.ReadDir(changesDir)
		if len(files) != 1 {
			t.Errorf("expected .changes dir to have 1 file, but it has %d files", len(files))
		}
		if files[0].Name() != "invalid.mdx" {
			t.Errorf("expected invalid.mdx to remain, but found %s", files[0].Name())
		}
	})

	varifyContinue := func(t *testing.T) {
		changesDir := ".changes"
		_ = os.Mkdir(changesDir, 0755)

		cs_valid := "---\nid: org.kde.testplasmoid\nbump: minor\nnext: 1.1.0\ndate: 2025-01-02\n---\n- This one is valid."
		_ = os.WriteFile(filepath.Join(changesDir, "valid.mdx"), []byte(cs_valid), 0644)

		err := ApplyChanges()
		if err != nil {
			t.Fatalf("ApplyChanges() failed: %v", err)
		}

		files, _ := os.ReadDir(changesDir)
		if len(files) != 1 {
			t.Errorf("expected .changes dir to have 1 file, but it has %d files", len(files))
		}
		if files[0].Name() != "valid.mdx" {
			t.Errorf("expected valid.mdx to remain, but found %s", files[0].Name())
		}
	}

	t.Run("continue on metadata update failure", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		_ = os.Chmod("metadata.json", 0444)
		defer func() { _ = os.Chmod("metadata.json", 0644) }()
		defer cleanup()

		varifyContinue(t)
	})
	
	t.Run("continue on changelog update failure", func(t *testing.T) {
		_, cleanup := tests.SetupTestProject(t)
		if err := os.WriteFile("CHANGELOG.md", []byte(""), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		if err := os.Chmod("CHANGELOG.md", 0444); err != nil {
			t.Fatalf("Failed to set file as read-only: %v", err)
		}
		defer cleanup()
		defer func() { _ = os.Chmod("CHANGELOG.md", 0644) }()

		varifyContinue(t)
	})
}

func TestMatterParse(t *testing.T) {
	t.Run("valid frontmatter", func(t *testing.T) {
		data := []byte(`---
id: test
bump: patch
next: 1.0.1
date: 2025-01-01
---
This is the body.`)
		meta, body, err := matterParse(data)
		if err != nil {
			t.Fatalf("matterParse failed: %v", err)
		}
		if meta.ID != "test" || meta.Bump != "patch" || meta.Next != "1.0.1" || meta.Date != "2025-01-01" {
			t.Errorf("parsed metadata is incorrect: %+v", meta)
		}
		if strings.TrimSpace(body) != "This is the body." {
			t.Errorf("parsed body is incorrect: %q", body)
		}
	})

	t.Run("invalid yaml", func(t *testing.T) {
		data := []byte(`---
id: test
invalid-yaml: :
---
This is the body.`)
		_, _, err := matterParse(data)
		if err == nil {
			t.Fatal("expected an error but got none")
		}
	})

	t.Run("no frontmatter", func(t *testing.T) {
		data := []byte(`This is just a body.`)
		meta, body, err := matterParse(data)
		if err != nil {
			t.Fatalf("matterParse failed: %v", err)
		}
		if (meta != ChangesetMeta{}) {
			t.Errorf("expected empty metadata, got %+v", meta)
		}
		if strings.TrimSpace(body) != "This is just a body." {
			t.Errorf("parsed body is incorrect: %q", body)
		}
	})

	t.Run("empty input", func(t *testing.T) {
		data := []byte(``)
		_, _, err := matterParse(data)
		if err != nil {
			t.Fatalf("expected an error for empty input but got none %v", err)
		}
	})
}