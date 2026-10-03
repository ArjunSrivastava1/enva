package validator

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"enva/internal/detectors"
	"enva/internal/scanner"
	"enva/internal/types"
	"enva/pkg/api/osv"
	"enva/pkg/api/pypi"
	"enva/pkg/venv"
)

// ValidateEnvironment validates a Python virtual environment, conda environment, or UV environment
func ValidateEnvironment(venvPath string) (*types.ValidationResult, error) {
	ctx := context.Background()
	start := time.Now()

	result := &types.ValidationResult{
		OverallStatus: "success",
		Score:         100,
		Issues:        []types.Issue{},
		Suggestions:   []types.Suggestion{},
	}

	// Detect environment type (venv, conda, or uv)
	envType, envPath, err := detectEnvironmentType(venvPath)
	if err != nil {
		result.Issues = append(result.Issues, types.Issue{
			Type:      "venv",
			Severity:  "error",
			Message:   fmt.Sprintf("Failed to detect environment: %v", err),
			Component: "environment",
		})
		result.OverallStatus = "error"
		result.Score = 0
		return result, nil
	}

	// 1. Validate environment structure
	venvInfo, err := validateEnvironmentStructure(envPath, envType)
	if err != nil {
		result.Issues = append(result.Issues, types.Issue{
			Type:      "venv",
			Severity:  "error",
			Message:   fmt.Sprintf("Invalid environment: %v", err),
			Component: "environment",
		})
		result.OverallStatus = "error"
		result.Score = 0
		return result, nil
	}
	result.VenvInfo = venvInfo

	// 2. Get actual installed packages based on environment type
	installedPackages, err := getInstalledPackages(envPath, envType)
	if err != nil {
		result.Issues = append(result.Issues, types.Issue{
			Type:      "dependency",
			Severity:  "error",
			Message:   fmt.Sprintf("Failed to get installed packages: %v", err),
			Component: "dependencies",
		})
		result.Dependencies = []types.Dependency{}
	} else {
		result.Dependencies = convertPackagesToDependencies(installedPackages, ctx)
	}

	// 3. Check for requirements.txt consistency (only for venv/pip)
	if envType == "venv" {
		checkRequirementsConsistency(envPath, result, ctx)
	}

	// 4. Security scan using OSV API
	scan, err := scanSecurity(result.Dependencies, ctx)
	if err != nil {
		result.Issues = append(result.Issues, types.Issue{
			Type:      "security",
			Severity:  "error",
			Message:   fmt.Sprintf("Failed to scan for security vulnerabilities: %v", err),
			Component: "security",
		})
		result.OverallStatus = "error"
		result.Score = 0
	} else {
		result.Security = scan
	}

	// 5. Performance analysis
	result.Performance = analyzePerformance(envPath, result.Dependencies, envType)

	// 6. Generate suggestions
	result.Suggestions = generateSuggestions(result, envType)

	// 7. Calculate final score and status
	result.Duration = time.Since(start)
	calculateScore(result)

	return result, nil
}

// calculateScore calculates the final score and overall status
func calculateScore(result *types.ValidationResult) {
	score := 100

	// Deduct points based on venv status
	if result.VenvInfo != nil {
		switch result.VenvInfo.Status {
		case "error":
			score -= 30
		case "warning":
			score -= 15
		}

		// Deduct for not activated
		if result.VenvInfo.Activated == "not_activated" {
			score -= 5
		}
	}

	// Deduct for outdated packages
	outdatedCount := 0
	for _, dep := range result.Dependencies {
		if dep.Status == "outdated" {
			outdatedCount++
		}
	}
	score -= outdatedCount * 3 // 3 points per outdated package

	// Deduct for security issues
	if result.Security != nil {
		score -= result.Security.Critical * 25
		score -= result.Security.High * 15
		score -= result.Security.Medium * 5
		score -= result.Security.Low * 2
	}

	// Deduct for performance warnings
	if result.Performance != nil && result.Performance.Status == "warning" {
		score -= 10
	}

	// Deduct for other issues
	for _, issue := range result.Issues {
		switch issue.Severity {
		case "error":
			score -= 20
		case "warning":
			score -= 10
		case "info":
			score -= 5
		}
	}

	// Ensure score is within bounds
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	result.Score = score

	// Determine overall status
	if score >= 80 {
		result.OverallStatus = "success"
	} else if score >= 50 {
		result.OverallStatus = "warning"
	} else {
		result.OverallStatus = "error"
	}
}

