package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/PRASSamin/prasmoid/internal"
	"github.com/fatih/color"
)

var DIST_DIR, _ = filepath.Abs("./dist/")
var ROOT_DIR, _ = filepath.Abs(".")

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Prasmoid CLI Handler")
		fmt.Println("Usage:")
		fmt.Println("  go run main.go [command]")
		fmt.Println("Available Commands:")
		fmt.Println("  build       Build the cli")
		fmt.Println("  watch       Watch the cli for auto development build")
		return
	}

	switch os.Args[1] {
	case "build":
		os.RemoveAll(DIST_DIR)
		color.Blue("Building cli...")
		var builds = []bool{
			false,
			true,
		}
		for _, build := range builds {
			BuildCli(build)
		}

        // generate checksums
		checksumMap, err := generateChecksums()
		if err != nil {
			color.Red("Failed to generate checksums: %v", err)
			return
		}

		prasmoidChecksum := checksumMap[strings.ToLower(internal.AppMetaData.Name)]
        // update package builds
		if err := updatePackageBuilds(prasmoidChecksum); err != nil {
			color.Red("Failed to update PKGBUILD files: %v", err)
			return
		}

		// update rpm spec file
		if err := updateSpecFile(); err != nil {
			color.Red("Failed to update spec file: %v", err)
			return
		}

	case "watch":
		Watcher()
	default:
		fmt.Println("Prasmoid CLI Handler")
		fmt.Println("Usage:")
		fmt.Println("  go run main.go [command]")
		fmt.Println("Available Commands:")
		fmt.Println("  build       Build the cli")
		fmt.Println("  watch       Watch the cli for auto development build")
	}
}
