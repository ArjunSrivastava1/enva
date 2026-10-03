package osv

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"

	"enva/internal/types"
	"golang.org/x/time/rate"
)

// Client handles OSV API interactions
type Client struct {
	client           *http.Client
	baseURL          string
	cache            *TimeBoundedCache
	cacheMu          sync.RWMutex
	cacheTTL         time.Duration
	cacheHitCount    int
	cacheMissCount   int
	cacheSizeLimit   int
	rateLimiter      *rate.Limiter
	ctx              context.Context
	cancel           context.CancelFunc
}

// NewClient creates a new OSV API client
func NewClient(ctx context.Context, baseURL string, cacheTTL time.Duration, rateLimit float64, maxCacheSize int) *Client {
	ctx, cancel := context.WithCancel(ctx)

	client := &Client{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:  baseURL,
		cache:    NewTimeBoundedCache(cacheTTL, maxCacheSize),
		rateLimiter: rate.NewLimiter(rate.Limit(rateLimit), int(rateLimit*2)),
		ctx:      ctx,
		cancel:   cancel,
	}

	return client
}

// GetVulnerabilities fetches vulnerabilities for a package and ecosystem
func (c *Client) GetVulnerabilities(ctx context.Context, packageName, ecosystem string) ([]*Vulnerability, error) {
	// Normalize package name
	packageName = normalizePackageName(packageName)

	// Check cache first
	key := fmt.Sprintf("%s:%s", packageName, ecosystem)
	if cache, hit := c.cache.Get(key); hit {
		return cache.Vulnerabilities, nil
	}

	// Rate limiting
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Build API URL
	osvAPI := c.baseURL + "/v1"
	urlStr := fmt.Sprintf("%s?q=%s", osvAPI, fmt.Sprintf("package:%s:%s", ecosystem, packageName))

	// Make API request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "enva/0.3.0")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle HTTP errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result struct {
		Vulns []*Vulnerability `json:"vulns"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	// Cache the results
	c.cache.Put(key, &VulnerabilityCache{
		PackageName:    packageName,
		Ecosystem:      ecosystem,
		Vulnerabilities: result.Vulns,
		ExpireTime:     time.Now().Add(c.cacheTTL),
	})

	return result.Vulns, nil
}

// GetVulnerability fetches a specific vulnerability by ID
func (c *Client) GetVulnerability(ctx context.Context, vulnID string) (*Vulnerability, error) {
	// Build API URL
	osvAPI := c.baseURL + "/v1"
	urlStr := fmt.Sprintf("%s/osv/%s", osvAPI, vulnID)

	// Make API request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "enva/0.3.0")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle HTTP errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var vuln Vulnerability
	if err := json.NewDecoder(resp.Body).Decode(&vuln); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	return &vuln, nil
}

// SearchVulnerabilities searches for vulnerabilities by keyword
func (c *Client) SearchVulnerabilities(ctx context.Context, query string) ([]*Vulnerability, error) {
	// Build API URL
	osvAPI := c.baseURL + "/v1"
	urlStr := fmt.Sprintf("%s?q=%s", osvAPI, url.QueryEscape(query))

	// Make API request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "enva/0.3.0")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle HTTP errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result struct {
		Vulns []*Vulnerability `json:"vulns"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	return result.Vulns, nil
}

