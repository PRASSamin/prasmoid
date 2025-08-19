package build

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	initCmd "github.com/PRASSamin/prasmoid/cmd/init"
	"github.com/PRASSamin/prasmoid/cmd"

	"github.com/PRASSamin/prasmoid/utils"
)

func TestBuildCommand(t *testing.T) {
	// Store the original working directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get original working directory: %v", err)
	}
	// Defer restoring the original working directory
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore original working directory: %v", err)
		}
	}()

	testCases := []struct {
		name        string
		setup       func(t *testing.T) (string, func())
		expectError string
		verify      func(t *testing.T, tmpDir string)
	}{
		{
			name: "successful build",
			setup: func(t *testing.T) (string, func()) {
				return initCmd.SetupTestProject(t)
			},
			
			expectError: "",
			verify: func(t *testing.T, tmpDir string) {
				buildOutputDir := filepath.Join(tmpDir, "build")
				plasmoidID, _ := utils.GetDataFromMetadata("Id")
				version, _ := utils.GetDataFromMetadata("Version")
				zipFileName := plasmoidID.(string) + "-" + version.(string) + ".plasmoid"
				zipFilePath := filepath.Join(buildOutputDir, zipFileName)

				if _, err := os.Stat(zipFilePath); os.IsNotExist(err) {
					t.Fatalf("Expected .plasmoid file '%s' to be created, but it was not", zipFilePath)
				}

				unzipDir, err := os.MkdirTemp("", "unzip-test-")
				if err != nil {
					t.Fatalf("Failed to create temp dir for unzipping: %v", err)
				}
				defer func() {
					if err := os.RemoveAll(unzipDir); err != nil {
						t.Errorf("Failed to remove unzip dir: %v", err)
					}
				}()

				if err := unzip(t, zipFilePath, unzipDir); err != nil {
					t.Fatalf("Failed to unzip .plasmoid file: %v", err)
				}

				expectedFiles := []string{
					"metadata.json",
					"contents/ui/main.qml",
					"contents/config/main.xml",
					"contents/icons/prasmoid.svg",
				}

				for _, file := range expectedFiles {
					path := filepath.Join(unzipDir, file)
					if _, err := os.Stat(path); os.IsNotExist(err) {
						t.Errorf("Expected file '%s' to be in the zip, but it was not", file)
					}
				}
			},
		},
		{
			name: "build without valid plasmoid",
			setup: func(t *testing.T) (string, func()) {
				tmpDir, cleanup := initCmd.SetupTestProject(t)
				_ = os.Remove(filepath.Join(tmpDir, "metadata.json"))
				_ = os.RemoveAll(filepath.Join(tmpDir, "contents"))
				return tmpDir, cleanup
			},
			
			expectError: "current directory is not a valid plasmoid",
			verify: func(t *testing.T, tmpDir string) {
				// No build output expected
				buildOutputDir := filepath.Join(tmpDir, "build")
				if _, err := os.Stat(buildOutputDir); !os.IsNotExist(err) {
					t.Errorf("Expected build output directory %s not to exist, but it does", buildOutputDir)
				}
			},
		},
		{
			name: "build without id and version",
			setup: func(t *testing.T) (string, func()) {
				tmpDir, cleanup := initCmd.SetupTestProject(t)
				metadataPath := filepath.Join(tmpDir, "metadata.json")
				_ = os.Remove(metadataPath)
				metadata := map[string]interface{}{
					"KPlugin": map[string]interface{}{
						"Name":    "Test Plasmoid",
					},
				}
				data, _ := json.MarshalIndent(metadata, "", "  ")
				if err := os.WriteFile(metadataPath, data, 0644); err != nil {
					t.Fatalf("Failed to write metadata.json: %v", err)
				}
				return tmpDir, cleanup
			},
			
			expectError: "invalid metadata",
			verify: func(t *testing.T, tmpDir string) {
				// No build output expected
				buildOutputDir := filepath.Join(tmpDir, "build")
				if _, err := os.Stat(buildOutputDir); !os.IsNotExist(err) {
					t.Errorf("Expected build output directory %s not to exist, but it does", buildOutputDir)
				}
			},
		},
		{
			name: "build with failing translation compilation",
			setup: func(t *testing.T) (string, func()) {
				tmpDir, cleanup := initCmd.SetupTestProject(t)
				localesDir := filepath.Join(tmpDir, "locales")
				if err := os.Mkdir(localesDir, 0755); err != nil {
					t.Fatalf("Failed to create locales dir: %v", err)
				}
				if err := os.WriteFile(filepath.Join(localesDir, "es.po"), []byte("invalid po file"), 0644); err != nil {
					t.Fatalf("Failed to create invalid .po file: %v", err)
				}
				rcPath := filepath.Join(tmpDir, ".prasmoidrc")
				rcContent := `
[l10n]
  dir = "locales"
  pot_file = "messages.pot"
`
				if err := os.WriteFile(rcPath, []byte(rcContent), 0644); err != nil {
					t.Fatalf("Failed to write .prasmoidrc: %v", err)
				}

				return tmpDir, cleanup
			},
			
			expectError: "", // should not error out
			verify: func(t *testing.T, tmpDir string) {
				// The build should still complete successfully
				buildOutputDir := filepath.Join(tmpDir, "build")
				plasmoidID, _ := utils.GetDataFromMetadata("Id")
				version, _ := utils.GetDataFromMetadata("Version")
				zipFileName := plasmoidID.(string) + "-" + version.(string) + ".plasmoid"
				zipFilePath := filepath.Join(buildOutputDir, zipFileName)

				if _, err := os.Stat(zipFilePath); os.IsNotExist(err) {
					t.Fatalf("Expected .plasmoid file '%s' to be created, but it was not", zipFilePath)
				}
			},
		},
		{
			name: "failed to clear existing build dir",
			setup: func(t *testing.T) (string, func()) {
				tmpDir, cleanup := initCmd.SetupTestProject(t)
				buildDir := filepath.Join(tmpDir, "build")
				if err := os.Mkdir(buildDir, 0755); err != nil {
					t.Fatalf("Failed to create build dir: %v", err)
				}
				// Add a file inside the directory to make it non-empty
				filePath := filepath.Join(buildDir, "dummy")
				if err := os.WriteFile(filePath, []byte("test"), 0444); err != nil {
					t.Fatalf("Failed to create dummy file: %v", err)
				}
				// Make the directory read-only
				if err := os.Chmod(buildDir, 0555); err != nil {
					t.Fatalf("Failed to make build dir read-only: %v", err)
				}
				return tmpDir, cleanup
			},
			
			expectError: "failed to clean build dir",
			verify: func(t *testing.T, tmpDir string) {
				// Make the directory writable again
				_ = os.Chmod(filepath.Join(tmpDir, "build"), 0755)
				// Verify that the build dir exists
				buildDir := filepath.Join(tmpDir, "build")
				if _, err := os.Stat(buildDir); os.IsNotExist(err) {
					t.Errorf("Expected build dir to exist, but it does not")
				}
			},
		},
		{
			name: "failed to create build dir",
			setup: func(t *testing.T) (string, func()) {
				tmpDir, cleanup := initCmd.SetupTestProject(t)
				_ = os.Chmod(tmpDir, 0555)
				return tmpDir, cleanup
			},
			
			expectError: "failed to create build dir",
			verify: func(t *testing.T, tmpDir string) {
				_ = os.Chmod(tmpDir, 0755)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir, cleanup := tc.setup(t)

			defer cleanup()

			cmd.ConfigRC = utils.LoadConfigRC()

			var err error
			
			if tc.name=="successful build" {
				BuildCmd.Run(nil, []string{})
			}else{
				err= BuildPlasmoid()
			}

			if tc.expectError != "" && err != nil {
				if !strings.Contains(err.Error(), tc.expectError) {
					t.Errorf("expected error message: %s, got: %s", tc.expectError, err.Error())
				}
			} else if err != nil {
				t.Errorf("Unexpected error in output: %s", err.Error())
			} else if tc.expectError != "" {
				t.Errorf("Expected error message: %s, but got no error", tc.expectError)
			}
		
			if tc.verify != nil {
				tc.verify(t, tmpDir)
			}
		})
	}
}