// detectEnvironmentType detects the type of environment (venv, conda, or uv)
func detectEnvironmentType(path string) (string, string, error) {
	// Check for UV environment first (uv.lock)
	uvLockPath := filepath.Join(path, "uv.lock")
	if _, err := os.Stat(uvLockPath); err == nil {
		return "uv", path, nil
	}

	// Check for conda environment (conda-meta directory)
	condaMetaPath := filepath.Join(path, "conda-meta")
	if _, err := os.Stat(condaMetaPath); err == nil {
		return "conda", path, nil
	}

	// Check for virtual environment
	if venv.IsValid(path) {
		return "venv", path, nil
	}

	return "", "", fmt.Errorf("no valid environment detected")
}

// validateEnvironmentStructure validates the environment structure based on type
func validateEnvironmentStructure(envPath string, envType string) (*types.VenvInfo, error) {
	info := &types.VenvInfo{
		Path:      envPath,
		Status:    "success",
		Integrity: "valid",
	}

	switch envType {
	case "venv":
		return validateVenvStructure(envPath)
	case "conda":
		return validateCondaStructure(envPath)
	case "uv":
		return validateUVStructure(envPath)
	default:
		return nil, fmt.Errorf("unsupported environment type: %s", envType)
	}
}

// validateVenvStructure validates the virtual environment
func validateVenvStructure(venvPath string) (*types.VenvInfo, error) {
	info := &types.VenvInfo{
		Path:      venvPath,
		Status:    "success",
		Integrity: "valid",
	}

	// Check if venv is valid
	if !venv.IsValid(venvPath) {
		info.Status = "error"
		info.Integrity = "invalid"
		return info, fmt.Errorf("invalid virtual environment at %s", venvPath)
	}

	info.Integrity = "valid"

	// Get Python version - try executable first, then fall back to pyvenv.cfg
	pyVersion, err := venv.GetPythonVersion(venvPath)
	if err != nil {
		// Try to extract python.version from pyvenv.cfg
		cfgPath := filepath.Join(venvPath, "pyvenv.cfg")
		if pythonVer, cfgErr := getPythonVersionFromCfg(cfgPath); cfgErr == nil && pythonVer != "" {
			pyVersion = pythonVer
		} else {
			info.Status = "warning"
			info.PythonVersion = fmt.Sprintf("error: %v", err)
		}
	} else {
		info.PythonVersion = pyVersion
	}

	// Get pip version
	pipVersion, err := venv.GetPipVersion(venvPath)
	if err != nil {
		if info.Status != "error" {
			info.Status = "warning"
		}
		info.PipVersion = fmt.Sprintf("error: %v", err)
	} else {
		info.PipVersion = pipVersion
	}

	// Check activation - try both VIRTUAL_ENV env var and config
	activated := venv.IsActivated(venvPath)
	if !activated {
		// Also check pyvenv.cfg for activation indicator
		if venv.IsActivatedFromConfig(venvPath) {
			activated = true
		}
	}

	if activated {
		info.Activated = "activated"
	} else {
		info.Activated = "not_activated"
		if info.Status != "error" {
			info.Status = "warning"
		}
	}

	return info, nil
}

// validateCondaStructure validates conda environment structure
func validateCondaStructure(condaPath string) (*types.VenvInfo, error) {
	info := &types.VenvInfo{
		Path:      condaPath,
		Status:    "success",
		Integrity: "valid",
	}

	// Check if conda-meta directory exists
	condaMetaPath := filepath.Join(condaPath, "conda-meta")
	if _, err := os.Stat(condaMetaPath); os.IsNotExist(err) {
		info.Status = "error"
		info.Integrity = "invalid"
		return info, fmt.Errorf("conda-meta directory not found")
	}

	// Detect Python version from conda packages
	pythonVersion := detectPythonVersionFromConda(condaMetaPath)
	if pythonVersion != "" {
		info.PythonVersion = pythonVersion
	}

	// For conda, we don't track activation status the same way
	info.Activated = "activated"

	return info, nil
}

