/*
Copyright 2025 PRAS
*/
package cmd

import (
	"bufio"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/PRASSamin/prasmoid/cmd/extendcli"
	"github.com/PRASSamin/prasmoid/consts"
	"github.com/PRASSamin/prasmoid/internal"
	"github.com/PRASSamin/prasmoid/types"
	"github.com/PRASSamin/prasmoid/utils"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// mockables
var (
	// utils
	utilsLoadConfigRC = utils.LoadConfigRC

	// extendcli
	extendcliDiscoverAndRegisterCustomCommands = extendcli.DiscoverAndRegisterCustomCommands

	// cobra
	rootCmdExecute = RootCmd.Execute

	// os
	osExit         = os.Exit
	osUserCacheDir = os.UserCacheDir
	osTempDir      = os.TempDir
	osReadFile     = os.ReadFile
	osWriteFile    = os.WriteFile
	osExecutable   = os.Executable
	osMkdirAll     = os.MkdirAll
	httpGet        = http.Get

	// time
	timeParse = time.Parse
	timeSince = time.Since
	timeNow   = time.Now

	// encoding/json
	jsonUnmarshal = json.Unmarshal
	jsonMarshal   = json.Marshal

	// golang.org/x/term
	termGetSize = term.GetSize

	// internal
	internalAppMetaDataVersion = internal.AppMetaData.Version

	// for testing purposes
	logPrintf = log.Printf
)

// project wise prasmoid config
var ConfigRC types.Config

func init() {
	ConfigRC = utilsLoadConfigRC()
	RootCmd.Flags().BoolP("version", "v", false, "show Prasmoid version")
	RootCmd.AddGroup(&cobra.Group{
		ID:    "custom",
		Title: "Custom Commands",
	})
	
	RootCmd.AddGroup(&cobra.Group{
		ID:    "cli",
		Title: "Maintenance Commands",
	})

	RootCmd.SetHelpCommandGroupID("cli")
}

// rootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "prasmoid",
	Short: "Manage plasmoid projects",
	Long:  "CLI for building, packaging, and managing KDE plasmoid projects efficiently.",
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.Flag("version").Changed {
			fmt.Println(internalAppMetaDataVersion)
			osExit(0)
		}

		if err := cmd.Help(); err != nil {
			logPrintf("Error displaying help: %v", err)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	extendcliDiscoverAndRegisterCustomCommands(RootCmd, ConfigRC)
	CheckForUpdates()
	
	RootCmd.SetUsageTemplate(consts.UsageTemplate)

	err := rootCmdExecute()
	if err != nil {
		osExit(1)
	}
}

// -------- UPDATE CHECKER --------

var CheckForUpdates = func() {
	const checkInterval = 24 * time.Hour

	cache, err := readUpdateCache()
	if err == nil {
		if lastCheckedStr, ok := cache["last_checked"].(string); ok {
			lastCheckedTime, err := timeParse(time.RFC3339, lastCheckedStr)
			if err == nil && timeSince(lastCheckedTime) < checkInterval {
				if latestHash, ok := cache["latest_hash"].(string); ok {
					isAvailable, currentHash := isUpdateAvailable(latestHash)
					if isAvailable {
						if latestTag, ok := cache["latest_tag"].(string); ok {
							printUpdateMessage(latestTag, latestHash, currentHash)
						}
					}
				}
				return
			}
		}
	}

	releaseJSON, err := fetchURL("https://api.github.com/repos/PRASSamin/prasmoid/releases/latest")
	if err != nil {
		return
	}

	var releaseData map[string]interface{}
	if err := jsonUnmarshal([]byte(releaseJSON), &releaseData); err != nil {
		return
	}

	sha256sums, err := getSha256Sums(releaseData["assets"])

	if err != nil {
		return
	}

	assetName := "prasmoid"
	if strings.Contains(internalAppMetaDataVersion, "-portable") {
		assetName = "prasmoid-portable"
	}

	latestHash := parseChecksums(sha256sums, assetName)
	latestTag := getLatestTag(releaseData)

	writeUpdateCache(latestTag, latestHash)

	isAvailable, currentHash := isUpdateAvailable(latestHash)
	if isAvailable {
		printUpdateMessage(latestTag, latestHash, currentHash)
	}
}

var getLatestTag = func(releaseData map[string]interface{}) string {
	tag, ok := releaseData["tag_name"].(string)
	if !ok {
		return ""
	}
	return strings.TrimPrefix(tag, "v")
}

var parseChecksums = func(content, assetName string) string {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) == 2 && parts[1] == assetName {
			return parts[0]
		}
	}
	return ""
}

