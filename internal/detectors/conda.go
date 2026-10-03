package detectors

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"enva/internal/types"
)

// CondaDetector handles conda package and environment detection
type CondaDetector struct {
	venvPath     string
	condaEnvPath string
	packages     map[string]string
	condaHistory []condaHistoryEntry
	envs         []types.CondaEnv
	packageName  string
}

// CondaEnv represents a conda environment
type CondaEnv struct {
	Name          string    `json:"name"`
	Path          string    `json:"path"`
	PackageCount  int       `json:"package_count"`
	PythonVersion string    `json:"python_version,omitempty"`
	Packages      []string  `json:"packages,omitempty"`
	CreationTime  time.Time `json:"creation_time"`
}

// condaHistoryEntry represents an entry in conda-meta/history
type condaHistoryEntry struct {
	Package    string    `json:"package"`
	Version    string    `json:"version"`
	OldVersion string    `json:"old_version"`
	Build      string    `json:"build"`
	Desc       string    `json:"desc"`
	Time       time.Time `json:"time"`
}

// NewCondaDetector creates a new conda detector
func NewCondaDetector(venvPath string) *CondaDetector {
	return &CondaDetector{
		venvPath:     venvPath,
		packages:     make(map[string]string),
		condaHistory: []condaHistoryEntry{},
	}
}

// Detect scans a conda environment and returns CondaEnv information
func (d *CondaDetector) Detect() (*types.CondaEnv, error) {
	env := &types.CondaEnv{
		Name:         filepath.Base(d.venvPath),
		Path:         d.venvPath,
		Packages:     []string{},
		CreationTime: time.Time{},
	}

	// Check if conda-meta directory exists
	condaMetaPath := filepath.Join(d.venvPath, "conda-meta")
	if _, err := os.Stat(condaMetaPath); os.IsNotExist(err) {
		// Not a conda environment
		return nil, nil
	}

	// Parse conda history file
	if err := d.parseCondaHistory(condaMetaPath); err != nil {
		return nil, fmt.Errorf("failed to parse conda history: %w", err)
	}

	// Parse package list from conda-meta
	if err := d.parseCondaPackages(condaMetaPath); err != nil {
		return nil, fmt.Errorf("failed to parse conda packages: %w", err)
	}

	// Detect python version
	pythonVersion := d.detectPythonVersion()

	// Detect creation time
	creationTime := d.detectCreationTime()

	env.PythonVersion = pythonVersion
	env.Packages = d.packagesToSlice()

	return env, nil
}

// parseCondaHistory parses the conda-meta/history file
func (d *CondaDetector) parseCondaHistory(condaMetaPath string) error {
	historyFile := filepath.Join(condaMetaPath, "history")

	file, err := os.Open(historyFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // history file doesn't exist, that's okay
		}
		return err
	}
	defer file.Close()

	// Parse history file
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		entry, err := d.parseHistoryLine(line)
		if err != nil {
			continue // Skip malformed lines
		}

		d.condaHistory = append(d.condaHistory, entry)
	}

	return scanner.Err()
}

// parseHistoryLine parses a single line from conda history
func (d *CondaDetector) parseHistoryLine(line string) (condaHistoryEntry, error) {
	entry := condaHistoryEntry{
		Time: time.Now(),
	}

	// Format: <timestamp> <package>[=<version>][?]<build> <old_version> <desc>
	// Example: 2024-01-01 12:00:00 numpy==1.21.0 py39_0 1.19.0 "numpy 1.21.0"

	// Split by spaces, but handle versions carefully
	parts := d.splitHistoryLine(line)

	if len(parts) < 5 {
		return entry, fmt.Errorf("invalid history line format")
	}

	// Parse timestamp
	timestampStr := strings.TrimSpace(parts[0])
	if ts, err := time.Parse("2006-01-02 15:04:05", timestampStr); err == nil {
		entry.Time = ts
	}

	// Parse package and version (parts[1] could be like "package==1.0.0" or "package")
	pkgSpec := strings.TrimSpace(parts[1])
	d.extractPackageNameAndVersion(pkgSpec)

	// Parse build string
	entry.Build = strings.TrimSpace(parts[2])

	// Parse old version
	entry.OldVersion = strings.TrimSpace(parts[3])

	// Parse description
	entry.Desc = strings.Join(parts[4:], " ")

	return entry, nil
}

