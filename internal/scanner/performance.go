package scanner

import (
	"fmt"

	"enva/internal/types"
)

// AnalyzePerformance checks for performance issues
func AnalyzePerformance(venvPath string, deps []types.Dependency) (*types.Performance, error) {
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

	return perf, nil
}