// unzip extracts a zip archive to a destination directory
func unzip(t *testing.T, src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			t.Errorf("Error closing zip reader: %v", err)
		}
	}()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fpath, os.ModePerm); err != nil {
				t.Errorf("Error creating directory %s: %v", fpath, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			return err
		}

		_, err = io.Copy(outFile, rc)

		if err := outFile.Close(); err != nil {
			t.Errorf("Error closing output file: %v", err)
		}
		if err := rc.Close(); err != nil {
			t.Errorf("Error closing read closer: %v", err)
		}

		if err != nil {
			return err
		}
	}
	return nil
}

// Helper function to create a zip writer with a buffer
func newZipWriter() (*zip.Writer, *bytes.Buffer) {
	var buf bytes.Buffer
	zipWriter := zip.NewWriter(&buf)
	return zipWriter, &buf
}

func TestAddFileToZip(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tmpDir := t.TempDir()
		zipWriter, _ := newZipWriter()

		// Create a dummy file to add
		filePath := filepath.Join(tmpDir, "test.txt")
		if err := os.WriteFile(filePath, []byte("hello"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		err := AddFileToZip(zipWriter, filePath)
		if err != nil {
			t.Errorf("AddFileToZip() returned an unexpected error: %v", err)
		}
	})

	t.Run("file does not exist", func(t *testing.T) {
		zipWriter, _ := newZipWriter()
		err := AddFileToZip(zipWriter, "non-existent-file.txt")
		if err == nil {
			t.Error("AddFileToZip() expected an error for non-existent file, but got nil")
		}
	})
}

