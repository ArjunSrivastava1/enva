package validator

import "time"

// ValidationResult holds the complete validation result
type ValidationResult struct {
	OverallStatus string        `json:"overall_status"`
	Score         int           `json:"score"`
	Duration      time.Duration `json:"duration"`
	Issues        []Issue       `json:"issues"`
	Suggestions   []Suggestion  `json:"suggestions"`
	ReportPath    string        `json:"report_path"`

	// Component results
	VenvInfo     *VenvInfo     `json:"venv_info,omitempty"`
	Dependencies []Dependency  `json:"dependencies,omitempty"`
	Security     *SecurityScan `json:"security,omitempty"`
	Performance  *Performance  `json:"performance,omitempty"`
}

// VenvInfo holds virtual environment information
type VenvInfo struct {
	Path          string `json:"path"`
	PythonVersion string `json:"python_version"`
	PipVersion    string `json:"pip_version"`
	Activated     string `json:"activated"` // "activated", "not_activated"
	Integrity     string `json:"integrity"` // "valid", "invalid"
	Status        string `json:"status"`    // "success", "warning", "error"
}

// Dependency represents a Python package
type Dependency struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Latest     string `json:"latest,omitempty"`
	Status     string `json:"status"` // "uptodate", "outdated", "vulnerable", "missing"
	RequiredBy string `json:"required_by,omitempty"`
}

// SecurityScan holds security analysis results
type SecurityScan struct {
	Critical        int             `json:"critical"`
	High            int             `json:"high"`
	Medium          int             `json:"medium"`
	Low             int             `json:"low"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities,omitempty"`
	Status          string          `json:"status"`
}

// Vulnerability represents a security vulnerability
type Vulnerability struct {
	ID          string `json:"id"`
	Package     string `json:"package"`
	Version     string `json:"version"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	FixedIn     string `json:"fixed_in,omitempty"`
}

// Performance holds performance analysis results
type Performance struct {
	UnusedPackages []string       `json:"unused_packages,omitempty"`
	LargePackages  []PackageSize  `json:"large_packages,omitempty"`
	Optimizations  []Optimization `json:"optimizations,omitempty"`
	Status         string         `json:"status"`
}

// PackageSize represents package size information
type PackageSize struct {
	Name string `json:"name"`
	Size string `json:"size"` // e.g., "450MB", "2.1GB"
}

// Optimization represents a performance optimization suggestion
type Optimization struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Impact      string `json:"impact"` // "high", "medium", "low"
}

// Issue represents a validation issue
type Issue struct {
	Type      string `json:"type"`     // "venv", "dependency", "security", "performance"
	Severity  string `json:"severity"` // "error", "warning", "info"
	Message   string `json:"message"`
	Component string `json:"component,omitempty"`
	Line      int    `json:"line,omitempty"`
}

// Suggestion represents a fix suggestion
type Suggestion struct {
	Type        string `json:"type"` // "update", "remove", "add", "fix", "config"
	Description string `json:"description"`
	Command     string `json:"command,omitempty"`
	AutoFixable bool   `json:"auto_fixable"`
	Priority    string `json:"priority"` // "high", "medium", "low"
}
