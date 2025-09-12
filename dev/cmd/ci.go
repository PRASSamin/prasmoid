/*
Copyright 2025 PRAS
Development commands for Prasmoid
*/
package cmd

import (
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(ciCmd)
}

var ciCmd = &cobra.Command{
	Use:   "ci",
	Short: "Run all CI checks locally.",
	Long: `This command runs the same checks that are executed in the GitHub Actions CI workflow.
It is recommended to run this command before pushing changes to a pull request.

This command performs the following checks:
1. go vet
2. go test with race detector and coverage
3. coverage check (minimum 80%)
4. golangci-lint (will be installed automatically if not found)

Please ensure you have the following dependencies installed:
- gettext
- plasma-sdk`,
	Run: func(cmd *cobra.Command, args []string) {
		color.Yellow("Running CI checks...")

		// Vet
		color.Yellow("\n--- Running go vet ---")
		vetCmd := exec.Command("go", "vet", "./...")
		vetCmd.Stdout = os.Stdout
		vetCmd.Stderr = os.Stderr
		if err := vetCmd.Run(); err != nil {
			color.Red("\nError: go vet failed: %v", err)
			os.Exit(1)
		}
		color.Green("--- go vet passed ---")

		// Test
		color.Yellow("\n--- Running tests ---")
		testCmd := exec.Command("go", "test", "./cmd/...", "./internal/...", "./utils/...", "-v", "-race", "-coverprofile=coverage.out", "-covermode=atomic")
		testCmd.Stdout = os.Stdout
		testCmd.Stderr = os.Stderr
		if err := testCmd.Run(); err != nil {
			color.Red("\nError: go test failed: %v", err)
			os.Exit(1)
		}
		color.Green("--- go test passed ---")

		// Coverage Check
		color.Yellow("\n--- Checking coverage ---")
		coverCmd := exec.Command("go", "tool", "cover", "-func=coverage.out")
		output, err := coverCmd.Output()
		if err != nil {
			color.Red("Error: could not run go tool cover: %v", err)
			os.Exit(1)
		}

		lines := strings.Split(string(output), "\n")
		var totalLine string
		for i := len(lines) - 1; i >= 0; i-- {
			if strings.HasPrefix(lines[i], "total:") {
				totalLine = lines[i]
				break
			}
		}

		if totalLine == "" {
			color.Red("Error: could not find total coverage in output.")
			os.Exit(1)
		}

		fields := strings.Fields(totalLine)
		if len(fields) < 3 {
			color.Red("Error: could not parse total coverage line.")
			os.Exit(1)
		}
		coverageStr := strings.TrimSuffix(fields[len(fields)-1], "%")
		coverage, err := strconv.ParseFloat(coverageStr, 64)
		if err != nil {
			color.Red("Error: could not parse coverage percentage: %v", err)
			os.Exit(1)
		}

		minCoverage := 80.0
		color.Cyan("Total coverage: %.1f%%", coverage)
		if coverage < minCoverage {
			color.Red("Error: Coverage is below %.1f%%", minCoverage)
			os.Exit(1)
		}
		color.Green("--- Coverage check passed ---")

		// Lint
		color.Yellow("\n--- Running golangci-lint ---")
		if _, err := exec.LookPath("golangci-lint"); err != nil {
			color.Yellow("golangci-lint not found, attempting to install...")
			installCmd := "curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.4.0"
			installExec := exec.Command("sh", "-c", installCmd)
			installExec.Stdout = os.Stdout
			installExec.Stderr = os.Stderr
			if err := installExec.Run(); err != nil {
				color.Red("\nError: failed to install golangci-lint: %v", err)
				os.Exit(1)
			}
			color.Green("golangci-lint installed successfully.")
		}
		lintCmd := exec.Command("golangci-lint", "run", "--timeout=5m")
		lintCmd.Stdout = os.Stdout
		lintCmd.Stderr = os.Stderr
		if err := lintCmd.Run(); err != nil {
			color.Red("\nError: golangci-lint failed: %v", err)
			os.Exit(1)
		}
		color.Green("--- golangci-lint passed ---")

		color.Green("\nAll CI checks passed!")
	},
}
