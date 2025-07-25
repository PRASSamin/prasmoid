/*
Copyright 2025 PRAS
*/
package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strings"

	"github.com/PRASSamin/prasmoid/internal"
	"github.com/fatih/color"
	"github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
)

func init() {
	updateCmd.AddCommand(meCmd)
}

var meCmd = &cobra.Command{
	Use:   "me",
	Short: "Update Prasmoid CLI",
	Long:  "Update Prasmoid CLI to the latest version from GitHub releases.",
	Run: func(cmd *cobra.Command, args []string) {
		if err := checkRoot(); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		
		name := strings.ToLower(internal.AppMeta.Name)
		current := internal.AppMeta.Version
		rawCurrent := strings.ReplaceAll(current, "-compressed", "")

		cache, err := ReadUpdateCache()
		if err != nil {
			logFail(fmt.Sprintf("Failed to read update cache: %v", err))
			return
		}

		latest := strings.Replace(cache.Data.Tag_name, "v", "", 1)

		if latest == rawCurrent {
			fmt.Println("You are already using the latest version of Prasmoid(v" + rawCurrent + ")")
			return
		} 

		logStep("Updating Prasmoid")

		logStep(fmt.Sprintf("Downloading v%s [%s]", latest, osArch()))

		// Pick download URL
		var downloadUrl string
		for _, asset := range cache.Data.Assets {
			if isCompressedVersion(latest) {
				if strings.HasPrefix(asset.Name, name) && strings.HasSuffix(asset.Name, "-compressed") {
					downloadUrl = asset.BrowserDownloadURL
					break
				}
			} else {
				if strings.HasPrefix(asset.Name, name) && !strings.HasSuffix(asset.Name, "-compressed") {
					downloadUrl = asset.BrowserDownloadURL
					break
				}
			}
		}

		if downloadUrl == "" {
			logFail(fmt.Sprintf("No suitable binary found for version %s", latest))
			return
		}

		resp, err := http.Get(downloadUrl)
		if err != nil {
			logFail(fmt.Sprintf("Download error: %v", err))
			return
		}
		defer resp.Body.Close()

		tempFile, err := os.CreateTemp("", "prasmoid-update-*")
		if err != nil {
			logFail(fmt.Sprintf("Temp file error: %v", err))
			return
		}
		defer tempFile.Close()

		clen := resp.ContentLength
		if clen <= 0 {
			logWarn("Unknown file size, progress bar may be weird.")
		}

		bar := progressbar.DefaultBytes(clen)

		_, err = io.Copy(io.MultiWriter(tempFile, bar), resp.Body)
		if err != nil {
			logFail(fmt.Sprintf("Failed to write binary: %v", err))
			return
		}
		logDone("✔ Download complete")

		if err := os.Chmod(tempFile.Name(), 0755); err != nil {
			logFail(fmt.Sprintf("chmod failed: %v", err))
			return
		}

		currentExe, err := os.Executable()
		if err != nil {
			logFail(fmt.Sprintf("Can't find current binary: %v", err))
			return
		}

		logStep("Updating")
		script := fmt.Sprintf(`
#!/bin/bash
sleep 1
echo "↳ Moving new binary..."
mv "%s" "%s"
chmod +x "%s"
echo "✔ Update complete."
echo "↳ Relaunching CLI..."
"%s" %s
`, tempFile.Name(), currentExe, currentExe, currentExe, strings.Join(os.Args[1:], " "))

		updateScriptPath := tempFile.Name() + ".sh"
		if err := os.WriteFile(updateScriptPath, []byte(script), 0755); err != nil {
			logFail(fmt.Sprintf("Script write error: %v", err))
			return
		}

		if err := exec.Command("sh", updateScriptPath).Start(); err != nil {
			logFail(fmt.Sprintf("Script exec failed: %v", err))
			return
		}

		logDone("✔ Update complete")
		os.Exit(0)
	},
}

func isCompressedVersion(version string) bool {
	return strings.Contains(version, "-compressed")
}

func osArch() string {
	return fmt.Sprintf("%s/%s", runtime.GOOS, arch())
}

func arch() string {
	var uname unix.Utsname
		if err := unix.Uname(&uname); err != nil {
			return "unknown"
		}
		var machine []byte
		for _, c := range uname.Machine {
			if c == 0 {
				break
			}
			machine = append(machine, byte(c))
		}
		return string(machine)
}

func checkRoot() error {
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current user: %v", err)
	}

	if currentUser.Uid != "0" {
		return fmt.Errorf("the requested operation requires superuser privileges. use `sudo %s`", strings.Join(os.Args[0:], " "))
	}
	return nil
}

func logStep(step string) {
	color.Cyan("→ %s", step)
}

func logDone(msg string) {
	color.Green("%s", msg)
}

func logWarn(msg string) {
	color.Yellow("%s", msg)
}

func logFail(msg string) {
	color.Red("✘ %s", msg)
}

