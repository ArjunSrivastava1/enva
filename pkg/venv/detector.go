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

// IsValid checks if the given path is a valid Python virtual environment
func IsValid(path string) bool {
	if path == "" {
		return false
	}

	// Check if the directory exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}

	// First, check for pyvenv.cfg with valid content
	cfgPath := filepath.Join(path, "pyvenv.cfg")
	if _, err := os.Stat(cfgPath); err == nil {
		// Check if pyvenv.cfg has valid content
		if hasValidVenvConfig(cfgPath) {
			return true
		}
		// pyvenv.cfg exists but is invalid/empty - continue checking other markers
	}

	// Check for OS-specific markers that are non-empty
	markers := getVenvMarkers()
	validFound := false
	for _, marker := range markers {
		if _, err := os.Stat(marker); err == nil {
			// Check that the file is not completely empty
			info, err := os.Stat(marker)
			if err == nil && info.Size() > 0 {
				validFound = true
				break
			}
		}
	}

	return validFound
}

// hasValidVenvConfig checks if pyvenv.cfg has a valid [venv] section
func hasValidVenvConfig(cfgPath string) bool {
	content, err := os.ReadFile(cfgPath)
	if err != nil {
		return false
	}

	text := string(content)

	// Check for [venv] section
	if !strings.Contains(text, "[venv]") {
		return false
	}

	// Extract the [venv] section
	lines := strings.Split(text, "\n")
	inVenvSection := false
	hasHome := false
	hasInclude := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Section header
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			if inVenvSection && hasHome && hasInclude {
				// Found all required fields in [venv] section
				return true
			}
			if line == "[venv]" {
				inVenvSection = true
				hasHome = false
				hasInclude = false
			}
			inVenvSection = false
			continue
		}

		if inVenvSection {
			// Parse key=value pairs
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				// Remove quotes if present
				value = strings.Trim(value, "\"'")

				if key == "home" {
					hasHome = true
				} else if key == "include" || key == "include_system_site_packages" {
					hasInclude = true
				}
			}
		}
	}

	// Also check top-level keys (before [venv] section)
	if hasHome || hasInclude {
		return true
	}

	return false
}

// Detect tries to find a virtual environment automatically
func Detect() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
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
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to get Python version: %w (stderr: %s)", err, stderr.String())
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
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("failed to get pip version: %w (stderr: %s)", err, stderr.String())
	}

	// Output format: "pip 22.3.1 from /path/to/pip (python 3.9)"
	output := strings.TrimSpace(out.String())
	parts := strings.Fields(output)
	if len(parts) >= 2 {
		return parts[1], nil
	}

	return "unknown", nil
}

// GetInstalledPackages gets actual installed packages from pip or pyvenv.cfg
func GetInstalledPackages(venvPath string) (map[string]string, error) {
	// First, try to get packages via pip
	pipPath := getPipExecutable(venvPath)
	if pipPath != "" {
		packages := map[string]string{}

		// Try direct pip list
		if cmd := exec.Command(pipPath, "list", "--format=freeze"); cmd.Run() == nil {
			var out, stderr bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &stderr

			if err := cmd.Run(); err != nil {
				// Fall through to pyvenv.cfg parsing
			} else {
				output := strings.TrimSpace(out.String())
				if output != "" {
					return parsePipFreezeOutput(output)
				}
			}
		}

		// Try pip list --path as fallback
		if cmd := exec.Command(pipPath, "list", "--path"); cmd.Run() == nil {
			var out, stderr bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &stderr

			if err := cmd.Run(); err != nil {
				// Fall through to pyvenv.cfg parsing
			} else {
				pathOutput := strings.TrimSpace(out.String())
				if pathOutput != "" {
					packages := parsePipPathOutput(pathOutput, venvPath)
					if len(packages) > 0 {
						return packages, nil
					}
				}
			}
		}
	}

	// Fallback: parse pyvenv.cfg [packages] section
	cfgPath := filepath.Join(venvPath, "pyvenv.cfg")
	if packages := parsePackagesFromCfg(cfgPath); len(packages) > 0 {
		return packages, nil
	}

	return map[string]string{}, nil
}

// parsePipFreezeOutput parses pip list --format=freeze output
func parsePipFreezeOutput(output string) (map[string]string, error) {
	packages := make(map[string]string)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Skip pip itself
		if strings.HasPrefix(line, "pip") {
			continue
		}

		// Skip editable installs for version tracking
		if strings.HasPrefix(line, "-e ") {
			parts := strings.TrimPrefix(line, "-e ")
			packageName := strings.TrimSpace(parts)
			packages[packageName] = "editable"
			continue
		}

		// Skip direct URL references
		if strings.HasPrefix(line, "-r ") || strings.Contains(line, " @ ") {
			continue
		}

		// Parse package==version
		if strings.Contains(line, "==") {
			parts := strings.SplitN(line, "==", 2)
			if len(parts) == 2 {
				packageName := strings.TrimSpace(parts[0])
				version := strings.TrimSpace(parts[1])
				packages[packageName] = version
				continue
			}
		}

		// Parse package>=version, package<=version, etc.
		if strings.Contains(line, ">=") && !strings.Contains(line, ">=") {
			parts := strings.SplitN(line, ">", 2)
			if len(parts) == 2 {
				packageName := strings.TrimSpace(parts[0])
				version := strings.TrimSpace(parts[1])
				packages[packageName] = version
				continue
			}
		}

		if strings.Contains(line, "<=") && !strings.Contains(line, ">=") {
			parts := strings.SplitN(line, "<=", 2)
			if len(parts) == 2 {
				packageName := strings.TrimSpace(parts[0])
				version := strings.TrimSpace(parts[1])
				packages[packageName] = version
				continue
			}
		}

		if strings.Contains(line, ">=") {
			parts := strings.SplitN(line, ">=", 2)
			if len(parts) == 2 {
				packageName := strings.TrimSpace(parts[0])
				version := strings.TrimSpace(parts[1])
				packages[packageName] = ">= " + version
				continue
			}
		}

		if strings.Contains(line, "<=") {
			parts := strings.SplitN(line, "<=", 2)
			if len(parts) == 2 {
				packageName := strings.TrimSpace(parts[0])
				version := strings.TrimSpace(parts[1])
				packages[packageName] = "<= " + version
				continue
			}
		}

		// Just package name
		packages[line] = ""
	}

	return packages, nil
}

