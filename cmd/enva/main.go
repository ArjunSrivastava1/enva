package main

import (
	"flag"
	"fmt"
	"os"

	"enva/internal/formatter"
	"enva/internal/validator"
)

func main() {
	var (
		venvPath    string
		jsonOutput  bool
		showHelp    bool
		showVersion bool
	)

	flag.StringVar(&venvPath, "venv", "", "Path to virtual environment")
	flag.BoolVar(&jsonOutput, "json", false, "Output JSON format")
	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.BoolVar(&showVersion, "version", false, "Show version")
	flag.Parse()

	if showVersion {
		fmt.Println("enva v0.1.0")
		return
	}

	if showHelp {
		printHelp()
		return
	}

	// If no venv specified, try to auto-detect
	if venvPath == "" {
		fmt.Println("[ℹ️] No venv specified, trying auto-detection...")
		// For now, use a dummy path - we'll implement detection later
		venvPath = "./venv"
	}

	// Validate the environment
	result, err := validator.ValidateEnvironment(venvPath)
	if err != nil {
		fmt.Printf("[❌ ERROR] %v\n", err)
		os.Exit(1)
	}

	// Format output
	var output string
	if jsonOutput {
		output = formatter.FormatJSON(result)
	} else {
		output = formatter.FormatChinese(result)
	}

	fmt.Print(output)

	// Exit with error code if validation failed
	if result.OverallStatus == "error" {
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`
🌿 enva - Environment Validator
───────────────────────────────

Usage: enva [options]

Options:
  --venv PATH    Path to virtual environment
  --json         Output in JSON format (for CI/CD)
  --version      Show version
  --help         Show this help

Examples:
  enva                         # Validate current environment
  enva --venv ./venv           # Validate specific venv
  enva --json                  # JSON output for automation

Currently supports:
  • Python virtual environments
  • Basic dependency checking
  • Security vulnerability scanning
  • Performance optimization suggestions`)
}