// validateUVStructure validates UV environment structure
func validateUVStructure(uvPath string) (*types.VenvInfo, error) {
	info := &types.VenvInfo{
		Path:      uvPath,
		Status:    "success",
		Integrity: "valid",
	}

	// Check if uv.lock exists
	uvLockPath := filepath.Join(uvPath, "uv.lock")
	if _, err := os.Stat(uvLockPath); os.IsNotExist(err) {
		info.Status = "error"
		info.Integrity = "invalid"
		return info, fmt.Errorf("uv.lock file not found")
	}

	// Detect Python version from .python-version
	pythonVersionFile := filepath.Join(uvPath, ".python-version")
	pythonVersion, err := readPythonVersionFromFile(pythonVersionFile)
	if err != nil {
		if os.IsNotExist(err) {
			info.PythonVersion = "unknown"
		} else {
			info.PythonVersion = fmt.Sprintf("error: %v", err)
		}
	} else if pythonVersion != "" {
		info.PythonVersion = pythonVersion
	}

	// For UV, we don't track activation status the same way
	info.Activated = "activated"

	return info, nil
}

// detectPythonVersionFromConda detects Python version from conda package metadata
func detectPythonVersionFromConda(condaMetaPath string) string {
	// Check conda-meta/python package
	pythonMeta := filepath.Join(condaMetaPath, "python")
	if info, err := os.Stat(pythonMeta); err == nil && !info.IsDir() {
		if metadata, err := parseCondaPackageMetadata(pythonMeta); err == nil {
			if version, ok := metadata["version"].(string); ok {
				return version
			}
		}
	}

	// Check conda-meta/history for python package
	historyFile := filepath.Join(condaMetaPath, "history")
	if content, err := os.ReadFile(historyFile); err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			// Parse: <timestamp> <package>[=<version>]... <old_version> <desc>
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				pkgSpec := parts[1]
				// Extract package name and version
				if strings.Contains(pkgSpec, "==") {
					pkgName := strings.SplitN(pkgSpec, "==", 2)[0]
					if pkgName == "python" {
						version := strings.SplitN(pkgSpec, "==", 2)[1]
						return version
					}
				}
			}
		}
	}

	return ""
}

// parseCondaPackageMetadata parses a conda package metadata file
func parseCondaPackageMetadata(pkgPath string) (map[string]interface{}, error) {
	file, err := os.Open(pkgPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var metadata map[string]interface{}
	if err := json.NewDecoder(file).Decode(&metadata); err != nil {
		return nil, err
	}

	return metadata, nil
}

// readPythonVersionFromFile reads .python-version file
func readPythonVersionFromFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	content := strings.TrimSpace(string(data))
	if content == "" {
		return "", nil
	}

	// Remove comments
	lines := strings.Split(content, "\n")
	var version string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Remove interpreter info in parentheses
		re := regexp.MustCompile(`\([^)]*\)`)
		line = re.ReplaceAllString(line, "")
		line = strings.TrimSpace(line)
		// Handle version ranges like "3.9-*"
		if strings.Contains(line, "-*") {
			line = strings.SplitN(line, "-*", 2)[0]
		}
		version = line
	}

	return version, nil
}

// getInstalledPackages gets installed packages from the environment based on type
func getInstalledPackages(envPath string, envType string) (map[string]string, error) {
	switch envType {
	case "venv":
		return venv.GetInstalledPackages(envPath)
	case "conda":
		return getInstalledPackagesFromConda(envPath)
	case "uv":
		return getInstalledPackagesFromUV(envPath)
	default:
		return nil, fmt.Errorf("unsupported environment type: %s", envType)
	}
}

