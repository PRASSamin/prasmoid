package init

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/AlecAivazis/survey/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SetupTestProject creates a temporary directory with a dummy metadata.json file.
// It returns the path to the temporary directory and a cleanup function.
func SetupTestProject(t *testing.T) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "plasmoid-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Create a dummy metadata.json
	metadata := map[string]interface{}{
		"KPlugin": map[string]interface{}{
			"Id":      "org.kde.testplasmoid",
			"Version": "1.0.0",
			"Name":    "Test Plasmoid",
		},
	}
	metadataPath := filepath.Join(tmpDir, "metadata.json")
	data, _ := json.MarshalIndent(metadata, "", "  ")
	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		t.Fatalf("Failed to write metadata.json: %v", err)
	}

	// Create a dummy contents directory
	if err := os.Mkdir(filepath.Join(tmpDir, "contents"), 0755); err != nil {
		t.Fatalf("Failed to create contents dir: %v", err)
	}

	for relPath, content := range FileTemplates {
		fullPath := filepath.Join(tmpDir, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Errorf("failed to create directory %s: %v", filepath.Dir(fullPath), err)
		}

		if _, err := os.Stat(fullPath); err == nil {
			t.Errorf("Skipping existing file: %s", relPath)
		}

		tmpl, err := template.New(relPath).Parse(content)
		if err != nil {
			t.Errorf("failed to parse template for %s: %v", relPath, err)
		}

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, Config); err != nil {
			t.Errorf("failed to execute template for %s: %v", relPath, err)
		}

		if err := os.WriteFile(fullPath, buf.Bytes(), 0644); err != nil {
			t.Errorf("failed to write file %s: %v", fullPath, err)
		}
	}

	// Change to the temp directory for the duration of the test
	originalWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory to %s: %v", tmpDir, err)
	}

	cleanup := func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("Failed to restore original directory: %v", err)
		}
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("Failed to remove temporary directory: %v", err)
		}
	}

	return tmpDir, cleanup
}

func SetupTempDir(t *testing.T) (string, func()) {
	tmpDir, err := os.MkdirTemp("", "plasmoid-test-")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	// Change to the temp directory for the duration of the test
	originalWd, _ := os.Getwd()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory to %s: %v", tmpDir, err)
	}

	cleanup := func() {
		if err := os.Chdir(originalWd); err != nil {
			t.Errorf("Failed to restore original directory: %v", err)
		}
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Errorf("Failed to remove temporary directory: %v", err)
		}
	}

	return tmpDir, cleanup
}

func TestInitCmd_Run(t *testing.T) {
	t.Run("fails when gatherProjectConfig errors", func(t *testing.T) {
		// Mock gatherProjectConfig to fail
		oldGather := gatherProjectConfig
		gatherProjectConfig = func() error {
			return fmt.Errorf("config error")
		}
		defer func() { gatherProjectConfig = oldGather }()

		// Capture output
		buf := &bytes.Buffer{}
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Run command
		InitCmd.Run(InitCmd, []string{})

		// Restore stdout
		_ = w.Close()
		os.Stdout = oldStdout
		_, _ = buf.ReadFrom(r)
		assert.Contains(t, buf.String(), "Failed to gather project config: config error")
	})

	t.Run("fails when InitPlasmoid errors", func(t *testing.T) {
		// Mock gatherProjectConfig to succeed
		oldGather := gatherProjectConfig
		gatherProjectConfig = func() error { return nil }
		defer func() { gatherProjectConfig = oldGather }()

		// Mock InitPlasmoid to fail
		oldInit := InitPlasmoid
		InitPlasmoid = func() error { return errors.New("init error") }
		defer func() { InitPlasmoid = oldInit }()

		buf := &bytes.Buffer{}
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		InitCmd.Run(InitCmd, []string{})

		_ = w.Close()
		os.Stdout = oldStdout
		_, _ = buf.ReadFrom(r)

		assert.Contains(t, buf.String(), "Failed to initialize plasmoid: init error")
	})

	t.Run("fails when InitGit errors", func(t *testing.T) {
		Config.InitGit = true

		// Mock gatherProjectConfig to succeed
		oldGather := gatherProjectConfig
		gatherProjectConfig = func() error { return nil }
		defer func() { gatherProjectConfig = oldGather }()

		// Mock InitPlasmoid to fail
		oldInit := InitPlasmoid
		InitPlasmoid = func() error { return nil }
		defer func() { InitPlasmoid = oldInit }()

		// Mock InitGit to fail
		oldGit := initializeGitRepo
		initializeGitRepo = func() error { return errors.New("init error") }
		defer func() { initializeGitRepo = oldGit }()

		buf := &bytes.Buffer{}
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		InitCmd.Run(InitCmd, []string{})

		_ = w.Close()
		os.Stdout = oldStdout
		_, _ = buf.ReadFrom(r)

		assert.Contains(t, buf.String(), "Could not initialize git repository")
	})
}

