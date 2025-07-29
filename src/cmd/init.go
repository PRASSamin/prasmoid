/*
Copyright 2025 PRAS
*/
package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/PRASSamin/prasmoid/consts"
	"github.com/PRASSamin/prasmoid/types"
	"github.com/PRASSamin/prasmoid/utils"
)

type Author struct {
	Name  string `json:"Name,omitempty"`
	Email string `json:"Email,omitempty"`
}

type KPlugin struct {
	Authors          []Author `json:"Authors,omitempty"`
	Description      string   `json:"Description"`
	EnabledByDefault bool     `json:"EnabledByDefault"`
	FormFactors      []string `json:"FormFactors"`
	Id               string   `json:"Id"`
	License          string   `json:"License"`
	Name             string   `json:"Name"`
	Version          string   `json:"Version"`
}

type Metadata struct {
	KPackageStructure      string   `json:"KPackageStructure"`
	KPlugin                KPlugin  `json:"KPlugin"`
	XPlasmaAPIMinimumVersion string   `json:"X-Plasma-API-Minimum-Version"`
	XPlasmaProvides        []string `json:"X-Plasma-Provides"`
}

type ProjectConfig struct {
	Name        string
	Path        string
	ID          string
	Description string
	AuthorName  string
	AuthorEmail string
	License     string
	InitGit     bool
	Locales     []string
}

var Config ProjectConfig

var FileTemplates = map[string]string{
	"contents/ui/main.qml": consts.MAIN_QML,
	"contents/config/main.xml": consts.MAIN_XML,
	"contents/icons/prasmoid.svg": consts.PRASMOID_SVG,
	".gitignore": consts.GITIGNORE,
	"prasmoid.d.ts": consts.PRASMOID_DTS,
}

func init() {
	InitCmd.Flags().StringVarP(&Config.Name, "name", "n", "", "Project name")
	rootCmd.AddCommand(InitCmd)
}

var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new plasmoid project",
	Run: func(cmd *cobra.Command, args []string) {
		clearLine()
		printHeader()

		if err := gatherProjectConfig(); err != nil {
			color.Red("Failed to gather project config: %v", err)
			return
		}

		color.Yellow("Creating project at: %s", Config.Path)
		fmt.Println()

		if err := InitPlasmoid(); err != nil {
			color.Red("Failed to initialize plasmoid: "+
				"%v", err)
			return
		}

		if Config.InitGit {
			if err := initializeGitRepo(); err != nil {
				color.Yellow("Could not initialize git repository: %v", err)
			} else {
				color.Green("Initialized git repository.")
			}
		}

		fmt.Println()
		color.Green("Plasmoid initialized successfully!")
		fmt.Println()
		printNextSteps()
	},
}

func gatherProjectConfig() error {
	var qs = []*survey.Question{
		{
			Name: "Description",
			Prompt: &survey.Input{
				Message: "Description:",
				Default: "A new plasmoid project.",
			},
		},
		{
			Name: "License",
			Prompt: &survey.Select{
				Message: "Choose a license:",
				Options: []string{"GPL-2.0+", "GPL-3.0+", "MIT"},
				Default: "GPL-2.0+",
			},
		},
		{
			Name: "AuthorName",
			Prompt: &survey.Input{
				Message: "Author:",
			},
		},
	}

	namePrompt := &survey.Input{
			Message: "Project name:",
			Default: "MyPlasmoid",
	}
	
	invalidChars := regexp.MustCompile(`[\\/:*?"<>|\s@]`)
	if strings.TrimSpace(Config.Name) == "" || invalidChars.MatchString(Config.Name) {
		if err := survey.AskOne(namePrompt, &Config.Name, survey.WithValidator(validateProjectName)); err != nil {
			return err
		}
	}

	// Ask initial questions
	if err := survey.Ask(qs, &Config); err != nil {
		return err
	}

	if Config.AuthorName != "" {
		AuthorEmailQuestion := &survey.Input{
			Message: "Author email:",
		}
	 	if err := survey.AskOne(AuthorEmailQuestion, &Config.AuthorEmail); err != nil {
			return err
		}	
	}

	// Set project path and ID based on name
	if Config.Name == "." {
		Config.Path, _ = os.Getwd()
	} else {
		Config.Path, _ = filepath.Abs(fmt.Sprintf("./%s", Config.Name))
	}
	Config.ID = fmt.Sprintf("org.kde.%s", strings.ToLower(strings.Split(Config.Path, string(os.PathSeparator))[len(strings.Split(Config.Path, string(os.PathSeparator))) - 1]))

	// Ask for ID confirmation
	idQuestion := &survey.Input{
		Message: "Plasmoid ID:",
		Default: Config.ID,
	}
	if err := survey.AskOne(idQuestion, &Config.ID); err != nil {
		return err
	}

	Config.Locales = utils.AskForLocales()

	// Check for git and ask to initialize
	if utils.IsPackageInstalled("git") {
		gitQuestion := &survey.Confirm{
			Message: "Initialize a git repository?",
			Default: true,
		}
		if err := survey.AskOne(gitQuestion, &Config.InitGit); err != nil {
			return err
		}
	}

	return nil
}

func validateProjectName(ans interface{}) error {
	name := ans.(string)
	if name == "." {
		files, err := os.ReadDir(".")
		if err != nil {
			return fmt.Errorf("failed to read current directory: %v", err)
		}
		if len(files) > 0 {
			return errors.New("current directory is not empty. Please choose a specific project name")
		}
		return nil
	}

	invalidChars := regexp.MustCompile(`[\\/:*?"<>|\s@]`)
	if invalidChars.MatchString(name) {
		return errors.New("invalid characters in project name")
	}

	if _, err := os.Stat(name); !os.IsNotExist(err) {
		return errors.New("project directory already exists")
	}
	return nil
}