// getInstalledPackagesFromConda gets installed packages from conda environment
func getInstalledPackagesFromConda(condaPath string) (map[string]string, error) {
	condaMetaPath := filepath.Join(condaPath, "conda-meta")
	packages := make(map[string]string)

	// Get list of package files in conda-meta
	packageFiles, err := os.ReadDir(condaMetaPath)
	if err != nil {
		return nil, err
	}

	// Filter for package files (not repodata.json or other metadata)
	for _, file := range packageFiles {
		if !file.IsDir() && !strings.HasPrefix(file.Name(), "repodata-") {
			if metadata, err := parseCondaPackageMetadata(filepath.Join(condaMetaPath, file.Name())); err == nil {
				if name, ok := metadata["name"].(string); ok {
					packages[name] = ""
					if version, ok := metadata["version"].(string); ok {
						packages[name] = version
					}
				}
			}
		}
	}

	return packages, nil
}

// getInstalledPackagesFromUV gets installed packages from UV environment
func getInstalledPackagesFromUV(uvPath string) (map[string]string, error) {
	// Detect using the UV detector
	detector := detectors.NewUVDetector(uvPath)
	env, err := detector.Detect()
	if err != nil {
		return nil, err
	}

	if env == nil {
		return nil, fmt.Errorf("UV environment not found")
	}

	return map[string]string{
		"packages": env.Packages,
	}, nil
}

// convertPackagesToDependencies converts package map to Dependency structs
func convertPackagesToDependencies(packages map[string]string, ctx context.Context) []types.Dependency {
	// Create a shared OSV client for vulnerability scanning
	osvClient := osv.NewClient(ctx, "", osv.Default().CacheTTL, osv.Default().RateLimit, osv.Default().MaxCacheSize)

	// Create a shared PyPI client for latest version lookup
	pypiClient := pypi.NewClient(ctx, "", pypi.Default().CacheTTL, pypi.Default().RateLimit, pypi.Default().MaxCacheSize)

	var deps []types.Dependency
	depsMu := &sync.Mutex{}

	for name, version := range packages {
		dep := types.Dependency{
			Name:    name,
			Version: version,
			Status:  "uptodate", // Default
		}

		// Skip packages without version or with special versions
		if version == "" || version == "editable" || version == "direct_url" {
			continue
		}

		// Extract version number for comparison
		versionNum := extractVersion(version)
		if versionNum == "" {
			continue
		}

		// Get latest version from PyPI API
		latest, err := pypiClient.GetLatestVersion(name)
		if err != nil {
			// If API fails, set Latest to empty and don't mark as outdated
			dep.Latest = ""
		} else {
			dep.Latest = latest
			if latest != "" {
				latestNum := extractVersion(latest)
				if latestNum != "" && versionNum < latestNum {
					dep.Status = "outdated"
				}
			}
		}

		// Check for vulnerabilities using OSV API
		vulns, err := osvClient.GetVulnerabilitiesForPackage(ctx, name, "pypi")
		if err != nil {
			// Log error but continue
			// fmt.Printf("Warning: Failed to scan %s for vulnerabilities: %v\n", name, err)
		} else {
			// Filter vulnerabilities by installed version
			filteredVulns := filterVulnerabilitiesByVersion(vulns, []types.Dependency{dep})

			if len(filteredVulns) > 0 {
				// Find the most severe vulnerability
				maxSeverity := getSeverityScore(dep.Severity)
				for _, vuln := range filteredVulns {
					vulnSeverity := getSeverityScore(vuln.Severity)
					if vulnSeverity > maxSeverity {
						maxSeverity = vulnSeverity
					}
				}

				// Map severity to status
				switch maxSeverity {
				case 9:
					dep.Status = "vulnerable_critical"
				case 7:
					dep.Status = "vulnerable_high"
				case 4:
					dep.Status = "vulnerable_medium"
				case 0:
					dep.Status = "vulnerable_low"
				}
			}
		}

		deps = append(deps, dep)
	}

	pypiClient.Close()
	osvClient.Close()
	return deps
}

// extractVersion extracts a comparable version string from a version specifier
func extractVersion(version string) string {
	version = strings.TrimSpace(version)

	// Remove version operators
	version = strings.TrimPrefix(version, ">=")
	version = strings.TrimPrefix(version, "<=")
	version = strings.TrimPrefix(version, ">")
	version = strings.TrimPrefix(version, "<")
	version = strings.TrimPrefix(version, "~=")
	version = strings.TrimPrefix(version, "!=")

	// Remove leading "==" if present
	version = strings.TrimPrefix(version, "==")

	return strings.TrimSpace(version)
}