// GetVulnerabilitiesForPackage fetches vulnerabilities for a specific package and ecosystem, returns internal types
func (c *Client) GetVulnerabilitiesForPackage(ctx context.Context, packageName, ecosystem string) ([]*types.Vulnerability, error) {
	// Normalize package name
	packageName = normalizePackageName(packageName)

	// Check cache first
	key := fmt.Sprintf("%s:%s", packageName, ecosystem)
	if cache, hit := c.cache.Get(key); hit {
		// Convert OSV vulnerabilities to internal types
		var vulns []*types.Vulnerability
		for _, osvVuln := range cache.Vulnerabilities {
			vuln := convertToInternalVulnerability(osvVuln, packageName)
			vulns = append(vulns, vuln)
		}
		return vulns, nil
	}

	// Rate limiting
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Build API URL
	osvAPI := "https://api.osv.dev"
	if c.baseURL != "" {
		osvAPI = c.baseURL
	}
	url := fmt.Sprintf("%s/v1?q=%s", osvAPI, fmt.Sprintf("package:%s:%s", ecosystem, packageName))

	// Make API request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "enva/0.3.0")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle HTTP errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var result struct {
		Vulns []*Vulnerability `json:"vulns"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	// Convert OSV vulnerabilities to internal types
	var vulns []*types.Vulnerability
	for _, osvVuln := range result.Vulns {
		vuln := convertToInternalVulnerability(osvVuln, packageName)
		vulns = append(vulns, vuln)
	}

	// Cache the raw OSV data
	c.cache.Put(key, &VulnerabilityCache{
		PackageName:    packageName,
		Ecosystem:      ecosystem,
		Vulnerabilities: result.Vulns,
		ExpireTime:     time.Now().Add(c.cacheTTL),
	})

	return vulns, nil
}

// ScanAllPackages scans all packages for vulnerabilities concurrently
func (c *Client) ScanAllPackages(ctx context.Context, packages []types.Dependency) ([]*types.Vulnerability, error) {
	var allVulns []*types.Vulnerability
	var mu sync.Mutex

	// Create a channel for concurrent scanning
	sem := make(chan struct{}, 5) // Limit to 5 concurrent requests

	var wg sync.WaitGroup
	for _, dep := range packages {
		// Skip packages without version
		if dep.Version == "" || dep.Version == "editable" || dep.Version == "direct_url" {
			continue
		}

		wg.Add(1)
		go func(dep types.Dependency) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			vulns, err := c.GetVulnerabilitiesForPackage(ctx, dep.Name, "pypi")
			if err != nil {
				// Log error but continue with other packages
				return
			}

			mu.Lock()
			allVulns = append(allVulns, vulns...)
			mu.Unlock()
		}(dep)
	}

	wg.Wait()

	// Filter out vulnerabilities that don't affect the installed version
	filteredVulns := filterVulnerabilitiesByVersion(allVulns, packages)

	return filteredVulns, nil
}

// ClearCache clears the cache
func (c *Client) ClearCache() {
	c.cache.Clear()
	c.cacheMu.Lock()
	c.cacheMu.Unlock()
}

// InvalidateCache invalidates cache for a specific package
func (c *Client) InvalidateCache(packageName string) {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	c.cache.Delete(packageName)
}

// GetCacheStats returns cache statistics
func (c *Client) GetCacheStats() (hitRate float64, size int) {
	c.cacheMu.RLock()
	defer c.cacheMu.RUnlock()

	size = c.cache.Size()
	if c.cacheHitCount+c.cacheMissCount > 0 {
		hitRate = float64(c.cacheHitCount) / float64(c.cacheHitCount+c.cacheMissCount)
	}
	return hitRate, size
}

// Close closes the client and cancels context
func (c *Client) Close() error {
	c.cancel()
	return nil
}

// normalizePackageName normalizes a package name according to PEP 503
func normalizePackageName(name string) string {
	// Convert to lowercase and replace runs of non-alphanumeric chars with -
	name = toLowerCase(name)
	name = normalizeName(name)
	return name
}

// toLowerCase converts a string to lowercase
func toLowerCase(s string) string {
	result := make([]rune, len(s))
	for i, r := range s {
		result[i] = toLower(r)
	}
	return string(result)
}

// toLower converts a single rune to lowercase
func toLower(r rune) rune {
	if r >= 'A' && r <= 'Z' {
		return r + ('a' - 'A')
	}
	return r
}

// normalizeName replaces runs of non-alphanumeric characters with single hyphens
func normalizeName(name string) string {
	result := []rune{}
	prevIsAlnum := false

	for _, r := range name {
		isAlnum := isAlphanumeric(r)
		if isAlnum {
			result = append(result, r)
			prevIsAlnum = true
		} else if !prevIsAlnum {
			result = append(result, '-')
			prevIsAlnum = false
		}
	}

	return string(result)
}

// isAlphanumeric checks if a rune is alphanumeric
func isAlphanumeric(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')
}

// convertToInternalVulnerability converts OSV vulnerability to internal type
func convertToInternalVulnerability(osvVuln *Vulnerability, packageName string) *types.Vulnerability {
	vuln := &types.Vulnerability{
		ID:          osvVuln.ID,
		Package:     packageName,
		Version:     "", // Will be set by the caller
		Severity:    MapVulnerabilityToSeverity(osvVuln),
		Description: osvVuln.Description,
		FixedIn:     "", // Will be set based on affected ranges
	}

	// Set FixedIn from affected ranges
	fixedVersion := GetVulnerabilityFixedVersion(&osvVuln.Affected[0])
	if fixedVersion != "" {
		vuln.FixedIn = fixedVersion
	} else {
		// Try to get the last vulnerable version as a fallback
		lastVulnVersion := GetVulnerabilityLastVulnerableVersion(&osvVuln.Affected[0])
		if lastVulnVersion != "" {
			// This is the last vulnerable version, not the fixed version
			// For now, we'll use "unknown" as a placeholder
			vuln.FixedIn = "unknown"
		}
	}

	// Add CVE IDs if available
	cveMap := MapVulnerabilitiesToCVEs([]*types.Vulnerability{vuln})
	for _, cve := range cveMap[vuln.ID] {
		vuln.CVEs = append(vuln.CVEs, cve)
	}

	return vuln
}

// filterVulnerabilitiesByVersion filters vulnerabilities that affect the installed version
func filterVulnerabilitiesByVersion(allVulns []*types.Vulnerability, packages []types.Dependency) []*types.Vulnerability {
	var filtered []*types.Vulnerability
	installedVersions := make(map[string]string)
	for _, dep := range packages {
		installedVersions[dep.Name] = dep.Version
	}

	for _, vuln := range allVulns {
		// Check if we have an installed version for this package
		installedVersion, ok := installedVersions[vuln.Package]
		if !ok {
			continue
		}

		// Check if the installed version is vulnerable
		if vuln.FixedIn != "" && compareVersions(installedVersion, vuln.FixedIn) < 0 {
			// Installed version is older than the fixed version, so it's vulnerable
			vuln.Version = installedVersion
			filtered = append(filtered, vuln)
		} else if vuln.FixedIn == "unknown" {
			// No fixed version known, assume vulnerable if version matches vulnerable range
			vuln.Version = installedVersion
			filtered = append(filtered, vuln)
		}
	}

	return filtered
}

// compareVersions compares two version strings
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func compareVersions(v1, v2 string) int {
	parts1 := parseVersionParts(v1)
	parts2 := parseVersionParts(v2)

	// Pad with zeros to match lengths
	for len(parts1) < len(parts2) {
		parts1 = append(parts1, 0)
	}
	for len(parts2) < len(parts1) {
		parts2 = append(parts2, 0)
	}

	for i := 0; i < len(parts1); i++ {
		p1 := parts1[i]
		p2 := parts2[i]

		// Compare as integers if both are numeric
		if p1 != "" && p2 != "" && isNumeric(p1) && isNumeric(p2) {
			if p1 > p2 {
				return 1
			} else if p1 < p2 {
				return -1
			}
		} else if p1 != p2 {
			// String comparison for non-numeric parts
			if p1 > p2 {
				return 1
			} else if p1 < p2 {
				return -1
			}
		}
	}

	return 0
}

// parseVersionParts parses a version string into numeric parts
func parseVersionParts(version string) []int {
	parts := []int{}
	fields := []string{}

	// Split by dots
	current := ""
	for _, r := range version {
		if r == '.' {
			fields = append(fields, current)
			current = ""
		} else {
			current += string(r)
		}
	}
	fields = append(fields, current)

	for _, field := range fields {
		field = trimVersionField(field)
		if num, err := strconv.Atoi(field); err == nil {
			parts = append(parts, num)
		} else {
			parts = append(parts, 0)
		}
	}

	return parts
}

// trimVersionField trims a version field
func trimVersionField(field string) string {
	// Remove leading zeros
	field = trimLeadingZeros(field)
	return field
}

// trimLeadingZeros removes leading zeros from a string
func trimLeadingZeros(s string) string {
	result := ""
	foundNonZero := false
	for _, r := range s {
		if r == '0' {
			if !foundNonZero {
				continue
			}
		} else {
			foundNonZero = true
			result += string(r)
		}
	}
	return result
}

// isNumeric checks if a string is a numeric value
func isNumeric(s string) bool {
	if _, err := strconv.Atoi(s); err == nil {
		return true
	}
	return false
}