// splitHistoryLine splits history line by spaces, handling version specs
func (d *CondaDetector) splitHistoryLine(line string) []string {
	// First, mark version spec positions
	versionSpec := regexp.MustCompile(`([a-zA-Z0-9_.-]+(?:=[<>=<>!]+[a-zA-Z0-9_.-]*)?)`)
	matches := versionSpec.FindAllStringIndex(line, -1)

	// Split by spaces, then recombine where version specs span multiple parts
	result := []string{}
	lastEnd := 0

	for _, match := range matches {
		start := match[0]
		end := match[1]

		// Add text before match
		if start > lastEnd {
			result = append(result, line[lastEnd:start])
		}
		result = append(result, line[start:end])
		lastEnd = end
	}

	// Add remaining text
	if lastEnd < len(line) {
		result = append(result, line[lastEnd:])
	}

	return result
}

// extractPackageNameAndVersion extracts package name and version from package spec
func (d *CondaDetector) extractPackageNameAndVersion(pkgSpec string) {
	// Handle package==version, package>=version, etc.
	if strings.Contains(pkgSpec, "==") {
		parts := strings.SplitN(pkgSpec, "==", 2)
		d.packageName = strings.TrimSpace(parts[0])
		d.packages[d.packageName] = parts[1]
	} else if strings.Contains(pkgSpec, ">=") {
		parts := strings.SplitN(pkgSpec, ">=", 2)
		d.packageName = strings.TrimSpace(parts[0])
		d.packages[d.packageName] = parts[1]
	} else if strings.Contains(pkgSpec, "<=") {
		parts := strings.SplitN(pkgSpec, "<=", 2)
		d.packageName = strings.TrimSpace(parts[0])
		d.packages[d.packageName] = parts[1]
	} else if strings.Contains(pkgSpec, ">") && !strings.Contains(pkgSpec, ">=") {
		parts := strings.SplitN(pkgSpec, ">", 2)
		d.packageName = strings.TrimSpace(parts[0])
		d.packages[d.packageName] = parts[1]
	} else if strings.Contains(pkgSpec, "<") && !strings.Contains(pkgSpec, "<=") {
		parts := strings.SplitN(pkgSpec, "<", 2)
		d.packageName = strings.TrimSpace(parts[0])
		d.packages[d.packageName] = parts[1]
	} else if strings.Contains(pkgSpec, "[") {
		// Handle version ranges like "package[version='1.0.0']"
		d.packageName = pkgSpec
	} else {
		// No version specified
		d.packageName = pkgSpec
		d.packages[d.packageName] = ""
	}
}

// parseCondaPackages parses conda package metadata files
func (d *CondaDetector) parseCondaPackages(condaMetaPath string) error {
	// Get list of package files in conda-meta
	packageFiles, err := os.ReadDir(condaMetaPath)
	if err != nil {
		return err
	}

	// Filter for package files (not repodata.json or other metadata)
	for _, file := range packageFiles {
		if !file.IsDir() && !strings.HasPrefix(file.Name(), "repodata-") {
			if err := d.parsePackageFile(filepath.Join(condaMetaPath, file.Name())); err != nil {
				// Log but continue
				continue
			}
		}
	}

	return nil
}

