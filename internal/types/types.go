package types

import "time"

// All the structs from before, but in separate package
type ValidationResult struct {
	OverallStatus string        `json:"overall_status"`
	Score         int           `json:"score"`
	Duration      time.Duration `json:"duration"`
	Issues        []Issue       `json:"issues"`
	Suggestions   []Suggestion  `json:"suggestions"`
	ReportPath    string        `json:"report_path"`

	VenvInfo     *VenvInfo     `json:"venv_info,omitempty"`
	Dependencies []Dependency  `json:"dependencies,omitempty"`
	Security     *SecurityScan `json:"security,omitempty"`
	Performance  *Performance  `json:"performance,omitempty"`
}

type VenvInfo struct {
	Path          string `json:"path"`
	PythonVersion string `json:"python_version"`
	PipVersion    string `json:"pip_version"`
	Activated     string `json:"activated"`
	Integrity     string `json:"integrity"`
	Status        string `json:"status"`
}

type Dependency struct {
	Name       string `json:"name"`
	Version    string `json:"version"`
	Latest     string `json:"latest,omitempty"`
	Status     string `json:"status"`
	RequiredBy string `json:"required_by,omitempty"`
}

type SecurityScan struct {
	Critical        int             `json:"critical"`
	High            int             `json:"high"`
	Medium          int             `json:"medium"`
	Low             int             `json:"low"`
	Vulnerabilities []Vulnerability `json:"vulnerabilities,omitempty"`
	Status          string          `json:"status"`
}

type Vulnerability struct {
	ID          string `json:"id"`
	Package     string `json:"package"`
	Version     string `json:"version"`
	Severity    string `json:"severity"`
	Description string `json:"description"`
	FixedIn     string `json:"fixed_in,omitempty"`
}

type Performance struct {
	UnusedPackages []string       `json:"unused_packages,omitempty"`
	LargePackages  []PackageSize  `json:"large_packages,omitempty"`
	Optimizations  []Optimization `json:"optimizations,omitempty"`
	Status         string         `json:"status"`
}

type PackageSize struct {
	Name string `json:"name"`
	Size string `json:"size"`
}

type Optimization struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
}

type Issue struct {
	Type      string `json:"type"`
	Severity  string `json:"severity"`
	Message   string `json:"message"`
	Component string `json:"component,omitempty"`
	Line      int    `json:"line,omitempty"`
}

type Suggestion struct {
	Type        string `json:"type"`
	Description string `json:"description"`
	Command     string `json:"command,omitempty"`
	AutoFixable bool   `json:"auto_fixable"`
	Priority    string `json:"priority"`
}