// parsePipPathOutput parses pip list --path output
func parsePipPathOutput(output string, venvPath string) map[string]string {
	packages := make(map[string]string)
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse paths like: ./bin/package1-1.0.0.dist-info, ./bin/package2-2.0.0.dist-info
		entries := strings.Split(line, ",")
		for _, entry := range entries {
			entry = strings.TrimSpace(entry)
			if strings.HasSuffix(entry, ".dist-info") || strings.HasSuffix(entry, ".egg-info") {
				// Extract package name from dist-info folder
				folder := strings.TrimSuffix(entry, ".dist-info")
				folder = strings.TrimSuffix(folder, ".egg-info")
				pkgName := strings.Split(folder, "-")[0]
				packages[pkgName] = "installed"
			}
		}
	}

	return packages
}

// parsePackagesFromCfg parses [packages] section from pyvenv.cfg
func parsePackagesFromCfg(cfgPath string) map[string]string {
	packages := make(map[string]string)

	content, err := os.ReadFile(cfgPath)
	if err != nil {
		return packages
	}

	text := string(content)
	lines := strings.Split(text, "\n")

	inPackagesSection := false
	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Section headers
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			if line == "[packages]" {
				inPackagesSection = true
			} else {
				inPackagesSection = false
			}
			continue
		}

		if inPackagesSection {
			// Parse package = version lines
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				packageName := strings.TrimSpace(parts[0])
				version := strings.TrimSpace(parts[1])
				// Remove quotes
				version = strings.Trim(version, "\"'")
				packages[packageName] = version
			}
		}
	}

	return packages
}

// IsActivated checks if the specified venv is currently activated
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

// IsActivatedFromConfig checks if a venv is activated by checking VIRTUALENV_PROMPT_VAR in pyvenv.cfg
func IsActivatedFromConfig(venvPath string) bool {
	cfgPath := filepath.Join(venvPath, "pyvenv.cfg")
	content, err := os.ReadFile(cfgPath)
	if err != nil {
		return false
	}

	text := string(content)
	return strings.Contains(text, "VIRTUALENV_PROMPT_VAR=")
}

// Helper: Get Python executable path from pyvenv.cfg or filesystem
func getPythonExecutable(venvPath string) string {
	// First, try to get Python path from pyvenv.cfg
	cfgPath := filepath.Join(venvPath, "pyvenv.cfg")
	if pythonPathFromCfg, err := getPythonPathFromCfg(cfgPath); err == nil && pythonPathFromCfg != "" {
		return filepath.Join(venvPath, pythonPathFromCfg)
	}

	// Fallback: check filesystem for executable
	possiblePaths := []string{
		filepath.Join(venvPath, "bin", "python"),
		filepath.Join(venvPath, "bin", "python3"),
		filepath.Join(venvPath, "Scripts", "python.exe"),
		filepath.Join(venvPath, "Scripts", "python3.exe"),
	}

	for _, path := range possiblePaths {
		if isExecutableAndValid(path) {
			return path
		}
	}

	return ""
}

// getPythonPathFromCfg extracts python.path from pyvenv.cfg
func getPythonPathFromCfg(cfgPath string) (string, error) {
	content, err := os.ReadFile(cfgPath)
	if err != nil {
		return "", err
	}

	text := string(content)

	// Look for python.path = ...
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "python.path=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				path := strings.TrimSpace(parts[1])
				path = strings.Trim(path, "\"'")
				return path
			}
		}
	}

	return "", fmt.Errorf("python.path not found in pyvenv.cfg")
}

// isExecutableAndValid checks if a path is executable and not empty
func isExecutableAndValid(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	// Check that the file is not empty
	if info.Size() == 0 {
		return false
	}

	// Check executable bit on Unix-like systems
	if runtime.GOOS != "windows" {
		return info.Mode()&0111 != 0
	}

	// Windows: check by extension
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".exe" || ext == ".bat" || ext == ".cmd" || ext == ""
}

// Helper: Get pip executable path
func getPipExecutable(venvPath string) string {
	// Try filesystem first
	possiblePaths := []string{
		filepath.Join(venvPath, "bin", "pip"),
		filepath.Join(venvPath, "bin", "pip3"),
		filepath.Join(venvPath, "Scripts", "pip.exe"),
		filepath.Join(venvPath, "Scripts", "pip3.exe"),
	}

	for _, path := range possiblePaths {
		if isExecutableAndValid(path) {
			return path
		}
	}

	// Fallback: use python -m pip
	pythonPath := getPythonExecutable(venvPath)
	if pythonPath != "" {
		// Try "pip" module via python -m pip
		return filepath.Join(pythonPath, "__main__.py")
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

// Helper: Get venv markers based on OS
func getVenvMarkers() []string {
	markers := []string{
		filepath.Join("pyvenv.cfg"),
	}

	if runtime.GOOS == "windows" {
		markers = append(markers,
			filepath.Join("Scripts", "python.exe"),
			filepath.Join("Scripts", "activate.bat"),
		)
	} else {
		markers = append(markers,
			filepath.Join("bin", "python"),
			filepath.Join("bin", "activate"),
		)
	}

	return markers
}