var calculateFileSHA256 = func(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

var fetchURL = func(rawURL string) (string, error) {
	resp, err := httpGet(rawURL)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non-200 status for %s: %s", rawURL, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

var getSha256Sums = func(assets interface{}) (string, error) {
	assetsList, ok := assets.([]interface{})

	if !ok {
		return "", fmt.Errorf("invalid assets format")
	}

	var shaURL string
	for _, asset := range assetsList {
		assetMap, ok := asset.(map[string]interface{})
		if !ok {
			continue
		}
		if name, ok := assetMap["name"].(string); ok && name == "SHA256SUMS" {
			if url, ok := assetMap["browser_download_url"].(string); ok {
				shaURL = url
				break
			}
		}
	}

	if shaURL == "" {
		return "", fmt.Errorf("SHA256SUMS URL not found in release assets")
	}

	sha256sums, err := fetchURL(shaURL)

	if err != nil {
		return "", err
	}
	return sha256sums, nil
}

var isUpdateAvailable = func(latestHash string) (bool, string) {
	if latestHash == "" {
		return false, ""
	}
	exePath, err := osExecutable()
	if err != nil {
		return false, ""
	}
	currentHash, err := calculateFileSHA256(exePath)
	if err != nil {
		return false, ""
	}
	return currentHash != latestHash, currentHash
}

var printUpdateMessage = func(latest string, latestHash string, currentHash string) {
	// Get terminal width
	width, _, err := termGetSize(int(os.Stdout.Fd()))
	if err != nil {
		width = 70
	}

	star := color.New(color.FgHiYellow, color.Bold).SprintFunc()

	bottom := strings.Repeat("─", width)

	printLine := func(content string) string {
		return fmt.Sprintf(" %s ", content)
	}

	currentVersion := fmt.Sprintf("%s.%s", strings.Split(internalAppMetaDataVersion, "-")[0], currentHash[:4])
	latestVersion := fmt.Sprintf("%s.%s", latest, latestHash[:4])

	fmt.Println(star(bottom))
	fmt.Println(star(printLine(fmt.Sprintf("💠 Prasmoid update available! %s → %s", currentVersion, latestVersion))))
	fmt.Println(star(printLine("Run `prasmoid upgrade` to update")))
	fmt.Println(star(bottom))
	fmt.Println()
}

var GetCacheFilePath = func() string {
	dir, err := osUserCacheDir()
	if err != nil || dir == "" {
		dir = osTempDir()
	}

	// Ensure the directory exists; if creation fails, fallback to temp dir
	if mkErr := osMkdirAll(dir, 0o755); mkErr != nil {
		dir = osTempDir()
		_ = osMkdirAll(dir, 0o755)
	}
	return filepath.Join(dir, ".prasmoid")
}

var readUpdateCache = func() (map[string]interface{}, error) {
	path := GetCacheFilePath()

	data, err := osReadFile(path)
	if err != nil {
		return nil, err
	}

	var cache map[string]interface{}
	err = jsonUnmarshal(data, &cache)
	return cache, err
}

var writeUpdateCache = func(tag, hash string) {
	cache := map[string]interface{}{
		"last_checked": timeNow().Format(time.RFC3339),
		"latest_tag":   tag,
		"latest_hash":  hash,
	}
	data, _ := jsonMarshal(cache)
	_ = osWriteFile(GetCacheFilePath(), data, 0o644)
}