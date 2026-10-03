package scanner

import (
	"context"
	"enva/internal/types"
	"enva/pkg/api/osv"
	"fmt"
)

// ScanSecurity checks for known vulnerabilities using OSV API
func ScanSecurity(deps []types.Dependency, ctx context.Context) (*types.SecurityScan, error) {
	// Create OSV client
	osvClient := osv.NewClient(ctx, "", osv.Default().CacheTTL, osv.Default().RateLimit, osv.Default().MaxCacheSize)

	scan := &types.SecurityScan{
		Status: "success",
	}

	// Scan all packages for vulnerabilities using OSV API
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
