package main

import (
	"flag"
	"fmt"
	"os"

	"enva/internal/detectors"
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
		// Try to detect venv, conda env, or uv environment
		venvPath, err := autoDetectEnvironment()
		if err != nil {
			fmt.Printf("[❌ ERROR] %v\n", err)
			os.Exit(1)
		}
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

// autoDetectEnvironment tries to detect venv, conda env, or uv environment
func autoDetectEnvironment() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	// Look for environment types in current and parent directories
	for dir != "/" {
		// Check for UV environment
		uvLockPath := filepath.Join(dir, "uv.lock")
		if _, err := os.Stat(uvLockPath); err == nil {
			return dir, nil
		}

		// Check for conda environment
		condaMetaPath := filepath.Join(dir, "conda-meta")
		if _, err := os.Stat(condaMetaPath); err == nil {
			return dir, nil
		}

		// Check for virtual environment
		venvPath := filepath.Join(dir, "venv")
		if detectors.VenvIsValid(venvPath) {
			return dir, nil
		}

		// Also check .venv, env, .env
		for _, venvName := range []string{".venv", "env", ".env"} {
			venvPath := filepath.Join(dir, venvName)
			if detectors.VenvIsValid(venvPath) {
				return dir, nil
			}
		}

		dir = filepath.Dir(dir)
	}

	return "", fmt.Errorf("no virtual environment, conda environment, or UV environment found")
}

func printHelp() {
	fmt.Println(`
🌿 enva - Environment Validator
───────────────────────────────

Usage: enva [options]

Options:
  --venv PATH       Path to virtual/conda/uv environment (auto-detect if not specified)
  --json            Output in JSON format (for CI/CD)
  --version         Show version
  --help            Show this help

Examples:
  enva                         # Auto-detect and validate environment
  enva --venv ./venv           # Validate specific venv
  enva --venv ./myenv          # Validate specific conda or uv env
  enva --json                  # JSON output for automation

Currently supports:
  • Python virtual environments
  • Conda environments
  • UV environments
  • Basic dependency checking
  • Security vulnerability scanning
  • Performance optimization suggestions`)
}
