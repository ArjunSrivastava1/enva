package venv

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// IsValid checks if a directory is a valid virtual environment
func IsValid(path string) bool {
	if path == "" {
		return false
	}

	// Check if the directory exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	// Check for venv markers
	markers := []string{
		filepath.Join(path, "bin", "python"),
		filepath.Join(path, "bin", "activate"),
		filepath.Join(path, "pyvenv.cfg"),
		filepath.Join(path, "Scripts", "python.exe"),   // Windows
		filepath.Join(path, "Scripts", "activate.bat"), // Windows
	}

	for _, marker := range markers {
		if _, err := os.Stat(marker); err == nil {
			return true
		}
	}

	return false
}

// Detect tries to find a virtual environment automatically
func Detect() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %v", err)
	}

	// Check current and parent directories
	for dir != "/" {
		for _, venvName := range []string{"venv", ".venv", "env", ".env"} {
			venvPath := filepath.Join(dir, venvName)
			if IsValid(venvPath) {
				return venvPath, nil
			}
		}
		dir = filepath.Dir(dir)
	}

	// Check environment variable
	if envPath := os.Getenv("VIRTUAL_ENV"); envPath != "" {
		if IsValid(envPath) {
			return envPath, nil
		}
	}

	return "", fmt.Errorf("no virtual environment found")
}

// GetPythonVersion gets the actual Python version from venv
func GetPythonVersion(venvPath string) (string, error) {
	pythonPath := getPythonExecutable(venvPath)
	if pythonPath == "" {
		return "", fmt.Errorf("Python executable not found in venv")
	}

	cmd := exec.Command(pythonPath, "--version")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to get Python version: %v", err)
	}

	version := strings.TrimSpace(out.String())
	// Output format: "Python 3.9.13"
	if strings.HasPrefix(version, "Python ") {
		version = strings.TrimPrefix(version, "Python ")
	}

	return version, nil
}

// GetPipVersion gets the actual pip version from venv
func GetPipVersion(venvPath string) (string, error) {
	pipPath := getPipExecutable(venvPath)
	if pipPath == "" {
		return "", fmt.Errorf("pip executable not found in venv")
	}

	cmd := exec.Command(pipPath, "--version")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to get pip version: %v", err)
	}

	// Output format: "pip 22.3.1 from /path/to/pip (python 3.9)"
	output := strings.TrimSpace(out.String())
	parts := strings.Fields(output)
	if len(parts) >= 2 {
		return parts[1], nil
	}

	return "unknown", nil
}

// GetInstalledPackages gets actual installed packages
func GetInstalledPackages(venvPath string) (map[string]string, error) {
	pipPath := getPipExecutable(venvPath)
	if pipPath == "" {
		return nil, fmt.Errorf("pip executable not found")
	}

	cmd := exec.Command(pipPath, "list", "--format=freeze")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to list packages: %v", err)
	}

	packages := make(map[string]string)
	output := strings.TrimSpace(out.String())

	if output == "" {
		return packages, nil // Empty venv
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Handle different formats
		if strings.Contains(line, "==") {
			parts := strings.SplitN(line, "==", 2)
			if len(parts) == 2 {
				packageName := strings.TrimSpace(parts[0])
				version := strings.TrimSpace(parts[1])
				packages[packageName] = version
			}
		}
		// Could handle other formats like "@" or editable installs
	}

	return packages, nil
}

// IsActivated checks if venv is currently activated
func IsActivated(venvPath string) bool {
	currentVenv := os.Getenv("VIRTUAL_ENV")
	if currentVenv == "" {
		return false
	}

	// Get absolute paths for comparison
	absVenv, err1 := filepath.Abs(venvPath)
	absCurrent, err2 := filepath.Abs(currentVenv)

	if err1 != nil || err2 != nil {
		return false
	}

	return absVenv == absCurrent
}

// Helper: Get Python executable path
func getPythonExecutable(venvPath string) string {
	possiblePaths := []string{
		filepath.Join(venvPath, "bin", "python"),
		filepath.Join(venvPath, "bin", "python3"),
		filepath.Join(venvPath, "Scripts", "python.exe"),
		filepath.Join(venvPath, "Scripts", "python3.exe"),
	}

	for _, path := range possiblePaths {
		if isExecutable(path) {
			return path
		}
	}

	return ""
}

// Helper: Get pip executable path
func getPipExecutable(venvPath string) string {
	possiblePaths := []string{
		filepath.Join(venvPath, "bin", "pip"),
		filepath.Join(venvPath, "bin", "pip3"),
		filepath.Join(venvPath, "Scripts", "pip.exe"),
		filepath.Join(venvPath, "Scripts", "pip3.exe"),
	}

	for _, path := range possiblePaths {
		if isExecutable(path) {
			return path
		}
	}

	// Fallback: python -m pip
	pythonPath := getPythonExecutable(venvPath)
	if pythonPath != "" {
		return pythonPath
	}

	return ""
}

// Helper: Check if file is executable
func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	// Check executable bit on Unix-like systems
	if runtime.GOOS != "windows" {
		return info.Mode()&0111 != 0
	}

	// Windows: check by extension
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".exe" || ext == ".bat" || ext == ".cmd"
}
