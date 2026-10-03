package formatter

import (
	"enva/internal/types"
	"fmt"
	"strings"
	"time"
)

// FormatHTML generates a styled HTML report from the validation result
func FormatHTML(result *types.ValidationResult) string {
	if result == nil {
		return "<!DOCTYPE html><html><body><p>No validation result provided.</p></body></html>"
	}

	// Build the HTML report
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Enva Security Report</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif;
            line-height: 1.6;
            color: #333;
            background: #f5f5f5;
            padding: 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        .header {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            padding: 30px;
            border-radius: 10px;
            margin-bottom: 20px;
        }
        .header h1 {
            font-size: 2.5em;
            margin-bottom: 10px;
        }
        .header .meta {
            opacity: 0.9;
            font-size: 0.9em;
        }
        .summary {
            background: white;
            padding: 25px;
            border-radius: 10px;
            margin-bottom: 20px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .summary h2 {
            margin-bottom: 20px;
            color: #333;
            border-bottom: 2px solid #667eea;
            padding-bottom: 10px;
        }
        .score-card {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 20px;
            margin-bottom: 20px;
        }
        .score-box {
            padding: 20px;
            border-radius: 8px;
            text-align: center;
        }
        .score-box.score {
            background: #e8f4e8;
            border: 2px solid #4caf50;
        }
        .score-box.warning {
            background: #fff8e1;
            border: 2px solid #ff9800;
        }
        .score-box.error {
            background: #ffebee;
            border: 2px solid #f44336;
        }
        .score-value {
            font-size: 3em;
            font-weight: bold;
            display: block;
        }
        .score-label {
            font-size: 0.9em;
            color: #666;
            margin-top: 5px;
        }
        .score-grid {
            display: grid;
            grid-template-columns: repeat(4, 1fr);
            gap: 10px;
        }
        .score-item {
            padding: 10px;
            border-radius: 5px;
            text-align: center;
        }
        .score-item.critical {
            background: #ffebee;
            color: #f44336;
        }
        .score-item.high {
            background: #fff3e0;
            color: #ff5722;
        }
        .score-item.medium {
            background: #fff8e1;
            color: #ffc107;
        }
        .score-item.low {
            background: #e3f2fd;
            color: #2196f3;
        }
        .score-item .count {
            font-size: 1.5em;
            font-weight: bold;
            display: block;
        }
        .score-item .label {
            font-size: 0.8em;
            color: #666;
        }
        .section {
            background: white;
            padding: 25px;
            border-radius: 10px;
            margin-bottom: 20px;
            box-shadow: 0 2px 10px rgba(0,0,0,0.1);
        }
        .section h2 {
            margin-bottom: 20px;
            color: #333;
            border-bottom: 2px solid #667eea;
            padding-bottom: 10px;
        }
        .issue {
            padding: 15px;
            margin-bottom: 15px;
            border-radius: 8px;
            border-left: 4px solid;
        }
        .issue.error {
            background: #ffebee;
            border-color: #f44336;
        }
        .issue.warning {
            background: #fff8e1;
            border-color: #ff9800;
        }
        .issue.info {
            background: #e3f2fd;
            border-color: #2196f3;
        }
        .issue-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 10px;
        }
        .issue-type {
            font-weight: bold;
            font-size: 0.9em;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
        .issue-severity {
            padding: 5px 12px;
            border-radius: 20px;
            font-size: 0.85em;
            font-weight: bold;
            text-transform: uppercase;
        }
        .severity.error {
            background: #f44336;
            color: white;
        }
        .severity.warning {
            background: #ff9800;
            color: white;
        }
        .severity.info {
            background: #2196f3;
            color: white;
        }
        .issue-message {
            color: #333;
            margin-bottom: 10px;
        }
        .issue-component {
            font-size: 0.9em;
            color: #666;
            margin-bottom: 10px;
        }
        .issue-line {
            font-size: 0.85em;
            color: #888;
        }
        .issue-fixed {
            background: #e8f5e9;
            padding: 10px;
            border-radius: 5px;
            margin-top: 10px;
        }
        .issue-fixed strong {
            color: #2e7d32;
        }
        .suggestion {
            padding: 15px;
            margin-bottom: 15px;
            background: #e8f5e9;
            border-radius: 8px;
            border-left: 4px solid #4caf50;
        }
        .suggestion h4 {
            margin-bottom: 10px;
            color: #2e7d32;
        }
        .suggestion p {
            margin-bottom: 10px;
        }
        .suggestion code {
            background: #f5f5f5;
            padding: 2px 8px;
            border-radius: 4px;
            font-family: 'Courier New', monospace;
        }
        .suggestion .command {
            background: #333;
            color: #fff;
            padding: 8px 15px;
            border-radius: 5px;
            display: inline-block;
            margin-top: 10px;
        }
        .dependency-item {
            padding: 12px;
            margin-bottom: 10px;
            border-radius: 8px;
            border-left: 4px solid #666;
        }
        .dependency-item.uptodate {
            border-color: #4caf50;
            background: #f1f8e9;
        }
        .dependency-item.outdated {
            border-color: #ff9800;
            background: #fff8e1;
        }
        .dependency-item.vulnerable {
            border-color: #f44336;
            background: #ffebee;
        }
        .dependency-item.missing {
            border-color: #9c27b0;
            background: #f3e5f5;
        }
        .dependency-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 8px;
        }
        .dependency-name {
            font-weight: bold;
        }
        .dependency-version {
            color: #666;
        }
        .dependency-status {
            padding: 4px 10px;
            border-radius: 15px;
            font-size: 0.75em;
            font-weight: bold;
            text-transform: uppercase;
        }
        .dependency-status.uptodate {
            background: #e8f5e9;
            color: #2e7d32;
        }
        .dependency-status.outdated {
            background: #fff3e0;
            color: #e65100;
        }
        .dependency-status.vulnerable {
            background: #ffebee;
            color: #c62828;
        }
        .dependency-status.missing {
            background: #f3e5f5;
            color: #6a1b9a;
        }
        .dependency-details {
            font-size: 0.9em;
            color: #666;
            padding-left: 20px;
        }
        .dependency-details p {
            margin: 5px 0;
        }
        .footer {
            text-align: center;
            padding: 20px;
            color: #666;
            font-size: 0.9em;
        }
        .badge {
            display: inline-block;
            padding: 3px 8px;
            border-radius: 4px;
            font-size: 0.8em;
            font-weight: bold;
            margin-left: 10px;
        }
        .badge.cve {
            background: #c62828;
            color: white;
        }
        .severity-badge {
            display: inline-block;
            padding: 3px 8px;
            border-radius: 4px;
            font-size: 0.8em;
            font-weight: bold;
            margin-right: 10px;
        }
        .severity-badge.error {
            background: #f44336;
            color: white;
        }
        .severity-badge.warning {
            background: #ff9800;
            color: white;
        }
        .severity-badge.info {
            background: #2196f3;
            color: white;
        }
        .empty-state {
            text-align: center;
            padding: 30px;
            color: #888;
        }
        @media (max-width: 768px) {
            .header h1 {
                font-size: 1.8em;
            }
            .score-grid {
                grid-template-columns: repeat(2, 1fr);
            }
            .score-box {
                padding: 15px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Enva Security Report</h1>
            <div class="meta">
                <p>Generated: %s</p>
                <p>Duration: %s</p>
            </div>
        </div>

        <div class="summary">
            <h2>Summary</h2>
            <div class="score-card">
                <div class="score-box %s">
                    <span class="score-value">%d</span>
                    <span class="score-label">Score</span>
                </div>
                <div class="score-box">
                    <span class="score-label">Status</span>
                    <strong>%s</strong>
                </div>
                <div class="score-box">
                    <span class="score-label">Packages</span>
                    <strong>%d</strong>
                </div>
                <div class="score-box">
                    <span class="score-label">Vulnerabilities</span>
                    <strong>%d</strong>
                </div>
            </div>
            <div class="score-grid">
                <div class="score-item critical">
                    <span class="count">%d</span>
                    <span class="label">Critical</span>
                </div>
                <div class="score-item high">
                    <span class="count">%d</span>
                    <span class="label">High</span>
                </div>
                <div class="score-item medium">
                    <span class="count">%d</span>
                    <span class="label">Medium</span>
                </div>
                <div class="score-item low">
                    <span class="count">%d</span>
                    <span class="label">Low</span>
                </div>
            </div>
        </div>

        <div class="section">
            <h2>Issues <span class="badge">%d</span></h2>
            <div class="issues-list">
` + generateIssuesHTML(result.Issues) + `
            </div>
        </div>

        <div class="section">
            <h2>Suggestions <span class="badge">%d</span></h2>
            <div class="suggestions-list">
` + generateSuggestionsHTML(result.Suggestions) + `
            </div>
        </div>

        <div class="section">
            <h2>Dependencies <span class="badge">%d</span></h2>
            <div class="dependencies-list">
` + generateDependenciesHTML(result.Dependencies) + `
            </div>
        </div>

        <div class="section">
            <h2>Security Vulnerabilities <span class="badge">%d</span></h2>
            <div class="vulnerabilities-list">
` + generateVulnerabilitiesHTML(result.Security) + `
            </div>
        </div>

        <div class="section">
            <h2>Performance <span class="badge">%d</span></h2>
            <div class="performance-list">
` + generatePerformanceHTML(result.Performance) + `
            </div>
        </div>

        <div class="footer">
            <p>Generated by Enva v0.3.0</p>
        </div>
    </div>
</body>
</html>`,
		result.Duration.Round(time.Millisecond).String(),
		getStatusClass(result.Score),
		result.Score,
		result.OverallStatus,
		len(result.Dependencies),
		result.Security.Critical+result.Security.High+result.Security.Medium+result.Security.Low,
		result.Security.Critical,
		result.Security.High,
		result.Security.Medium,
		result.Security.Low,
		len(result.Issues),
		len(result.Suggestions),
		len(result.Dependencies),
		len(result.Security.Vulnerabilities),
		len(result.Performance.Optimizations),
	)
}

func generateIssuesHTML(issues []types.Issue) string {
	if len(issues) == 0 {
		return "<div class=\"empty-state\">No issues found.</div>"
	}

	var html string
	for _, issue := range issues {
		html += fmt.Sprintf(`
                    <div class="issue %s">
                        <div class="issue-header">
                            <span class="issue-type">%s</span>
                            <span class="issue-severity %s">%s</span>
                        </div>
                        <div class="issue-message">%s</div>
`+strings.Repeat(" ", 10)+`
                        <div class="issue-component">Component: <strong>%s</strong></div>
`+strings.Repeat(" ", 10)+`
`+strings.Repeat(" ", 10)+`
`+strings.Repeat(" ", 10)+`
                        <div class="issue-line">Line %d</div>
                    </div>`,
			issue.Severity,
			issue.Type,
			getSeverityClass(issue.Severity),
			issue.Severity,
			issue.Message,
			issue.Component,
			issue.Line,
		)
	}
	return html
}

func generateSuggestionsHTML(suggestions []types.Suggestion) string {
	if len(suggestions) == 0 {
		return "<div class=\"empty-state\">No suggestions available.</div>"
	}

	var html string
	for _, suggestion := range suggestions {
		priorityText := strings.ToUpper(suggestion.Priority)
		html += fmt.Sprintf(`
                    <div class="suggestion">
                        <h4>
                            <span class="severity-badge %s">%s</span>
                            %s
                        </h4>
                        <p>%s</p>
`+strings.Repeat(" ", 10)+`
`+strings.Repeat(" ", 10)+`
                        %s
                    </div>`,
			getPriorityClass(suggestion.Priority),
			priorityText,
			suggestion.Type,
			suggestion.Description,
			getCommandHTML(suggestion.Command),
		)
	}
	return html
}

func generateDependenciesHTML(dependencies []types.Dependency) string {
	if len(dependencies) == 0 {
		return "<div class=\"empty-state\">No dependencies found.</div>"
	}

	var html string
	for _, dep := range dependencies {
		statusClass := getDependencyStatusClass(dep.Status)
		html += fmt.Sprintf(`
                    <div class="dependency-item %s">
                        <div class="dependency-header">
                            <div>
                                <span class="dependency-name">%s</span>
                                <span class="dependency-version">v%s</span>
                            </div>
                            <span class="dependency-status %s">%s</span>
                        </div>
                        <div class="dependency-details">
`+strings.Repeat(" ", 10)+`
`+strings.Repeat(" ", 10)+`
`+strings.Repeat(" ", 10)+`
                        %s
                        </div>
                    </div>`,
			statusClass,
			dep.Name,
			dep.Version,
			getStatusClass(dep.Status),
			dep.Status,
			getDependencyDetailsHTML(dep),
		)
	}
	return html
}

func generateVulnerabilitiesHTML(security *types.SecurityScan) string {
	if security == nil || len(security.Vulnerabilities) == 0 {
		return "<div class=\"empty-state\">No security vulnerabilities found.</div>"
	}

	var html string
	for _, vuln := range security.Vulnerabilities {
		html += fmt.Sprintf(`
                    <div class="issue %s">
                        <div class="issue-header">
                            <div>
                                <span class="issue-type">Vulnerability</span>
                                <span class="issue-severity %s">%s</span>
                            </div>
                        </div>
                        <div class="issue-message">
                            <strong>%s</strong> <span class="badge %s">%s</span>
                        </div>
                        <div class="issue-component">Package: <strong>%s</strong> v%s</div>
                        <div class="issue-description">%s</div>
`+strings.Repeat(" ", 10)+`
`+strings.Repeat(" ", 10)+`
`+strings.Repeat(" ", 10)+`
                        %s
                    </div>`,
			"vulnerability",
			getSeverityClass(vuln.Severity),
			vuln.Severity,
			vuln.ID,
			getSeverityClass(vuln.Severity),
			vuln.ID,
			vuln.Package,
			vuln.Version,
			vuln.Description,
			getFixedVersionHTML(vuln.FixedIn),
		)
	}
	return html
}

func generatePerformanceHTML(performance *types.Performance) string {
	if performance == nil {
		return "<div class=\"empty-state\">No performance analysis available.</div>"
	}

	var html string

	// Unused packages
	if len(performance.UnusedPackages) > 0 {
		html += fmt.Sprintf(`
                    <div class="issue warning">
                        <div class="issue-header">
                            <span class="issue-type">Performance</span>
                            <span class="issue-severity warning">Warning</span>
                        </div>
                        <div class="issue-message">
                            <strong>%d unused package(s) detected</strong>
                        </div>
                        <div class="issue-component">
                            <strong>Unused packages:</strong><br>
                            <ul>
`+strings.Repeat(" ", 10)+`
`+strings.Repeat(" ", 10)+`
                            %s
                            </ul>
                        </div>
                    </div>`,
			len(performance.UnusedPackages),
			generateUnusedPackagesHTML(performance.UnusedPackages),
		)
	}

	// Large packages
	if len(performance.LargePackages) > 0 {
		html += fmt.Sprintf(`
                    <div class="issue warning">
                        <div class="issue-header">
                            <span class="issue-type">Performance</span>
                            <span class="issue-severity warning">Warning</span>
                        </div>
                        <div class="issue-message">
                            <strong>%d large package(s) detected</strong>
                        </div>
                        <div class="issue-component">
                            <strong>Large packages:</strong><br>
                            <ul>
`+strings.Repeat(" ", 10)+`
`+strings.Repeat(" ", 10)+`
                            %s
                            </ul>
                        </div>
                    </div>`,
			len(performance.LargePackages),
			generateLargePackagesHTML(performance.LargePackages),
		)
	}

	// Optimizations
	if len(performance.Optimizations) > 0 {
		html += fmt.Sprintf(`
                    <div class="section">
                        <h2>Optimizations <span class="badge">%d</span></h2>
                        <div class="suggestions-list">
`+strings.Repeat(" ", 10)+`
`+strings.Repeat(" ", 10)+`
                        %s
                        </div>
                    </div>`,
			len(performance.Optimizations),
			generateOptimizationsHTML(performance.Optimizations),
		)
	}

	if html == "" {
		html = "<div class=\"empty-state\">No performance issues or optimization opportunities found.</div>"
	}

	return html
}

func generateUnusedPackagesHTML(packages []string) string {
	var html string
	for _, pkg := range packages {
		html += fmt.Sprintf(`<li><strong>%s</strong></li>`, pkg)
	}
	return html
}

func generateLargePackagesHTML(packages []types.PackageSize) string {
	var html string
	for _, pkg := range packages {
		html += fmt.Sprintf(`<li><strong>%s</strong> - <em>%s</em></li>`, pkg.Name, pkg.Size)
	}
	return html
}

func generateOptimizationsHTML(optimizations []types.Optimization) string {
	var html string
	for _, opt := range optimizations {
		impactClass := strings.ToLower(opt.Impact)
		html += fmt.Sprintf(`
                        <div class="suggestion">
                            <h4>
                                <span class="severity-badge %s">%s</span>
                                <span class="severity-badge %s">%s</span>
                            </h4>
                            <p>%s</p>
                        </div>`,
			getPriorityClass(opt.Priority),
			strings.ToUpper(opt.Type),
			impactClass,
			opt.Impact,
			opt.Description,
		)
	}
	return html
}

func getSeverityClass(severity string) string {
	switch severity {
	case "critical":
		return "error"
	case "high":
		return "warning"
	case "medium":
		return "info"
	case "low":
		return "info"
	default:
		return "info"
	}
}

func getDependencyStatusClass(status string) string {
	switch status {
	case "vulnerable":
		return "vulnerable"
	case "outdated":
		return "outdated"
	case "missing":
		return "missing"
	default:
		return "uptodate"
	}
}

func getStatusClass(score int) string {
	if score >= 90 {
		return "score"
	} else if score >= 70 {
		return "warning"
	} else {
		return "error"
	}
}

func getPriorityClass(priority string) string {
	switch priority {
	case "high":
		return "error"
	case "medium":
		return "warning"
	default:
		return "info"
	}
}

func getCommandHTML(command string) string {
	if command == "" {
		return ""
	}
	return fmt.Sprintf(`<div class="command">%s</div>`, command)
}

func getFixedVersionHTML(fixedIn string) string {
	if fixedIn == "" {
		return ""
	}
	return fmt.Sprintf(`<div class="issue-fixed"><strong>Fixed in: </strong>%s</div>`, fixedIn)
}

func getDependencyDetailsHTML(dep types.Dependency) string {
	var html string
	if dep.Latest != "" {
		html += fmt.Sprintf(`<p><strong>Latest:</strong> v%s</p>`, dep.Latest)
	}
	if dep.RequiredBy != "" {
		html += fmt.Sprintf(`<p><strong>Required by:</strong> %s</p>`, dep.RequiredBy)
	}
	return html
}
