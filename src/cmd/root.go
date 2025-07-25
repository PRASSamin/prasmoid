/*
Copyright 2025 PRAS
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/PRASSamin/prasmoid/internal"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func init() {
	rootCmd.Flags().BoolP("version", "v", false, "show Prasmoid version")
	rootCmd.AddGroup(&cobra.Group{
		ID:    "custom",
		Title: "Custom Commands:",
	})
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "prasmoid",
	Short: "Manage plasmoid projects",
	Long:  "CLI for building, packaging, and managing KDE plasmoid projects efficiently.",
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.Flag("version").Changed {
			fmt.Println(internal.AppMeta.Version)
			os.Exit(0)
		}

		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	DiscoverAndRegisterCustomCommands(rootCmd)

	CheckForUpdates()

	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}


// -------- UPDATE CHECKER --------

type LTSVersionData struct {
	Tag_name string `json:"tag_name"`
	Assets []struct {
		URL string `json:"url"`
		Name string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
type UpdateCache struct {
	LastChecked time.Time `json:"last_checked"`
	LatestTag   string    `json:"latest_tag"`
	Data LTSVersionData
}

func CheckForUpdates() {
	const githubAPI = "https://api.github.com/repos/PRASSamin/prasmoid/releases/latest"
	const checkInterval = 24 * time.Hour

	cache, err := ReadUpdateCache()
	if err == nil && time.Since(cache.LastChecked) < checkInterval {
		if isUpdateAvailable(cache.LatestTag) {
			printUpdateMessage(cache.LatestTag)
		}
		return
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(githubAPI)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if resp.StatusCode == http.StatusOK {
		latestTag := getLatestTag(body)
		writeUpdateCache(latestTag, body)
		if isUpdateAvailable(latestTag) {
			printUpdateMessage(latestTag)
		}
	}
}

func getLatestTag(body []byte) string {
	var tagData map[string]interface{}

	err := json.Unmarshal(body, &tagData)
	if err != nil {
		return ""
	}

	tag, ok := tagData["tag_name"].(string)
	if !ok {
		return ""
	}

	re := regexp.MustCompile(`^v`)
	return re.ReplaceAllString(tag, "")
}

func isUpdateAvailable(latestTag string) bool {
	return latestTag != internal.AppMeta.Version
}

func printUpdateMessage(latest string) {
	// Get terminal width
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 70 // fallback 
	}

	star := color.New(color.FgHiYellow, color.Bold).SprintFunc()

	// Borders
	bottom := strings.Repeat("─", width)

	printLine := func(content string) string {
		return fmt.Sprintf(" %s ", content)
	}

	fmt.Println(star(bottom))
	fmt.Println(star(printLine(fmt.Sprintf("💠 Prasmoid update available! %s → %s", internal.AppMeta.Version, latest))))
	fmt.Println(star(printLine("Run `prasmoid update me` to update")))
	fmt.Println(star(bottom))
	fmt.Println()
}

func getCacheFilePath() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		dir = os.TempDir()
	}
	return filepath.Join(dir, "prasmoid_update.json")
}

func ReadUpdateCache() (*UpdateCache, error) {
	path := getCacheFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cache UpdateCache
	err = json.Unmarshal(data, &cache)
	return &cache, err
}

func writeUpdateCache(tag string, body []byte) {
	var vdata LTSVersionData
	_ = json.Unmarshal(body, &vdata)

	cache := UpdateCache{
		LastChecked: time.Now(),
		LatestTag:   tag,
		Data: vdata,
	}
	data, _ := json.Marshal(cache)
	_ = os.WriteFile(getCacheFilePath(), data, 0o644)
}