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

// UVDetector handles UV package and environment detection
type UVDetector struct {
	uvPath            string
	pythonVersionFile string
	uvLockFile        string
	packages          map[string]string
	pythonVersions    []string
	envs              []types.UVEnv
}

// UVEnv represents a UV environment
type UVEnv struct {
	Name            string    `json:"name"`
	Path            string    `json:"path"`
	PackageCount    int       `json:"package_count"`
	PythonVersion   string    `json:"python_version,omitempty"`
	Packages        []string  `json:"packages,omitempty"`
	CreationTime    time.Time `json:"creation_time"`
}

// UVLockPackage represents a package entry in uv.lock
type UVLockPackage struct {
	Name      string   `json:"name"`
	Version   string   `json:"version"`
	Source    string   `json:"source,omitempty"`
	Directory string   `json:"directory,omitempty"`
	Dev       bool     `json:"dev,omitempty"`
	Manifest  string   `json:"manifest,omitempty"`
	Dependencies []struct {
		Name     string `json:"name"`
		Version  string `json:"version,omitempty"`
		Inner    string `json:"inner,omitempty"`
		Optional bool   `json:"optional,omitempty"`
	} `json:"dependencies,omitempty"`
}

// NewUVDetector creates a new UV detector
func NewUVDetector(uvPath string) *UVDetector {
	return &UVDetector{
		uvPath:            uvPath,
		packages:          make(map[string]string),
		pythonVersions:    []string{},
		envs:              []types.UVEnv{},
	}
}

// Detect scans a UV environment and returns UVEnv information
func (d *UVDetector) Detect() (*types.UVEnv, error) {
	env := &types.UVEnv{
		Name:            filepath.Base(d.uvPath),
		Path:            d.uvPath,
		Packages:        []string{},
		CreationTime:    time.Time{},
	}

	// Check if .python-version file exists
	pythonVersion, err := d.readPythonVersion()
	if err != nil {
		return nil, fmt.Errorf("failed to read .python-version: %w", err)
	}
	if pythonVersion != "" {
		env.PythonVersion = pythonVersion
	}

	// Check if uv.lock file exists
	lockFile := filepath.Join(d.uvPath, "uv.lock")
	if _, err := os.Stat(lockFile); os.IsNotExist(err) {
		// No uv.lock file - not a valid UV environment
		return nil, nil
	}

	// Parse uv.lock file
	if err := d.parseUVLock(lockFile); err != nil {
		return nil, fmt.Errorf("failed to parse uv.lock: %w", err)
	}

	// Detect creation time from uv.lock metadata
	creationTime := d.detectCreationTime()

	env.Packages = d.packagesToSlice()
	env.CreationTime = creationTime

	return env, nil
}

// readPythonVersion reads the .python-version file
func (d *UVDetector) readPythonVersion() (string, error) {
	versionFile := filepath.Join(d.uvPath, ".python-version")

	data, err := os.ReadFile(versionFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // File doesn't exist, that's okay
		}
		return "", err
	}

	content := strings.TrimSpace(string(data))

	// Skip if file is empty
	if content == "" {
		return "", nil
	}

	// Handle different formats:
	// - "3.9" or "3.9.13" (exact version)
	// - "3.9-*" or "3.9.13-*" (version range)
	// - "3.9 (py39, cli)" (with interpreter info)

	version := d.parsePythonVersionContent(content)
	return version, nil
}

// parsePythonVersionContent parses the .python-version file content
func (d *UVDetector) parsePythonVersionContent(content string) string {
	// Remove comments and whitespace
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Remove interpreter info in parentheses
		re := regexp.MustCompile`\([^)]*\)`
		line = re.ReplaceAllString(line, "")

		// Trim again after removing parentheses
		line = strings.TrimSpace(line)

		// Handle version ranges like "3.9-*"
		if strings.Contains(line, "-*") {
			// Extract the base version
			line = strings.SplitN(line, "-*", 2)[0]
		}

		return line
	}

	return ""
}

// parseUVLock parses the uv.lock file
func (d *UVDetector) parseUVLock(lockFile string) error {
	file, err := os.Open(lockFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Lock file doesn't exist, that's okay
		}
		return err
	}
	defer file.Close()

	// Try to parse as JSON first
	if err := d.parseUVLockJSON(file); err == nil {
		return nil
	}

	// If JSON parsing fails, try text-based parsing (uv.lock can be text format)
	if err := d.parseUVLockText(file); err == nil {
		return nil
	}

	return fmt.Errorf("failed to parse uv.lock as JSON or text: %w", err)
}

