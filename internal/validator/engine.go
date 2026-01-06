package validator

import (
	"fmt"
	"time"

	"enva/internal/types"
	"enva/pkg/venv"
)

// ValidateEnvironment - main validation function
func ValidateEnvironment(venvPath string) (*types.ValidationResult, error) {
	start := time.Now()

	result := &types.ValidationResult{
		OverallStatus: "success",
		Score:         100,
		Issues:        []types.Issue{},
		Suggestions:   []types.Suggestion{},
	}

	// 1. Validate venv structure
	venvInfo, err := validateVenvStructure(venvPath)
	if err != nil {
		return nil, fmt.Errorf("venv validation failed: %v", err)
	}
	result.VenvInfo = venvInfo

	// 2. Get actual installed packages
	packages, err := venv.GetInstalledPackages(venvPath)
	if err != nil {
		result.Issues = append(result.Issues, types.Issue{
			Type:     "dependency",
			Severity: "warning",
			Message:  fmt.Sprintf("Failed to get packages: %v", err),
		})
		result.Dependencies = []types.Dependency{}
	} else {
		result.Dependencies = convertPackagesToDependencies(packages)
	}

	// 3. Check for requirements.txt consistency
	checkRequirementsConsistency(venvPath, result)

	// 4. Security scan
	result.Security = scanSecurity(result.Dependencies)

	// 5. Performance analysis
	result.Performance = analyzePerformance(venvPath, result.Dependencies)

	// 6. Generate suggestions
	result.Suggestions = generateSuggestions(result)

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

	// Get Python version
	pyVersion, err := venv.GetPythonVersion(venvPath)
	if err != nil {
		info.Status = "warning"
		info.PythonVersion = fmt.Sprintf("error: %v", err)
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

	// Check activation
	if venv.IsActivated(venvPath) {
		info.Activated = "activated"
	} else {
		info.Activated = "not_activated"
		if info.Status != "error" {
			info.Status = "warning"
		}
	}

	return info, nil
}

// convertPackagesToDependencies converts package map to Dependency structs
func convertPackagesToDependencies(packages map[string]string) []types.Dependency {
	var deps []types.Dependency

	for name, version := range packages {
		dep := types.Dependency{
			Name:    name,
			Version: version,
			Status:  "uptodate", // Default
		}

		// Simple outdated detection (in real app, query PyPI)
		if name == "requests" && version < "2.31.0" {
			dep.Status = "outdated"
			dep.Latest = "2.31.0"
		}

		deps = append(deps, dep)
	}

	return deps
}

// checkRequirementsConsistency checks requirements.txt
func checkRequirementsConsistency(venvPath string, result *types.ValidationResult) {
	// TODO: Implement requirements.txt checking
	// For now, just a placeholder
}

// scanSecurity checks for security vulnerabilities
func scanSecurity(deps []types.Dependency) *types.SecurityScan {
	scan := &types.SecurityScan{
		Status: "success",
	}

	// Simple security check
	for _, dep := range deps {
		if dep.Name == "requests" && dep.Version < "2.31.0" {
			scan.Medium++
			scan.Vulnerabilities = append(scan.Vulnerabilities, types.Vulnerability{
				ID:          "CVE-2023-XXXXX",
				Package:     dep.Name,
				Version:     dep.Version,
				Severity:    "medium",
				Description: "Example security vulnerability",
				FixedIn:     "2.31.0",
			})
		}
	}

	if len(scan.Vulnerabilities) > 0 {
		scan.Status = "warning"
	}

	return scan
}

// analyzePerformance checks performance issues
func analyzePerformance(venvPath string, deps []types.Dependency) *types.Performance {
	perf := &types.Performance{
		Status: "success",
	}

	// Check for large packages
	largePackages := []string{"tensorflow", "torch", "opencv-python"}
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

// generateSuggestions creates fix suggestions
func generateSuggestions(result *types.ValidationResult) []types.Suggestion {
	var suggestions []types.Suggestion

	// Activation suggestion
	if result.VenvInfo != nil && result.VenvInfo.Activated == "not_activated" {
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
			suggestions = append(suggestions, types.Suggestion{
				Type:        "update",
				Description: fmt.Sprintf("Update %s to version %s", dep.Name, dep.Latest),
				Command:     fmt.Sprintf("pip install %s==%s", dep.Name, dep.Latest),
				AutoFixable: true,
				Priority:    "high",
			})
		}
	}

	// Security fixes
	if result.Security != nil {
		for _, vuln := range result.Security.Vulnerabilities {
			suggestions = append(suggestions, types.Suggestion{
				Type:        "security",
				Description: fmt.Sprintf("Fix vulnerability %s in %s", vuln.ID, vuln.Package),
				Command:     fmt.Sprintf("pip install %s==%s", vuln.Package, vuln.FixedIn),
				AutoFixable: true,
				Priority:    "high",
			})
		}
	}

	return suggestions
}
