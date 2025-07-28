/*
Copyright 2025 PRAS
*/
package cmd

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
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

func CheckForUpdates() {
	const host = "api.github.com"
	const path = "/repos/PRASSamin/prasmoid/releases/latest"
	const checkInterval = 24 * time.Hour

	cache, err := ReadUpdateCache()
	if err == nil {
        if lastCheckedStr, ok := cache["last_checked"].(string); ok {
            lastCheckedTime, err := time.Parse(time.RFC3339, lastCheckedStr)
            if err == nil && time.Since(lastCheckedTime) < checkInterval {
                if latestTag, ok := cache["latest_tag"].(string); ok {
                    if isUpdateAvailable(latestTag) {
                        printUpdateMessage(latestTag)
                    }
                }
                return
            }
        }
	}

	// Establish TLS connection to GitHub
	conn, err := tls.Dial("tcp", host+":443", nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// Manually write the HTTP GET request
	request := fmt.Sprintf(
		"GET %s HTTP/1.1\r\nHost: %s\r\nUser-Agent: Prasmoid-Updater\r\nConnection: close\r\n\r\n",
		path, host,
	)
	_, err = conn.Write([]byte(request))
	if err != nil {
		return
	}

	// Read the raw HTTP response
	raw, err := io.ReadAll(conn)
	if err != nil {
		return
	}

	// Parse the body from the response (after \r\n\r\n)
	parts := strings.SplitN(string(raw), "\r\n\r\n", 2)
	if len(parts) < 2 {
		return
	}
	headers := parts[0]
	body := parts[1]

	// Optional: parse status code from headers
	if !strings.Contains(headers, "200 OK") {
		return
	}

	// Extract latest tag and do the same flow
	latestTag := getLatestTag([]byte(body))
	writeUpdateCache(latestTag, []byte(body))
	if isUpdateAvailable(latestTag) {
		printUpdateMessage(latestTag)
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

	return strings.TrimPrefix(tag, "v")
}

func isUpdateAvailable(latestTag string) bool {
	if latestTag == "" {
		return false
	}
	re := regexp.MustCompile(`^([^-|_]+)`)
	matches := re.FindStringSubmatch(internal.AppMeta.Version)
	if len(matches) > 1 {
		return latestTag != matches[1]
	}
	
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

func ReadUpdateCache() (map[string]interface{}, error) {
	path := getCacheFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cache map[string]interface{}
	err = json.Unmarshal(data, &cache)
	return cache, err
}

func writeUpdateCache(tag string, body []byte) {
	var releaseData map[string]interface{}
	_ = json.Unmarshal(body, &releaseData)

	cache := map[string]interface{}{
		"last_checked": time.Now().Format(time.RFC3339),
		"latest_tag":   tag,
		"data":         releaseData,
	}
	data, _ := json.Marshal(cache)
	_ = os.WriteFile(getCacheFilePath(), data, 0o644)
}