// checkRequirementsConsistency checks requirements.txt against installed packages
func checkRequirementsConsistency(venvPath string, result *types.ValidationResult, ctx context.Context) {
	// Look for requirements.txt in project root
	reqFile := findRequirementsFile(venvPath)
	if reqFile == "" {
		return // No requirements.txt found, skip check
	}

	// Parse requirements.txt
	parsedReqs, err := scanner.ParseRequirementsFile(reqFile)
	if err != nil {
		result.Issues = append(result.Issues, types.Issue{
			Type:      "dependency",
			Severity:  "warning",
			Message:   fmt.Sprintf("Failed to parse requirements.txt: %v", err),
			Component: "requirements",
			Line:      0,
		})
		return
	}

	// Create a map of installed packages for quick lookup
	installedMap := make(map[string]string)
	for _, dep := range result.Dependencies {
		installedMap[dep.Name] = dep.Version
	}

	// Check for missing packages (in requirements but not installed)
	for reqName, reqVersion := range parsedReqs {
		if reqVersion == "" {
			// Just checking presence
			if _, ok := installedMap[reqName]; !ok {
				result.Issues = append(result.Issues, types.Issue{
					Type:      "dependency",
					Severity:  "warning",
					Message:   fmt.Sprintf("%s is required but not installed", reqName),
					Component: "requirements",
				})
			}
		} else {
			// Check version constraints
			installedVer := installedMap[reqName]
			if installedVer != "" && !versionSatisfies(installedVer, reqVersion) {
				result.Issues = append(result.Issues, types.Issue{
					Type:      "dependency",
					Severity:  "warning",
					Message:   fmt.Sprintf("%s version %s does not satisfy requirement %s", reqName, installedVer, reqVersion),
					Component: "requirements",
				})
			}
		}
	}

	// Check for extra packages (installed but not in requirements)
	for instName := range installedMap {
		if _, ok := parsedReqs[instName]; !ok {
			// This is an extra package - only warn, don't error
			result.Suggestions = append(result.Suggestions, types.Suggestion{
				Type:        "config",
				Description: fmt.Sprintf("%s is installed but not in requirements.txt", instName),
				Command:     "pip freeze > requirements.txt",
				AutoFixable: true,
				Priority:    "low",
			})
		}
	}
}

// findRequirementsFile finds requirements.txt near the venv path
func findRequirementsFile(venvPath string) string {
	dir := filepath.Dir(venvPath)

	// Look for requirements.txt in current and parent directories
	for dir != "/" {
		for _, marker := range []string{"requirements.txt", "requirements.in", "Pipfile"} {
			if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
				return filepath.Join(dir, marker)
			}
		}
		dir = filepath.Dir(dir)
	}

	return ""
}

// versionSatisfies checks if installed version satisfies a version specifier
func versionSatisfies(installed, specifier string) bool {
	installed = strings.TrimSpace(installed)

	// Handle == specifier
	if strings.HasPrefix(specifier, "==") {
		target := strings.TrimPrefix(specifier, "==")
		return installed == target
	}

	// Handle >= specifier
	if strings.HasPrefix(specifier, ">=") {
		target := strings.TrimPrefix(specifier, ">=")
		return versionCompare(installed, target) >= 0
	}

	// Handle <= specifier
	if strings.HasPrefix(specifier, "<=") {
		target := strings.TrimPrefix(specifier, "<=")
		return versionCompare(installed, target) <= 0
	}

	// Handle > specifier (but not >=)
	if strings.HasPrefix(specifier, ">") && !strings.Contains(specifier, ">=") {
		target := strings.TrimPrefix(specifier, ">")
		return versionCompare(installed, target) > 0
	}

	// Handle < specifier (but not <=)
	if strings.HasPrefix(specifier, "<") && !strings.Contains(specifier, "<=") {
		target := strings.TrimPrefix(specifier, "<")
		return versionCompare(installed, target) < 0
	}

	// Handle ~= specifier (compatible release)
	if strings.HasPrefix(specifier, "~=") {
		target := strings.TrimPrefix(specifier, "~=")
		// ~= 1.4 means >= 1.4, == 1.*
		if versionCompare(installed, target) >= 0 {
			installedMajorMinor := fmt.Sprintf("%s.*", target)
			if versionMatchWildcard(installed, installedMajorMinor) {
				return true
			}
		}
		return false
	}

	// Handle != specifier
	if strings.HasPrefix(specifier, "!=") {
		target := strings.TrimPrefix(specifier, "!=")
		return installed != target
	}

	// No specifier - just check if installed
	return installed != ""
}

