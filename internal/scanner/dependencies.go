package scanner

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"enva/internal/types"
	"enva/pkg/api/pypi"
)

// ScanDependencies scans packages in virtual environment
func ScanDependencies(venvPath string) ([]types.Dependency, error) {
	var deps []types.Dependency

	// Look for requirements.txt in project root
	projectRoot := findProjectRoot(venvPath)
	reqFile := filepath.Join(projectRoot, "requirements.txt")

	if _, err := os.Stat(reqFile); err == nil {
		// Parse requirements.txt
		parsed, err := parseRequirementsFile(reqFile)
		if err != nil {
			return nil, err
		}

		// Convert to Dependency structs
		for name, version := range parsed {
			dep := types.Dependency{
				Name:    name,
				Version: version,
				Status:  "uptodate", // Default
			}

			// Check if outdated (simplified logic)
			if strings.Contains(version, "==") {
				// Check if this is latest (simplified)
				latest, err := getLatestVersion(name)
				if err != nil {
					// If we can't get latest version from API, skip outdated check
					continue
				}
				dep.Latest = latest
				if latest != "" && latest != strings.TrimPrefix(version, "==") {
					dep.Status = "outdated"
				}
			}

			deps = append(deps, dep)
		}
	}

	return deps, nil
}

// findProjectRoot finds the project root directory from venv path
func findProjectRoot(venvPath string) string {
	// Go up from venv to find project root
	dir := filepath.Dir(venvPath)

	// Look for common project markers
	markers := []string{
		"requirements.txt",
		"pyproject.toml",
		"setup.py",
		"Pipfile",
		".git",
	}

	for dir != "/" {
		for _, marker := range markers {
			if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
				return dir
			}
		}
		dir = filepath.Dir(dir)
	}

	// Fallback to parent of venv
	return filepath.Dir(venvPath)
}

// ParseRequirementsFile parses a requirements.txt file
func ParseRequirementsFile(path string) (map[string]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Skip options like -r, -e, --global-option
		if strings.HasPrefix(line, "-") {
			continue
		}

		// Parse package spec
		if strings.Contains(line, "==") {
			parts := strings.SplitN(line, "==", 2)
			if len(parts) == 2 {
				pkg := strings.TrimSpace(parts[0])
				version := strings.TrimSpace(parts[1])
				result[pkg] = "==" + version
			}
		} else if strings.Contains(line, ">=") {
			parts := strings.SplitN(line, ">=", 2)
			if len(parts) == 2 {
				pkg := strings.TrimSpace(parts[0])
				version := strings.TrimSpace(parts[1])
				result[pkg] = ">=" + version
			}
		} else if strings.Contains(line, "<=") {
			parts := strings.SplitN(line, "<=", 2)
			if len(parts) == 2 {
				pkg := strings.TrimSpace(parts[0])
				version := strings.TrimSpace(parts[1])
				result[pkg] = "<=" + version
			}
		} else if strings.Contains(line, ">") && !strings.Contains(line, ">=") {
			parts := strings.SplitN(line, ">", 2)
			if len(parts) == 2 {
				pkg := strings.TrimSpace(parts[0])
				version := strings.TrimSpace(parts[1])
				result[pkg] = ">" + version
			}
		} else if strings.Contains(line, "<") && !strings.Contains(line, "<=") {
			parts := strings.SplitN(line, "<", 2)
			if len(parts) == 2 {
				pkg := strings.TrimSpace(parts[0])
				version := strings.TrimSpace(parts[1])
				result[pkg] = "<" + version
			}
		} else {
			// Just package name
			result[line] = ""
		}
	}

	return result, nil
}

// getLatestVersion returns the latest version for a package using PyPI API
func getLatestVersion(packageName string) (string, error) {
	client := pypi.NewClient(context.Background(), "", pypi.Default().CacheTTL, pypi.Default().RateLimit, pypi.Default().MaxCacheSize)
	defer client.Close()

	version, err := client.GetLatestVersion(packageName)
	return version, err
}
