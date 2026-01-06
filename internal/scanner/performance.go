package scanner

import (
	"fmt"

	"enva/internal/validator"
)

// AnalyzePerformance checks for performance issues
func AnalyzePerformance(venvPath string, deps []validator.Dependency) (*validator.Performance, error) {
	perf := &validator.Performance{
		Status: "success",
	}

	// Check for large packages (simplified)
	largePackages := []string{"tensorflow", "pytorch", "opencv-python"}
	for _, pkg := range largePackages {
		for _, dep := range deps {
			if dep.Name == pkg {
				perf.LargePackages = append(perf.LargePackages, validator.PackageSize{
					Name: pkg,
					Size: "450MB+", // Example size
				})
			}
		}
	}

	// Check for unused packages (simplified logic)
	// In real implementation, would analyze imports vs installed
	if len(deps) > 20 {
		perf.UnusedPackages = []string{"example-unused-package"}
		perf.Optimizations = append(perf.Optimizations, validator.Optimization{
			Type:        "cleanup",
			Description: "Remove unused packages to reduce environment size",
			Impact:      "medium",
		})
	}

	// Check for outdated packages affecting performance
	outdatedCount := 0
	for _, dep := range deps {
		if dep.Status == "outdated" {
			outdatedCount++
		}
	}

	if outdatedCount > 5 {
		perf.Status = "warning"
		perf.Optimizations = append(perf.Optimizations, validator.Optimization{
			Type:        "update",
			Description: fmt.Sprintf("Update %d outdated packages for performance improvements", outdatedCount),
			Impact:      "high",
		})
	}

	return perf, nil
}