// parseUVLockJSON parses uv.lock as JSON
func (d *UVDetector) parseUVLockJSON(file *os.File) error {
	// Read entire file
	data, err := os.ReadFile(file.Name())
	if err != nil {
		return err
	}

	var lock struct {
		Version     int                  `json:"version"`
		Metadata    map[string]string    `json:"metadata"`
		ResolverVersion string               `json:"resolver-version"`
		ProjectName string                `json:"project-name"`
		Dev         bool                 `json:"dev"`
		Hash        string               `json:"hash"`
		Workspace   []struct {
			Name        string   `json:"name"`
			Directory   string   `json:"directory"`
			Dev         bool     `json:"dev"`
			Locked      bool     `json:"locked"`
			Dependencies []struct {
				Name     string `json:"name"`
				Version  string `json:"version"`
				Dev      bool   `json:"dev"`
				Optional bool   `json:"optional"`
			} `json:"dependencies"`
		} `json:"workspace"`
		Installed []struct {
			Name      string            `json:"name"`
			Version   string            `json:"version"`
			Source    string            `json:"source"`
			Directory string            `json:"directory"`
			Hash      string            `json:"hash"`
			Dependencies []struct {
				Name     string `json:"name"`
				Version  string `json:"version,omitempty"`
				Inner    string `json:"inner,omitempty"`
				Optional bool   `json:"optional,omitempty"`
			} `json:"dependencies"`
		} `json:"installed"`
	}

	if err := json.Unmarshal(data, &lock); err != nil {
		return err
	}

	// Extract Python version from metadata
	if meta, ok := lock.Metadata["content-hash"]; ok {
		// Check if there's Python version in metadata
		if pythonVer, ok := lock.Metadata["python-version"]; ok {
			d.pythonVersions = append(d.pythonVersions, pythonVer)
		}
	}

	// Process workspace packages
	for _, ws := range lock.Workspace {
		for _, dep := range ws.Dependencies {
			d.packages[dep.Name] = dep.Version
		}
	}

	// Process installed packages
	for _, pkg := range lock.Installed {
		d.packages[pkg.Name] = pkg.Version
	}

	return nil
}

// parseUVLockText parses uv.lock in text format
func (d *UVDetector) parseUVLockText(file *os.File) error {
	scanner := bufio.NewScanner(file)
	inMetadata := false
	inWorkspace := false
	inInstalled := false

	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Detect sections
		if strings.HasPrefix(line, "[metadata]") {
			inMetadata = true
			inWorkspace = false
			inInstalled = false
			continue
		}
		if strings.HasPrefix(line, "[workspace]") {
			inWorkspace = true
			inMetadata = false
			inInstalled = false
			continue
		}
		if strings.HasPrefix(line, "[package]") {
			inInstalled = true
			inMetadata = false
			inWorkspace = false
			continue
		}
		if strings.HasPrefix(line, "[dev-dependencies]") || strings.HasPrefix(line, "[dev-packages]") {
			// dev-dependencies section
			inWorkspace = true
			inMetadata = false
			inInstalled = false
			continue
		}

		if inMetadata {
			// Parse metadata key = value
			if strings.HasPrefix(line, "python-version") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					version := strings.TrimSpace(parts[1])
					d.pythonVersions = append(d.pythonVersions, version)
				}
			}
		} else if inWorkspace {
			// Parse workspace entries
			if strings.HasPrefix(line, "path =") || strings.HasPrefix(line, "directory =") {
				// Workspace entry
				name := d.extractWorkspaceName(line)
				if name != "" {
					// Continue reading dependencies for this workspace entry
					for scanner.Scan() {
						line2 := scanner.Text()
						if strings.HasPrefix(line2, "[package]") || strings.HasPrefix(line2, "[dev-dependencies]") ||
							strings.HasPrefix(line2, "[dev-packages]") || strings.HasPrefix(line2, "[metadata]") ||
							strings.HasPrefix(line2, "path =") || strings.HasPrefix(line2, "directory =") {
							break
						}
						if strings.HasPrefix(line2, "name =") {
							parts := strings.SplitN(line2, "=", 2)
							if len(parts) == 2 {
								depName := strings.TrimSpace(parts[1])
								d.packages[depName] = ""
							}
						}
					}
				}
			}
		} else if inInstalled {
			// Parse installed package
			if strings.HasPrefix(line, "name =") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					name := strings.TrimSpace(parts[1])
					d.packages[name] = ""

					// Look for version in next lines
					for scanner.Scan() {
						line2 := scanner.Text()
						if strings.HasPrefix(line2, "version =") {
							parts2 := strings.SplitN(line2, "=", 2)
							if len(parts2) == 2 {
								d.packages[name] = strings.TrimSpace(parts2[1])
							}
							break
						}
					}
				}
			}
		}
	}

	return scanner.Err()
}

// extractWorkspaceName extracts the workspace package name from a line
func (d *UVDetector) extractWorkspaceName(line string) string {
	// Try "name = " first
	if strings.HasPrefix(line, "name =") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			return strings.TrimSpace(parts[1])
		}
	}

	// Try "directory = " as fallback
	if strings.HasPrefix(line, "directory =") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			name := strings.TrimSpace(parts[1])
			// Remove path prefix if present
			if strings.HasPrefix(name, "./") {
				name = name[2:]
			}
			return name
		}
	}

	return ""
}

// detectCreationTime detects the creation time of the UV environment
func (d *UVDetector) detectCreationTime() time.Time {
	// Use the timestamp from uv.lock metadata if available
	if len(d.pythonVersions) > 0 {
		// Use current time as fallback since uv.lock doesn't have a timestamp
		// The actual timestamp would be in the full lock metadata
		return time.Now()
	}

	return time.Now()
}

// packagesToSlice converts the packages map to a slice
func (d *UVDetector) packagesToSlice() []string {
	result := make([]string, 0, len(d.packages))
	for name := range d.packages {
		result = append(result, name)
	}
	return result
}

// GetPackages returns all detected packages
func (d *UVDetector) GetPackages() []types.Dependency {
	// Convert packages map to Dependency structs
	// For now, return empty and let the caller handle conversion
	return nil
}

// GetPythonVersions returns all detected Python versions
func (d *UVDetector) GetPythonVersions() []string {
	return d.pythonVersions
}
