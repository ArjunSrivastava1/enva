package formatter

import (
	"fmt"
	"strings"
	"time"

	"enva/internal/types"
)

func FormatChinese(result *types.ValidationResult) string {
	var output strings.Builder

	// Header
	statusIcon := getStatusIcon(result.OverallStatus)
	timestamp := time.Now().Format("15:04:05")

	output.WriteString(fmt.Sprintf("[%s STATUS] Environment Validation\n", statusIcon))
	output.WriteString("───────────────────────────────────────────\n")
	output.WriteString(fmt.Sprintf("Time: %s | Duration: %.2fs\n\n", timestamp, result.Duration.Seconds()))

	// Virtual Environment
	if result.VenvInfo != nil {
		output.WriteString(formatVenvSection(result.VenvInfo))
		output.WriteString("\n")
	}

	// Dependencies
	if len(result.Dependencies) > 0 {
		output.WriteString(formatDependenciesSection(result.Dependencies))
		output.WriteString("\n")
	}

	// Security
	if result.Security != nil {
		output.WriteString(formatSecuritySection(result.Security))
		output.WriteString("\n")
	}

	// Summary
	output.WriteString(formatSummarySection(result))

	return output.String()
}

func formatVenvSection(info *types.VenvInfo) string {
	var section strings.Builder

	statusIcon := getStatusIcon(info.Status)
	section.WriteString(fmt.Sprintf("[%s VENV] Virtual Environment\n", statusIcon))

	items := []struct {
		label  string
		value  string
		status string
	}{
		{"Path", info.Path, "info"},
		{"Python", info.PythonVersion, "success"},
		{"Pip", info.PipVersion, "success"},
		{"Status", info.Activated, getStatusFromActivation(info.Activated)},
		{"Integrity", info.Integrity, info.Status},
	}

	for _, item := range items {
		icon := getStatusIcon(item.status)
		section.WriteString(fmt.Sprintf("  %s %s: %s\n", icon, item.label, item.value))
	}

	return section.String()
}

func formatDependenciesSection(deps []types.Dependency) string {
	var section strings.Builder

	// Determine section status
	sectionStatus := "success"
	for _, dep := range deps {
		if dep.Status == "vulnerable" {
			sectionStatus = "error"
			break
		}
		if dep.Status == "outdated" {
			sectionStatus = "warning"
		}
	}

	statusIcon := getStatusIcon(sectionStatus)
	section.WriteString(fmt.Sprintf("[%s PACKAGES] Dependencies (%d)\n", statusIcon, len(deps)))

	for _, dep := range deps {
		icon := getStatusIcon(dep.Status)

		line := fmt.Sprintf("  %s %s", icon, dep.Name)
		if dep.Version != "" {
			line += fmt.Sprintf(" %s", dep.Version)
		}

		if dep.Latest != "" && dep.Latest != dep.Version {
			line += fmt.Sprintf(" → latest: %s", dep.Latest)
		}

		section.WriteString(line + "\n")
	}

	return section.String()
}

func formatSecuritySection(security *types.SecurityScan) string {
	var section strings.Builder

	statusIcon := getStatusIcon(security.Status)
	section.WriteString(fmt.Sprintf("[%s SECURITY] Vulnerability Scan\n", statusIcon))

	if security.Critical > 0 || security.High > 0 {
		section.WriteString(fmt.Sprintf("  ❌ Critical: %d\n", security.Critical))
		section.WriteString(fmt.Sprintf("  ❌ High: %d\n", security.High))
	}
	if security.Medium > 0 {
		section.WriteString(fmt.Sprintf("  ⚠️  Medium: %d\n", security.Medium))
	}
	if security.Low > 0 {
		section.WriteString(fmt.Sprintf("  ℹ️  Low: %d\n", security.Low))
	}

	if len(security.Vulnerabilities) > 0 {
		section.WriteString("\n  [ISSUES] Found:\n")
		for _, vuln := range security.Vulnerabilities {
			section.WriteString(fmt.Sprintf("    • %s in %s\n", vuln.ID, vuln.Package))
			if vuln.FixedIn != "" {
				section.WriteString(fmt.Sprintf("      Fix: upgrade to %s\n", vuln.FixedIn))
			}
		}
	}

	return section.String()
}

func formatSummarySection(result *types.ValidationResult) string {
	var section strings.Builder

	statusIcon := getStatusIcon(result.OverallStatus)
	section.WriteString(fmt.Sprintf("[%s SUMMARY] Final Result\n", statusIcon))

	section.WriteString(fmt.Sprintf("  Score: %d/100\n", result.Score))
	section.WriteString(fmt.Sprintf("  Status: %s\n", strings.ToUpper(result.OverallStatus)))
	section.WriteString(fmt.Sprintf("  Time: %.2f seconds\n", result.Duration.Seconds()))

	// Final message
	section.WriteString("\n")
	switch result.OverallStatus {
	case "success":
		section.WriteString("[✅ READY] Environment is production-ready\n")
	case "warning":
		section.WriteString("[⚠️  REVIEW] Environment has issues to address\n")
	case "error":
		section.WriteString("[❌ BLOCKED] Critical issues found\n")
	}

	return section.String()
}

func getStatusIcon(status string) string {
	// Map package statuses to icons
	statusMap := map[string]string{
		// Package statuses
		"uptodate":   "✅",
		"outdated":   "⚠️",
		"vulnerable": "❌",
		"missing":    "❌",

		// Validation statuses
		"success": "✅",
		"warning": "⚠️",
		"error":   "❌",
		"info":    "ℹ️",
	}

	if icon, ok := statusMap[status]; ok {
		return icon
	}

	return "❓"
}

func getStatusFromActivation(activated string) string {
	if activated == "activated" {
		return "success"
	}
	return "warning"
}
