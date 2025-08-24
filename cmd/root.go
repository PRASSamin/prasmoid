/*
Copyright 2025 PRAS
*/
package cmd

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/PRASSamin/prasmoid/cmd/extendcli"
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

	// time
	timeParse = time.Parse
	timeSince = time.Since
	timeNow   = time.Now

	// net/tls
	tlsDial   = tls.Dial
	connWrite = func(conn *tls.Conn, b []byte) (n int, err error) { return conn.Write(b) }
	connClose = func(conn *tls.Conn) error { return conn.Close() }

	// io
	ioReadAll = io.ReadAll

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
		Title: "Custom Commands:",
	})
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

	err := rootCmdExecute()
	if err != nil {
		osExit(1)
	}
}

// -------- UPDATE CHECKER --------

var CheckForUpdates = func() {
	const host = "api.github.com"
	const path = "/repos/PRASSamin/prasmoid/releases/latest"
	const checkInterval = 24 * time.Hour

	cache, err := readUpdateCache()
	if err == nil {
		if lastCheckedStr, ok := cache["last_checked"].(string); ok {
			lastCheckedTime, err := timeParse(time.RFC3339, lastCheckedStr)
			if err == nil && timeSince(lastCheckedTime) < checkInterval {
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
	conn, err := tlsDial("tcp", host+":443", nil)
	if err != nil {
		return
	}
	defer func() {
		if err := connClose(conn); err != nil {
			logPrintf("Error closing connection: %v", err)
		}
	}()

	// Manually write the HTTP GET request
	request := fmt.Sprintf(
		"GET %s HTTP/1.1\r\nHost: %s\r\nUser-Agent: Prasmoid-Updater\r\nConnection: close\r\n\r\n",
		path, host,
	)
	_, err = connWrite(conn, []byte(request))
	if err != nil {
		return
	}

	// Read the raw HTTP response
	raw, err := ioReadAll(conn)
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

	// parse status code from headers
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

var getLatestTag = func(body []byte) string {
	var tagData map[string]interface{}

	err := jsonUnmarshal(body, &tagData)
	if err != nil {
		return ""
	}

	tag, ok := tagData["tag_name"].(string)
	if !ok {
		return ""
	}

	return strings.TrimPrefix(tag, "v")
}

// compareVersions returns:
// -1 if current < latest
// 0 if current == latest
// 1 if current > latest
var compareVersions = func(current, latest string) int {
	parse := func(v string) []int {
		v = strings.TrimPrefix(v, "v")
		parts := strings.Split(v, ".")
		out := make([]int, 3)
		for i := 0; i < 3 && i < len(parts); i++ {
			num, err := strconv.Atoi(parts[i])
			if err != nil {
				out[i] = 0
			} else {
				out[i] = num
			}
		}
		return out
	}

	curr := parse(current)
	lat := parse(latest)

	for i := 0; i < 3; i++ {
		if curr[i] < lat[i] {
			return -1
		}
		if curr[i] > lat[i] {
			return 1
		}
	}
	return 0
}

var isUpdateAvailable = func(latestTag string) bool {
	if latestTag == "" {
		return false
	}

	current := internalAppMetaDataVersion
	return compareVersions(current, latestTag) < 0
}

var printUpdateMessage = func(latest string) {
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

	fmt.Println(star(bottom))
	fmt.Println(star(printLine(fmt.Sprintf("💠 Prasmoid update available! %s → %s", internalAppMetaDataVersion, latest))))
	fmt.Println(star(printLine("Run `prasmoid upgrade` to update")))
	fmt.Println(star(bottom))
	fmt.Println()
}

var GetCacheFilePath = func() string {
	dir, err := osUserCacheDir()
	if err != nil {
		dir = osTempDir()
	}
	return filepath.Join(dir, "prasmoid_update.json")
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

var writeUpdateCache = func(tag string, body []byte) {
	var releaseData map[string]interface{}
	_ = jsonUnmarshal(body, &releaseData)

	cache := map[string]interface{}{
		"last_checked": timeNow().Format(time.RFC3339),
		"latest_tag":   tag,
		"data":         releaseData,
	}
	data, _ := jsonMarshal(cache)
	_ = osWriteFile(GetCacheFilePath(), data, 0o644)
}
