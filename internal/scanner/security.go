package scanner

import (
	"enva/internal/validator"
	"fmt"
)

// ScanSecurity checks for known vulnerabilities
func ScanSecurity(deps []validator.Dependency) (*validator.SecurityScan, error) {
	scan := &validator.SecurityScan{
		Status: "success",
	}

	// Known vulnerable packages (simplified)
	vulnerablePackages := map[string]string{
		"requests": "2.28.2", // CVE-2023-12345 example
		"urllib3":  "1.26.0", // CVE-2023-12346 example
	}

	for _, dep := range deps {
		if fixedVersion, ok := vulnerablePackages[dep.Name]; ok {
			if dep.Version == "=="+fixedVersion ||
				(dep.Version == "" && dep.Name == fixedVersion) {

				scan.Medium++ // Count as medium severity
				scan.Vulnerabilities = append(scan.Vulnerabilities, validator.Vulnerability{
					ID:          "CVE-2023-XXXXX",
					Package:     dep.Name,
					Version:     dep.Version,
					Severity:    "medium",
					Description: fmt.Sprintf("Security vulnerability in %s", dep.Name),
					FixedIn:     getLatestVersion(dep.Name),
				})

				// Update dependency status
				dep.Status = "vulnerable"
			}
		}
	}

	if len(scan.Vulnerabilities) > 0 {
		scan.Status = "warning"
		if scan.Critical > 0 || scan.High > 0 {
			scan.Status = "error"
		}
	}

	return scan, nil
}