func TestGatherProjectConfig(t *testing.T) {
	t.Run("fails when project name validation fails", func(t *testing.T) {
		// Invalid name in config
		Config.Name = "Invalid@Name"
		Config.AuthorName = "PRAS"

		// Mock survey.AskOne to return error when asking project name
		oldAskOne := surveyAskOne
		surveyAskOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
			return errors.New("askone error")
		}
		defer func() { surveyAskOne = oldAskOne }()

		err := gatherProjectConfig()
		assert.Error(t, err)
		assert.Equal(t, "askone error", err.Error())
	})

	t.Run("fails when survey.Ask returns error", func(t *testing.T) {
		Config.Name = "MyPlasmoid"

		// Mock survey.Ask to fail
		oldAsk := surveyAsk
		surveyAsk = func(qs []*survey.Question, response interface{}, opts ...survey.AskOpt) error {
			return errors.New("ask error")
		}
		defer func() { surveyAsk = oldAsk }()

		err := gatherProjectConfig()
		assert.Error(t, err)
		assert.Equal(t, "ask error", err.Error())
	})

	t.Run("sets path and ID correctly", func(t *testing.T) {
		// Set a clean directory for this test
		tmpDir, cleanup := SetupTempDir(t)
		defer cleanup()

		Config.Name = "MyPlasmoid"

		// Mock survey to always succeed
		oldAsk := surveyAsk
		oldAskOne := surveyAskOne
		surveyAsk = func(qs []*survey.Question, response interface{}, opts ...survey.AskOpt) error {
			return nil
		}
		surveyAskOne = func(p survey.Prompt, response interface{}, opts ...survey.AskOpt) error {
			// This mock simulates survey's behavior of writing the default value
			// back to the response if the user provides no input.
			if s, ok := response.(*string); ok {
				if input, ok := p.(*survey.Input); ok {
					*s = input.Default
				}
			}
			if b, ok := response.(*bool); ok {
				*b = true // Assume user confirms "yes" to git init
			}
			return nil
		}
		defer func() {
			surveyAsk = oldAsk
			surveyAskOne = oldAskOne
		}()

		// Mock utils.AskForLocales
		oldLocales := utilsAskForLocales
		utilsAskForLocales = func(defaultLocales ...[]string) []string { return []string{"en"} }
		defer func() { utilsAskForLocales = oldLocales }()

		// Mock utils.IsPackageInstalled
		oldIsPkg := utilsIsPackageInstalled
		utilsIsPackageInstalled = func(s string) bool { return true } // Say git is installed
		defer func() { utilsIsPackageInstalled = oldIsPkg }()

		err := gatherProjectConfig()
		require.NoError(t, err)

		expectedPath := filepath.Join(tmpDir, "MyPlasmoid")
		assert.Equal(t, expectedPath, Config.Path)
		assert.Equal(t, "org.kde.myplasmoid", Config.ID)
		assert.True(t, Config.InitGit)
	})
}

func TestValidateProjectName(t *testing.T) {
	t.Run("allows valid name", func(t *testing.T) {
		_, cleanup := SetupTestProject(t)
		defer cleanup()
		err := validateProjectName("MyValidProject")
		assert.NoError(t, err)
	})

	t.Run("disallows existing directory", func(t *testing.T) {
		_, cleanup := SetupTestProject(t)
		defer cleanup()
		_ = os.Mkdir("ExistingDir", 0755)
		err := validateProjectName("ExistingDir")
		assert.Error(t, err)
		assert.Equal(t, "project directory already exists", err.Error())
	})

	t.Run("disallows invalid characters", func(t *testing.T) {
		err := validateProjectName("Invalid@Name")
		assert.Error(t, err)
		assert.Equal(t, "invalid characters in project name", err.Error())
	})

	t.Run("disallows dot in non-empty directory", func(t *testing.T) {
		_, cleanup := SetupTestProject(t)
		defer cleanup()
		_ = os.WriteFile("somefile.txt", []byte("hello"), 0644)
		err := validateProjectName(".")
		assert.Error(t, err)
		assert.Equal(t, "current directory is not empty. Please choose a specific project name", err.Error())
	})

	t.Run("allows dot in empty directory", func(t *testing.T) {
		_, cleanup := SetupTempDir(t)
		defer cleanup()
		err := validateProjectName(".")
		assert.NoError(t, err)
	})

	t.Run("fails if cannot read directory", func(t *testing.T) {
		// Mock osReadDir to return an error
		oldReadDir := osReadDir
		osReadDir = func(name string) ([]fs.DirEntry, error) {
			return nil, errors.New("read error")
		}
		defer func() { osReadDir = oldReadDir }()

		err := validateProjectName(".")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to read current directory")
	})
}