func InitPlasmoid() error {
	pm, err := utils.DetectPackageManager()
	if err != nil {
		return err
	}

	if !utils.IsPackageInstalled(consts.QmlFormatPackageName["binary"]) {
		color.Yellow("Installing qmlformat...")
		if err := utils.InstallQmlformat(pm); err != nil {
			return err
		}
	}

	if !utils.IsPackageInstalled(consts.PlasmoidPreviewPackageName["binary"]) {
		color.Yellow("Installing plasmoidviewer...")
		if err := utils.InstallPlasmoidPreview(pm); err != nil {
			return err
		}
	}

	if !utils.IsPackageInstalled(consts.CurlPackageName["binary"]) {
		color.Yellow("Installing curl...")
		if err := utils.InstallPackage(pm, consts.CurlPackageName["binary"], consts.CurlPackageName); err != nil {
			return err
		}
	}

	// Create project files from templates
	for relPath, content := range FileTemplates {
		if err := createFileFromTemplate(relPath, content); err != nil {
			return err
		}
	}

	// Create metadata.json
	if err := createMetadataFile(); err != nil {
		return err
	}

	// Create config.js
	if err := CreateConfigFile(Config.Locales); err != nil {
		return err
	}

	// Create custom commands directory
	_ = os.MkdirAll(filepath.Join(Config.Path, ".prasmoid/commands"), 0755)

	dest := filepath.Join(os.Getenv("HOME"), ".local/share/plasma/plasmoids", Config.ID) 
	
	// Remove if exists
	_ = os.Remove(dest)
	_ = os.RemoveAll(dest)
	
	// retrive current dir
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	
	// Link
	os.Symlink(filepath.Join(cwd, Config.Name), dest)
	return nil
}

func createFileFromTemplate(relPath, contentTmpl string) error {
	fullPath := filepath.Join(Config.Path, relPath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", filepath.Dir(fullPath), err)
	}

	tmpl, err := template.New(relPath).Parse(contentTmpl)
	if err != nil {
		return fmt.Errorf("failed to parse template for %s: %w", relPath, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, Config); err != nil {
		return fmt.Errorf("failed to execute template for %s: %w", relPath, err)
	}

	return os.WriteFile(fullPath, buf.Bytes(), 0644)
}

func CreateConfigFile(locales []string) error {
	fullPath := filepath.Join(Config.Path, "prasmoid.config.js")
	RC := types.Config{
		Commands: types.ConfigCommands{
			Dir: ".prasmoid/commands",
			Ignore: []string{},
		},
		I18n: types.ConfigI18n{
			Dir: "translations",
			Locales: locales,
		},
	}
	configData, _ := json.MarshalIndent(RC, "", "  ")
	content := fmt.Sprintf(`/// <reference path="prasmoid.d.ts" />
/** @type {PrasmoidConfig} */
const config = %v;`, string(configData))
	return os.WriteFile(fullPath, []byte(content), 0644)
}

func createMetadataFile() error {
	fullPath := filepath.Join(Config.Path, "metadata.json")
	var authors []map[string]interface{}
	if strings.TrimSpace(Config.AuthorName) != "" || strings.TrimSpace(Config.AuthorEmail) != "" {
		authorMap := map[string]interface{}{
			"Name":  Config.AuthorName,
			"Email": Config.AuthorEmail,
		}
		// Add localized author names
		for _, locale := range Config.Locales {
			cleanLocale := strings.Trim(locale, "\"")
			authorMap[fmt.Sprintf("Name[%s]", cleanLocale)] = Config.AuthorName
		}
		authors = append(authors, authorMap)
	}

	// Use a map for dynamic keys
	metadata := map[string]interface{}{
		"KPackageStructure": "Plasma/Applet",
		"KPlugin": map[string]interface{}{
			"Authors":          authors,
			"Description":      Config.Description,
			"EnabledByDefault": true,
			"FormFactors":      []string{"desktop", "tablet", "handset"},
			"Id":               Config.ID,
			"License":          Config.License,
			"Name":             Config.Name,
			"Version":          "0.0.1",
		},
		"X-Plasma-API-Minimum-Version": "6.0",
		"X-Plasma-Provides":          []string{},
	}

	// Add localized placeholders for Name and Description directly under KPlugin
	kplugin := metadata["KPlugin"].(map[string]interface{})
	for _, locale := range Config.Locales {
		cleanLocale := strings.Trim(locale, "\"")
		kplugin[fmt.Sprintf("Name[%s]", cleanLocale)] = Config.Name
		kplugin[fmt.Sprintf("Description[%s]", cleanLocale)] = Config.Description
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal metadata JSON: %w", err)
	}

	return os.WriteFile(fullPath, data, 0644)
}

func initializeGitRepo() error {
	cmd := exec.Command("git", "init")
	cmd.Dir = Config.Path
	return cmd.Run()
}

func printHeader() {
	star := color.New(color.FgHiMagenta, color.Bold).SprintFunc()
	fmt.Println(star("╭────────────────────────────────────────────╮"))
	fmt.Println(star("│    💠 Plasmoid Applet Project Generator    │"))
	fmt.Println(star("╰────────────────────────────────────────────╯"))
	fmt.Println()
}

func printNextSteps() {
	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Println("Next steps:")
	if Config.Name != "." {
		fmt.Printf("1. %s\n", cyan("cd ", Config.Name))
	}
	fmt.Printf("2. %s\n", cyan("plasmoid preview"))
}

func clearLine() {
	fmt.Print("\033[H\033[2J")
}
