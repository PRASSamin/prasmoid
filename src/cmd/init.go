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

	"github.com/PRASSamin/prasmoid/deps"
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
}

var Config ProjectConfig

var FileTemplates = map[string]string{
	"contents/ui/main.qml": `import QtQuick 6.5
import QtQuick.Layouts 6.5
import org.kde.kirigami 2.20 as Kirigami
import org.kde.plasma.core as PlasmaCore
import org.kde.plasma.plasmoid 2.0

PlasmoidItem {
    id: root

    Plasmoid.backgroundHints: PlasmaCore.Types.NoBackground

    RowLayout {
        id: rowLayout
        anchors.centerIn: root

        Text {
            text: "Hello from {{.Name}}!"
            color: "white"
            font.pointSize: 18
        }
    }
}`,
	"contents/config/main.xml": `<kcfg>
  <group name="General">
    <entry name="exampleOption" type="Bool">
      <default>true</default>
    </entry>
  </group>
</kcfg>
`,
	"contents/icons/plasmoid.svg": `
<svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 32 32">
  <circle cx="16" cy="16" r="14" fill="#4A90E2"/>
  <text x="16" y="21" text-anchor="middle" font-size="12" fill="#fff">P</text>
</svg>
`,
	".gitignore": `# Build artifacts
build/
*.plasmoid

# IDE files
.vscode/
.idea/

# OS-specific
.DS_Store
`,
"prasmoid.config.js": `const config = {
  commands: {
    dir: ".prasmoid/commands",
    ignore: []
  },
};
`,
"prasmoid.d.ts": `/**
 * This file provides type definitions for the custom command environment in Prasmoid.
 * It is used by code editors like VS Code to provide autocompletion and type-checking.
 *
 * @see https://www.typescriptlang.org/docs/handbook/jsdoc-supported-types.html
 */

/**
 * The context object passed to every custom command's Run function.
 */
interface CommandContext {
  /**
   * Returns the command-line arguments passed after the command name.
   * @returns {string[]} An array of arguments.
   * @example
   * const args = ctx.Args();
   * console.log(args[0]);
   */
  Args(): string[];

  /**
   * Provides access to the flags passed to the command.
   */
  Flags(): {
    /**
     * Retrieves the value of a specific flag.
     * @param name The name of the flag to retrieve.
     * @returns {string | boolean | undefined} The value of the flag, or undefined if not found.
     * @example
     * const name = ctx.Flags().get("name");
     */
    get(name: string): string | boolean | undefined;

    /**
     * You can also access flags directly as properties, but using get() is recommended for better type safety.
     * @example
     * const myFlag = ctx.Flags().myFlagName; // Type is 'any'
     */
    [key: string]: any;
  };
}

/**
 * The global prasmoid module for interacting with the project.
 */
declare module "prasmoid" {
  /**
   * Retrieves a value from the project's metadata.json file.
   * @param key The key from the "KPlugin" section of metadata.json (e.g., "Id", "Version").
   * @returns {string | undefined} The value from the metadata, or undefined if not found.
   */
  export function getMetadata(key: string): string | undefined;
  /**
   * Registers a custom command.
   * @param config The configuration for the command.
   */
  export function Command(config: Config): void;
}

/**
 * Configuration for the custom command.
 */
interface Config {
  run: (ctx: CommandContext) => void;
  /** A brief description of your command. */
  short: string;
  /** A longer description that spans multiple lines. */
  long: string;
  /** Optional aliases for the command. */
  alias?: string[];
  /** Flag definitions for the command. */
  flags?: {
    name: string;
    shorthand?: string;
    type: "string" | "boolean";
    default?: string | boolean;
    description: string;
  }[];
}

interface Console {
  /**
   * Logs a red-colored message.
   */
  red(...message: any[]): void;

  /**
   * Logs a green-colored message.
   */
  green(...message: any[]): void;

  /**
   * Logs a yellow-colored message.
   */
  yellow(...message: any[]): void;

  /**
   * Flexible color logger. Last argument must be a color string.
   * @example
   * console.color("Hey", "you!", "red")
   */
  color(
    ...message: any[],
    color: "red" | "green" | "yellow" | "blue" | "magenta" | "cyan" | "black"
  ): void;
}

declare var console: Console;
`,
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

		fmt.Println()
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
	var missingDeps []string
	for _, dep := range deps.Dependencies {
		if !utils.IsPackageInstalled(dep) {
			missingDeps = append(missingDeps, dep)
		}
	}

	if len(missingDeps) > 0 {
		color.Yellow("Installing missing dependencies: %s", strings.Join(missingDeps, ", "))
		if err := utils.InstallPackages(missingDeps); err != nil {
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

	if _, err := os.Stat(fullPath); err == nil {
		color.Yellow("Skipping existing file: %s", relPath)
		return nil
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

func createMetadataFile() error {
	fullPath := filepath.Join(Config.Path, "metadata.json")
	var authors []Author
	if strings.TrimSpace(Config.AuthorName) != "" || strings.TrimSpace(Config.AuthorEmail) != "" {
		authors = append(authors, Author{Name: Config.AuthorName, Email: Config.AuthorEmail})
	}

	metadata := Metadata{
		KPackageStructure: "Plasma/Applet",
		KPlugin: KPlugin{
			Authors:          authors,
			Description:      Config.Description,
			EnabledByDefault: true,
			FormFactors:      []string{"desktop", "tablet", "handset"},
			Id:               Config.ID,
			License:          Config.License,
			Name:             Config.Name,
			Version:          "0.0.1",
		},
		XPlasmaAPIMinimumVersion: "6.0",
		XPlasmaProvides:          []string{},
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