func TestCreateFileFromTemplate(t *testing.T) {
	t.Run("creates file successfully", func(t *testing.T) {
		_, cleanup := SetupTestProject(t)
		defer cleanup()

		Config = ProjectConfig{Path: ".", Name: "Test"}
		relPath := "test.txt"
		contentTmpl := "Hello, {{.Name}}"

		err := CreateFileFromTemplate(relPath, contentTmpl)
		require.NoError(t, err)

		content, err := os.ReadFile("test.txt")
		require.NoError(t, err)
		assert.Equal(t, "Hello, Test", string(content))
	})

	t.Run("fails on mkdir error", func(t *testing.T) {
		// Mock osMkdirAll to fail
		oldMkdirAll := osMkdirAll
		osMkdirAll = func(path string, perm os.FileMode) error {
			return errors.New("mkdir failed")
		}
		defer func() { osMkdirAll = oldMkdirAll }()

		err := CreateFileFromTemplate("dir/test.txt", "content")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create directory")
	})

	t.Run("fails on template parse error", func(t *testing.T) {
		_, cleanup := SetupTestProject(t)
		defer cleanup()
		err := CreateFileFromTemplate("test.txt", "Hello, {{.Invalid}")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to parse template")
	})

	t.Run("fails on template execute error", func(t *testing.T) {
		_, cleanup := SetupTestProject(t)
		defer cleanup()
		// Create a template that expects a field that doesn't exist in ProjectConfig
		err := CreateFileFromTemplate("test.txt", "Hello, {{.NonExistentField}}")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to execute template")
	})
}

func TestCreateMetadataFile(t *testing.T) {
	t.Run("creates metadata without authors", func(t *testing.T) {
		_, cleanup := SetupTempDir(t)
		defer cleanup()
		Config = ProjectConfig{Path: ".", Name: "NoAuthorPlasmoid", AuthorName: "", AuthorEmail: ""}

		err := createMetadataFile()
		require.NoError(t, err)

		content, err := os.ReadFile("metadata.json")
		require.NoError(t, err)
		assert.Contains(t, string(content), "\"Authors\": []")
	})

	t.Run("creates metadata with authors and locales", func(t *testing.T) {
		_, cleanup := SetupTempDir(t)
		defer cleanup()
		Config = ProjectConfig{
			Path:        ".",
			Name:        "MyPlasmoid",
			Description: "A test plasmoid.",
			AuthorName:  "Test Author",
			AuthorEmail: "test@example.com",
			Locales:     []string{"en_GB", "fr"},
		}

		err := createMetadataFile()
		require.NoError(t, err)

		content, err := os.ReadFile("metadata.json")
		require.NoError(t, err)

		// Check author
		assert.Contains(t, string(content), "\"Name\": \"Test Author\"")
		assert.Contains(t, string(content), "\"Email\": \"test@example.com\"")

		// Check localized author name
		assert.Contains(t, string(content), "\"Name[en_GB]\": \"Test Author\"")
		assert.Contains(t, string(content), "\"Name[fr]\": \"Test Author\"")

		// Check localized plugin name and description
		assert.Contains(t, string(content), "\"Name[en_GB]\": \"MyPlasmoid\"")
		assert.Contains(t, string(content), "\"Description[en_GB]\": \"A test plasmoid.\"")
		assert.Contains(t, string(content), "\"Name[fr]\": \"MyPlasmoid\"")
		assert.Contains(t, string(content), "\"Description[fr]\": \"A test plasmoid.\"")
	})

	t.Run("fails on json marshal error", func(t *testing.T) {
		// Mock json.MarshalIndent to fail
		oldMarshal := jsonMarshalIndent
		jsonMarshalIndent = func(v interface{}, prefix, indent string) ([]byte, error) {
			return nil, errors.New("json error")
		}
		defer func() { jsonMarshalIndent = oldMarshal }()

		err := createMetadataFile()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to marshal metadata JSON")
	})
}

