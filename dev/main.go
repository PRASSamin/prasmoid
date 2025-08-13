package main

import (
	"fmt"
	"os"
	"path/filepath"

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
		if err := os.RemoveAll(DIST_DIR); err != nil {
			color.Red("Warning: Failed to remove distribution directory: %v", err)
		}
		color.Blue("Building cli...")
		var builds = []bool{
			false,
			true,
		}
		for _, build := range builds {
			BuildCli(build)
		}

		// generate checksums
		_, err := generateChecksums()
		if err != nil {
			color.Red("Failed to generate checksums: %v", err)
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
