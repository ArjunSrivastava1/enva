package osv

import (
	"fmt"
	"strings"
)

// MapVulnerabilitiesToCVEs maps OSV vulnerabilities to CVE IDs
func MapVulnerabilitiesToCVEs(vulns []*Vulnerability) map[string][]string {
	cveMap := make(map[string][]string)

	for _, vuln := range vulns {
		// Check if the vulnerability has a CVE ID
		if vuln.Severity.CVE != "" {
			cveMap[vuln.ID] = append(cveMap[vuln.ID], vuln.Severity.CVE)
		}

		// Also check references for CVEs
		for _, ref := range vuln.References {
			if ref.Tags == nil {
				continue
			}
			for _, tag := range ref.Tags {
				if strings.Contains(tag, "CVE") {
					cveMap[vuln.ID] = append(cveMap[vuln.ID], fmt.Sprintf("CVE-%s", strings.TrimPrefix(tag, "CVE-")))
				}
			}
		}
	}

	return cveMap
}

// MapVulnerabilityToSeverity maps OSV severity to enva severity
func MapVulnerabilityToSeverity(vuln *Vulnerability) string {
	if vuln.Severity.Text != "" {
		return vuln.Severity.Text
	}
	return ""
}

// MapVulnerabilitySeverityScoreTextToEnvaseverity maps OSV severity score text to enva severity
func MapVulnerabilitySeverityScoreTextToEnvaseverity(scoreText string) string {
	switch scoreText {
	case "CRITICAL", "critical":
		return "critical"
	case "HIGH", "high":
		return "high"
	case "MEDIUM", "medium":
		return "medium"
	case "LOW", "low":
		return "low"
	default:
		return ""
	}
}

// MapVulnerabilitySeverityScoreToEnvaseverity maps OSV severity score value to enva severity
func MapVulnerabilitySeverityScoreToEnvaseverity(score float64) string {
	if score >= 9.0 {
		return "critical"
	} else if score >= 7.0 {
		return "high"
	} else if score >= 4.0 {
		return "medium"
	} else if score >= 0.0 {
		return "low"
	}
	return ""
}

// GetVulnerabilityFixedVersion extracts the fixed version from an affected range event
func GetVulnerabilityFixedVersion(affected *Affected) string {
	for _, rangeDef := range affected.Ranges {
		if rangeDef.Type == "version" {
			for _, event := range rangeDef.Events {
				if event.Fixed != "" {
					return event.Fixed
				}
			}
		}
	}
	return ""
}

// GetVulnerabilityVulnerableVersions extracts vulnerable versions from an affected range
func GetVulnerabilityVulnerableVersions(affected *Affected) []string {
	var vulnerableVersions []string

	for _, rangeDef := range affected.Ranges {
		if rangeDef.Type == "version" {
			for _, event := range rangeDef.Events {
				if event.Intro != "" && event.Last != "" {
					vulnerableVersions = append(vulnerableVersions, event.Intro)
				} else if event.Last != "" {
					vulnerableVersions = append(vulnerableVersions, event.Last)
				}
			}
		}
	}

	return vulnerableVersions
}

// GetVulnerabilityIntroductionVersion extracts the version where vulnerability was introduced
func GetVulnerabilityIntroductionVersion(affected *Affected) string {
	for _, rangeDef := range affected.Ranges {
		if rangeDef.Type == "version" {
			for _, event := range rangeDef.Events {
				if event.Intro != "" {
					return event.Intro
				}
			}
		}
	}
	return ""
}

// GetVulnerabilityLastVulnerableVersion extracts the last known vulnerable version
func GetVulnerabilityLastVulnerableVersion(affected *Affected) string {
	for _, rangeDef := range affected.Ranges {
		if rangeDef.Type == "version" {
			for _, event := range rangeDef.Events {
				if event.Last != "" {
					return event.Last
				}
			}
		}
	}
	return ""
}