func TestAddDirToZip(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		tmpDir := t.TempDir()
		zipWriter, _ := newZipWriter()

		// Create a dummy directory structure
		contentsDir := filepath.Join(tmpDir, "contents")
		if err := os.Mkdir(contentsDir, 0755); err != nil {
			t.Fatalf("Failed to create contents dir: %v", err)
		}
		filePath := filepath.Join(contentsDir, "test.txt")
		if err := os.WriteFile(filePath, []byte("hello"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		err := AddDirToZip(zipWriter, contentsDir)
		if err != nil {
			t.Errorf("AddDirToZip() returned an unexpected error: %v", err)
		}
	})

	t.Run("directory does not exist", func(t *testing.T) {
		zipWriter, _ := newZipWriter()
		err := AddDirToZip(zipWriter, "non-existent-dir")
		if err == nil {
			t.Error("AddDirToZip() expected an error for non-existent directory, but got nil")
		}
	})

    t.Run("unreadable file in directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		zipWriter, _ := newZipWriter()

		// Create a file and make it unreadable
		filePath := filepath.Join(tmpDir, "unreadable.txt")
		if err := os.WriteFile(filePath, []byte("cant read this"), 0000); err != nil {
			t.Fatalf("Failed to create unreadable file: %v", err)
		}
        // Make writable again for cleanup
        defer func() { _ = os.Chmod(filePath, 0644) }()

		err := AddDirToZip(zipWriter, tmpDir)
		if err == nil {
			t.Error("AddDirToZip() expected an error for unreadable file, but got nil")
		}
	})
    t.Run("run with relative path", func(t *testing.T) {
		tmpDir := t.TempDir()
		zipWriter, _ := newZipWriter()
		originalDir, _ := os.Getwd()
		_ = os.Chdir(tmpDir)

		newFolder := filepath.Join(tmpDir, "newFolder")
		if err := os.Mkdir(newFolder, 0755); err != nil {
			t.Fatalf("Failed to create new folder: %v", err)
		}

		filePath := filepath.Join(newFolder, "unreadable.txt")
		if err := os.WriteFile(filePath, []byte("cant read this"), 0644); err != nil {
			t.Fatalf("Failed to create unreadable file: %v", err)
		}

		err := AddDirToZip(zipWriter, "./newFolder"); if err != nil {
			t.Errorf("AddDirToZip() returned an unexpected error: %v", err)
		}

		_ = os.Chdir(originalDir)
	})
}
