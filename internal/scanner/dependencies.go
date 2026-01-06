package scanner

import (
	"os"
	"path/filepath"
	"strings"

	"enva/internal/validator"
)

// ScanDependencies scans packages in virtual environment
func ScanDependencies(venvPath string) ([]validator.Dependency, error) {
	var deps []validator.Dependency

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
			dep := validator.Dependency{
				Name:    name,
				Version: version,
				Status:  "uptodate", // Default
			}

			// Check if outdated (simplified logic)
			if strings.Contains(version, "==") {
				// Check if this is latest (simplified)
				dep.Latest = getLatestVersion(name)
				if dep.Latest != "" && dep.Latest != strings.TrimPrefix(version, "==") {
					dep.Status = "outdated"
				}
			}

			deps = append(deps, dep)
		}
	}

	return deps, nil
}

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

func parseRequirementsFile(path string) (map[string]string, error) {
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

		// Parse package spec
		// Simple parsing for now: package==version
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
		} else {
			// Just package name
			result[line] = ""
		}
	}

	return result, nil
}

func getLatestVersion(packageName string) string {
	// In a real implementation, this would query PyPI API
	// For now, return placeholder
	versionMap := map[string]string{
		"django":       "4.2.0",
		"requests":     "2.31.0",
		"flask":        "2.3.3",
		"numpy":        "1.24.3",
		"pandas":       "2.0.3",
		"tensorflow":   "2.13.0",
		"cryptography": "41.0.0",
	}

	if version, ok := versionMap[packageName]; ok {
		return version
	}

	return "1.0.0" // Default
}