// parsePackageFile parses a single conda package metadata file
func (d *CondaDetector) parsePackageFile(pkgPath string) error {
	file, err := os.Open(pkgPath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Read JSON metadata
	var metadata map[string]interface{}
	if err := json.NewDecoder(file).Decode(&metadata); err != nil {
		return err
	}

	// Extract package name
	name, ok := metadata["name"].(string)
	if !ok {
		return fmt.Errorf("invalid package metadata: missing name")
	}
	d.packageName = name
	d.packages[name] = ""

	// Extract version
	version, ok := metadata["version"].(string)
	if ok {
		d.packages[name] = version
	}

	// Extract build string
	if build, ok := metadata["build"].(string); ok {
		d.packages[name] = d.packages[name] + " " + build
	}

	// Extract build number
	if buildNum, ok := metadata["build_number"].(float64); ok {
		d.packages[name] = d.packages[name] + "[build=" + strconv.Itoa(int(buildNum)) + "]"
	}

	// Extract timestamp for creation time detection
	if timestamp, ok := metadata["timestamp"].(float64); ok {
		d.creationTime = time.Unix(int64(timestamp), 0)
	}

	// Extract install time
	if installTime, ok := metadata["installed"].(float64); ok {
		d.creationTime = time.Unix(int64(installTime), 0)
	}

	// Extract python version from dependencies
	dependencies, ok := metadata["dependencies"].([]interface{})
	if ok {
		if pythonDep := d.extractPythonVersionFromDeps(dependencies); pythonDep != "" {
			d.pythonVersion = pythonDep
		}
	}

	return nil
}

// extractPythonVersionFromDeps extracts Python version from dependencies
func (d *CondaDetector) extractPythonVersionFromDeps(deps []interface{}) string {
	for _, dep := range deps {
		if depStr, ok := dep.(string); ok {
			if strings.HasPrefix(depStr, "python") {
				// Handle python=3.9, python>=3.8, etc.
				if strings.Contains(depStr, "==") {
					parts := strings.SplitN(depStr, "==", 2)
					return strings.TrimSpace(parts[1])
				} else if strings.Contains(depStr, ">=") {
					parts := strings.SplitN(depStr, ">=", 2)
					return strings.TrimSpace(parts[1])
				} else if strings.Contains(depStr, "<=") {
					parts := strings.SplitN(depStr, "<=", 2)
					return strings.TrimSpace(parts[1])
				} else if strings.Contains(depStr, ">") && !strings.Contains(depStr, ">=") {
					parts := strings.SplitN(depStr, ">", 2)
					return strings.TrimSpace(parts[1])
				} else if strings.Contains(depStr, "<") && !strings.Contains(depStr, "<=") {
					parts := strings.SplitN(depStr, "<", 2)
					return strings.TrimSpace(parts[1])
				} else {
					return strings.TrimSpace(depStr)
				}
			}
		}
	}
	return ""
}

// detectPythonVersion detects Python version from various sources
func (d *CondaDetector) detectPythonVersion() string {
	// First check if we have it from package metadata
	if d.pythonVersion != "" {
		return d.pythonVersion
	}

	// Check conda-meta/python package
	pythonMeta := filepath.Join(d.venvPath, "conda-meta", "python")
	if info, err := os.Stat(pythonMeta); err == nil && !info.IsDir() {
		// Parse the python package metadata
		d.parsePackageFile(pythonMeta)
		return d.packages["python"]
	}

	// Check conda-meta/history for python package
	for _, entry := range d.condaHistory {
		if entry.Package == "python" {
			return entry.Version
		}
	}

	return ""
}

// detectCreationTime detects the creation time of the conda environment
func (d *CondaDetector) detectCreationTime() time.Time {
	// Use the earliest timestamp from conda history
	if len(d.condaHistory) > 0 {
		minTime := d.condaHistory[0].Time
		for _, entry := range d.condaHistory {
			if entry.Time.Before(minTime) {
				minTime = entry.Time
			}
		}
		return minTime
	}

	// Fallback: use current time
	return time.Now()
}

// packagesToSlice converts the packages map to a slice
func (d *CondaDetector) packagesToSlice() []string {
	result := make([]string, 0, len(d.packages))
	for name := range d.packages {
		result = append(result, name)
	}
	return result
}

// GetPackages returns all detected packages
func (d *CondaDetector) GetPackages() []types.Dependency {
	// This can be used by the scanner to get conda packages
	// For now, we return empty and rely on the conda env info
	return nil
}

// GetCondaHistory returns the conda history entries
func (d *CondaDetector) GetCondaHistory() []condaHistoryEntry {
	return d.condaHistory
}