func TestInitPlasmoid(t *testing.T) {
	t.Run("fails if install dependencies fails", func(t *testing.T) {
		_, cleanup := SetupTempDir(t)
		defer cleanup()

		oldInstall := utilsInstallDependencies
		utilsInstallDependencies = func() error { return errors.New("install error") }
		defer func() { utilsInstallDependencies = oldInstall }()

		err := InitPlasmoid()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "install error")
	})

	t.Run("fails if create file from template fails", func(t *testing.T) {
		_, cleanup := SetupTempDir(t)
		defer cleanup()

		// Mock dependencies
		oldInstall := utilsInstallDependencies
		utilsInstallDependencies = func() error { return nil }
		defer func() { utilsInstallDependencies = oldInstall }()
		oldCreate := CreateFileFromTemplate
		CreateFileFromTemplate = func(relPath, contentTmpl string) error {
			return errors.New("template error")
		}
		defer func() { CreateFileFromTemplate = oldCreate }()

		err := InitPlasmoid()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "template error")
	})

	t.Run("fails if create metadata file fails", func(t *testing.T) {
		_, cleanup := SetupTempDir(t)
		defer cleanup()
		// Mock dependencies
		oldInstall := utilsInstallDependencies
		utilsInstallDependencies = func() error { return nil }
		defer func() { utilsInstallDependencies = oldInstall }()
		oldCreateMeta := createMetadataFile
		createMetadataFile = func() error { return errors.New("metadata error") }
		defer func() { createMetadataFile = oldCreateMeta }()
		oldCreateTmpl := CreateFileFromTemplate
		CreateFileFromTemplate = func(relPath, contentTmpl string) error { return nil }
		defer func() { CreateFileFromTemplate = oldCreateTmpl }()

		err := InitPlasmoid()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "metadata error")
	})

	t.Run("fails if create config file fails", func(t *testing.T) {
		_, cleanup := SetupTempDir(t)
		defer cleanup()
		// Mock dependencies
		oldInstall := utilsInstallDependencies
		utilsInstallDependencies = func() error { return nil }
		defer func() { utilsInstallDependencies = oldInstall }()
		oldCreateMeta := createMetadataFile
		createMetadataFile = func() error { return nil }
		defer func() { createMetadataFile = oldCreateMeta }()
		oldCreateTmpl := CreateFileFromTemplate
		CreateFileFromTemplate = func(relPath, contentTmpl string) error { return nil }
		defer func() { CreateFileFromTemplate = oldCreateTmpl }()
		oldCreateConfig := CreateConfigFile
		CreateConfigFile = func(locales []string) error { return errors.New("config error") }
		defer func() { CreateConfigFile = oldCreateConfig }()

		err := InitPlasmoid()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "config error")
	})

	t.Run("fails if symlink fails", func(t *testing.T) {
		_, cleanup := SetupTempDir(t)
		defer cleanup()
		Config.Path = "."
		Config.Name = "MyPlasmoid" // Symlink needs a name

		// Mock os.Symlink to fail
		oldSymlink := osSymlink
		osSymlink = func(oldname, newname string) error { return errors.New("symlink error") }
		defer func() { osSymlink = oldSymlink }()

		// Mock other dependencies to succeed
		oldInstall := utilsInstallDependencies
		utilsInstallDependencies = func() error { return nil }
		defer func() { utilsInstallDependencies = oldInstall }()
		oldCreateMeta := createMetadataFile
		createMetadataFile = func() error { return nil }
		defer func() { createMetadataFile = oldCreateMeta }()
		oldCreateTmpl := CreateFileFromTemplate
		CreateFileFromTemplate = func(relPath, contentTmpl string) error { return nil }
		defer func() { CreateFileFromTemplate = oldCreateTmpl }()
		oldCreateConfig := CreateConfigFile
		CreateConfigFile = func(locales []string) error { return nil }
		defer func() { CreateConfigFile = oldCreateConfig }()

		err := InitPlasmoid()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to create symlink")
	})
}

func TestCreateConfigFile(t *testing.T) {
	// make a temp dir to sandbox file writes
	tmpDir := t.TempDir()
	origPath := Config.Path
	Config.Path = tmpDir
	defer func() { Config.Path = origPath }()

	locales := []string{"en", "bn", "fr"}

	// run the function
	err := CreateConfigFile(locales)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// verify file created
	fullPath := filepath.Join(tmpDir, "prasmoid.config.js")
	data, err := os.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("expected file to exist, got read error: %v", err)
	}

	content := string(data)

	require.Contains(t, content, `"translations"`)
	for _, loc := range locales {
		require.Contains(t, content, loc)
	}

	// ensure proper JS wrapper
	require.Contains(t, content, "/// <reference path=")
	require.Contains(t, content, "const config =")
}

func TestInitializeGitRepo(t *testing.T) {
	t.Run("fails when git command errors", func(t *testing.T) {
		// Mock exec.Command to return a failing command
		oldExec := execCommand
		execCommand = func(name string, arg ...string) *exec.Cmd {
			return exec.Command("non-existent-command")
		}
		defer func() { execCommand = oldExec }()

		err := initializeGitRepo()
		assert.Error(t, err)
	})
}