// versionCompare compares two version strings
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func versionCompare(v1, v2 string) int {
	parts1 := parseVersionParts(v1)
	parts2 := parseVersionParts(v2)

	// Pad with zeros to match lengths
	for len(parts1) < len(parts2) {
		parts1 = append(parts1, 0)
	}
	for len(parts2) < len(parts1) {
		parts2 = append(parts2, 0)
	}

	for i := 0; i < len(parts1); i++ {
		p1 := parts1[i]
		p2 := parts2[i]

		// Compare as integers if both are numeric
		if p1 != "" && p2 != "" && isNumeric(p1) && isNumeric(p2) {
			if p1 > p2 {
				return 1
			} else if p1 < p2 {
				return -1
			}
		} else if p1 != p2 {
			// String comparison for non-numeric parts
			if p1 > p2 {
				return 1
			} else if p1 < p2 {
				return -1
			}
		}
	}

	return 0
}

// parseVersionParts parses a version string into numeric parts
func parseVersionParts(version string) []int {
	parts := []int{}
	fields := strings.Split(version, ".")

	for _, field := range fields {
		field = strings.TrimSpace(field)
		if num, err := strconv.Atoi(field); err == nil {
			parts = append(parts, num)
		} else {
			parts = append(parts, 0)
		}
	}

	return parts
}

// isNumeric checks if a string is a numeric value
func isNumeric(s string) bool {
	if _, err := strconv.Atoi(s); err == nil {
		return true
	}
	return false
}

// versionMatchWildcard checks if version matches X.Y.* pattern
func versionMatchWildcard(version, pattern string) bool {
	parts := strings.Split(pattern, ".")
	versionParts := strings.Split(version, ".")

	if len(parts) != 2 {
		return false
	}

	if len(versionParts) < 2 {
		return false
	}

	if parts[0] == versionParts[0] && parts[1] == "*" {
		return true
	}

	if parts[0] == versionParts[0] && parts[1] != "" {
		return true
	}

	return false
}

// scanSecurity checks for security vulnerabilities using OSV API
func scanSecurity(deps []types.Dependency, ctx context.Context) (*types.SecurityScan, error) {
	// Create OSV client
	osvClient := osv.NewClient(ctx, "", osv.Default().CacheTTL, osv.Default().RateLimit, osv.Default().MaxCacheSize)

	scan := &types.SecurityScan{
		Status: "success",
	}

	// Scan all packages for vulnerabilities concurrently
	vulns, err := osvClient.ScanAllPackages(ctx, deps)
	if err != nil {
		// Log error but continue with partial results
		// fmt.Printf("Warning: Failed to scan all packages: %v\n", err)
		// Return empty scan
		return scan, nil
	}

	// Update scan with vulnerabilities
	for _, vuln := range vulns {
		// Map OSV severity to enva severity
		severity := getSeverityScore(vuln.Severity)

		switch severity {
		case 9:
			scan.Critical++
		case 7:
			scan.High++
		case 4:
			scan.Medium++
		case 0:
			scan.Low++
		}

		vulnObj := types.Vulnerability{
			ID:          vuln.ID,
			Package:     vuln.Package,
			Version:     vuln.Version,
			Severity:    getSeverityText(severity),
			Description: vuln.Description,
			FixedIn:     vuln.FixedIn,
			CVEs:        vuln.CVEs,
		}

		scan.Vulnerabilities = append(scan.Vulnerabilities, vulnObj)
	}

	// Determine overall status based on severity
	if len(scan.Vulnerabilities) > 0 {
		scan.Status = "warning"
		if scan.Critical > 0 || scan.High > 0 {
			scan.Status = "error"
		}
	}

	osvClient.Close()
	return scan, nil
}

