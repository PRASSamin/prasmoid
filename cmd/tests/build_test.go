package cmd_tests

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/PRASSamin/prasmoid/cmd"

	"github.com/PRASSamin/prasmoid/utils"
)

func TestBuildCommand(t *testing.T) {
	// Set up a temporary project
	tmpDir, cleanup := setupTestProject(t)
	defer cleanup()

	// Execute the build command
	buildOutputDir := filepath.Join(tmpDir, "build") // Set the output dir for the test
	cmd.BuildCmd.Run(nil, []string{})

	// Verify the .plasmoid file was created
	plasmoidID, _ := utils.GetDataFromMetadata("Id")
	version, _ := utils.GetDataFromMetadata("Version")
	zipFileName := plasmoidID.(string) + "-" + version.(string) + ".plasmoid"
	zipFilePath := filepath.Join(buildOutputDir, zipFileName)

	if _, err := os.Stat(zipFilePath); os.IsNotExist(err) {
		t.Fatalf("Expected .plasmoid file '%s' to be created, but it was not", zipFilePath)
	}

	// Unzip and verify the contents
	unzipDir, err := os.MkdirTemp("", "unzip-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir for unzipping: %v", err)
	}
	defer os.RemoveAll(unzipDir)

	if err := unzip(zipFilePath, unzipDir); err != nil {
		t.Fatalf("Failed to unzip .plasmoid file: %v", err)
	}
	
	// Check for expected files in the unzipped directory
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
}

// unzip extracts a zip archive to a destination directory
func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
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

		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}
	return nil
}