// analyzePerformance checks performance issues based on environment type
func analyzePerformance(envPath string, deps []types.Dependency, envType string) *types.Performance {
	perf := &types.Performance{
		Status: "success",
	}

	// Check for large packages
	largePackages := []string{"tensorflow", "pytorch", "opencv-python"}
	for _, pkg := range largePackages {
		for _, dep := range deps {
			if dep.Name == pkg {
				perf.LargePackages = append(perf.LargePackages, types.PackageSize{
					Name: pkg,
					Size: "100MB+",
				})
			}
		}
	}

	if len(perf.LargePackages) > 0 {
		perf.Status = "warning"
		perf.Optimizations = append(perf.Optimizations, types.Optimization{
			Type:        "size",
			Description: "Large packages may slow down environment",
			Impact:      "medium",
		})
	}

	// Check for many packages
	if len(deps) > 20 {
		perf.Status = "warning"
		perf.Optimizations = append(perf.Optimizations, types.Optimization{
			Type:        "quantity",
			Description: fmt.Sprintf("Many packages (%d), consider streamlining", len(deps)),
			Impact:      "low",
		})
	}

	return perf
}

// generateSuggestions creates fix suggestions based on environment type
func generateSuggestions(result *types.ValidationResult, envType string) []types.Suggestion {
	var suggestions []types.Suggestion

	// Activation suggestion for venv
	if envType == "venv" && result.VenvInfo != nil && result.VenvInfo.Activated == "not_activated" {
		suggestions = append(suggestions, types.Suggestion{
			Type:        "config",
			Description: "Activate virtual environment for development",
			Command:     fmt.Sprintf("source %s/bin/activate", result.VenvInfo.Path),
			AutoFixable: false,
			Priority:    "medium",
		})
	}

	// Update outdated packages
	for _, dep := range result.Dependencies {
		if dep.Status == "outdated" && dep.Latest != "" {
			var command string
			switch envType {
			case "venv":
				command = fmt.Sprintf("pip install %s==%s", dep.Name, dep.Latest)
			case "conda":
				command = fmt.Sprintf("conda install %s==%s", dep.Name, dep.Latest)
			case "uv":
				command = fmt.Sprintf("uv pip install %s==%s", dep.Name, dep.Latest)
			}
			suggestions = append(suggestions, types.Suggestion{
				Type:        "update",
				Description: fmt.Sprintf("Update %s to version %s", dep.Name, dep.Latest),
				Command:     command,
				AutoFixable: true,
				Priority:    "high",
			})
		}
	}

	// Security fixes
	if result.Security != nil {
		for _, vuln := range result.Security.Vulnerabilities {
			var command string
			switch envType {
			case "venv":
				command = fmt.Sprintf("pip install %s==%s", vuln.Package, vuln.FixedIn)
			case "conda":
				command = fmt.Sprintf("conda install %s==%s", vuln.Package, vuln.FixedIn)
			case "uv":
				command = fmt.Sprintf("uv pip install %s==%s", vuln.Package, vuln.FixedIn)
			}
			suggestions = append(suggestions, types.Suggestion{
				Type:        "security",
				Description: fmt.Sprintf("Fix vulnerability %s in %s", vuln.ID, vuln.Package),
				Command:     command,
				AutoFixable: vuln.FixedIn != "unknown",
				Priority:    "high",
			})
		}
	}

	return suggestions
}

// getSeverityScore maps severity string to numeric score
func getSeverityScore(severity string) int {
	switch strings.ToLower(severity) {
	case "critical":
		return 9
	case "high":
		return 7
	case "medium":
		return 4
	case "low":
		return 0
	default:
		return 0
	}
}

// getSeverityText maps numeric score to severity string
func getSeverityText(score int) string {
	switch score {
	case 9:
		return "critical"
	case 7:
		return "high"
	case 4:
		return "medium"
	case 0:
		return "low"
	default:
		return ""
	}
}
