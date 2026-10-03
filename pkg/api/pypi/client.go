package pypi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// Client handles PyPI API interactions
type Client struct {
	client           *http.Client
	baseURL          string
	cache            map[string]*VersionInfo
	cacheMu          sync.RWMutex
	cacheTTL         time.Duration
	cacheHitCount    int
	cacheMissCount   int
	cacheSizeLimit   int
	rateLimiter      *rate.Limiter
	ctx              context.Context
	cancel           context.CancelFunc
}

// VersionInfo holds information about a package version
type VersionInfo struct {
	Name         string   `json:"name"`
	Version      string   `json:"version"`
	UploadTime   string   `json:"upload_time"`
	Downloads    string   `json:"downloads"`
	HomePage     string   `json:"home_page"`
	RequiresDist []string `json:"requires_dist"`
	License      string   `json:"license"`
	Size         int64    `json:"size"`
}

// ReleaseInfo holds release information from PyPI API
type ReleaseInfo struct {
	Info         VersionInfo  `json:"info"`
	ReleaseURL   string       `json:"urls"`
	RequiredMeta *RequiredMeta `json:"required_meta"`
}

// RequiredMeta holds required metadata
type RequiredMeta struct {
	RequiredDist []string `json:"required_dist"`
}

// NewClient creates a new PyPI API client
func NewClient(ctx context.Context, baseURL string, cacheTTL time.Duration, rateLimit float64, maxCacheSize int) *Client {
	ctx, cancel := context.WithCancel(ctx)
	
	client := &Client{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL:  baseURL,
		cacheTTL: cacheTTL,
		rateLimiter: rate.NewLimiter(rate.Limit(rateLimit), int(rateLimit*2)),
		ctx:      ctx,
		cancel:   cancel,
	}

	return client
}

// GetVersionInfo fetches version information for a package
func (c *Client) GetVersionInfo(ctx context.Context, packageName string) (*VersionInfo, error) {
	// Normalize package name
	packageName = normalizePackageName(packageName)

	// Check cache first
	if info, hit := c.getCached(packageName); hit {
		return info, nil
	}

	// Rate limiting
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Build API URL
	pypiAPI := "https://pypi.org"
	if c.baseURL != "" {
		pypiAPI = c.baseURL
	}
	url := fmt.Sprintf("%s/pypi/%s/json", pypiAPI, packageName)

	// Make API request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "enva/0.2.0")

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
	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	// Get latest version
	versionInfo := &release.Info
	c.cache[packageName] = versionInfo
	c.cacheMu.Unlock()
	c.cacheHitCount++
	c.cacheMissCount++

	return versionInfo, nil
}

// GetVersionInfoWithCache fetches version info with caching
func (c *Client) GetVersionInfoWithCache(packageName string) (*VersionInfo, error) {
	ctx, cancel := context.WithCancel(c.ctx)
	defer cancel()

	info, err := c.GetVersionInfo(ctx, packageName)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// GetLatestVersion returns the latest version of a package
func (c *Client) GetLatestVersion(packageName string) (string, error) {
	ctx, cancel := context.WithCancel(c.ctx)
	defer cancel()

	info, err := c.GetVersionInfo(ctx, packageName)
	if err != nil {
		return "", err
	}

	return info.Version, nil
}

// GetVersions returns all available versions of a package
func (c *Client) GetVersions(ctx context.Context, packageName string) ([]string, error) {
	// Normalize package name
	packageName = normalizePackageName(packageName)

	// Check cache first
	if _, hit := c.getCached(packageName); hit {
		return nil, nil // Already have latest, caller can determine versions
	}

	// Rate limiting
	if err := c.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	// Build API URL
	pypiAPI := "https://pypi.org"
	if c.baseURL != "" {
		pypiAPI = c.baseURL
	}
	url := fmt.Sprintf("%s/pypi/%s/json", pypiAPI, packageName)

	// Make API request
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "enva/0.2.0")

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
	var release ReleaseInfo
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	// Extract all versions from URLs
	var versions []string
	for _, urlInfo := range release.ReleaseURLs {
		// Extract version from URL like https://pypi.org/pypi/requests/2.31.0/json
		if len(urlInfo.URL) > 0 {
			parts := splitURL(urlInfo.URL)
			if len(parts) >= 3 {
				versions = append(versions, parts[2])
			}
		}
	}

	// Sort versions (simple version sorting)
	versions = sortVersions(versions)

	return versions, nil
}

// GetPackageInfo fetches detailed package information
func (c *Client) GetPackageInfo(ctx context.Context, packageName string) (*VersionInfo, error) {
	ctx, cancel := context.WithCancel(c.ctx)
	defer cancel()

	info, err := c.GetVersionInfo(ctx, packageName)
	if err != nil {
		return nil, err
	}

	return info, nil
}

// HasVersion checks if a package has a specific version
func (c *Client) HasVersion(ctx context.Context, packageName, version string) (bool, error) {
	ctx, cancel := context.WithCancel(c.ctx)
	defer cancel()

	info, err := c.GetVersionInfo(ctx, packageName)
	if err != nil {
		return false, err
	}

	return info.Version == version, nil
}

// IsVersionLatest checks if a version is the latest
func (c *Client) IsVersionLatest(ctx context.Context, packageName, version string) (bool, error) {
	ctx, cancel := context.WithCancel(c.ctx)
	defer cancel()

	info, err := c.GetVersionInfo(ctx, packageName)
	if err != nil {
		return false, err
	}

	return info.Version == version, nil
}

// GetVersionsSince fetches versions newer than a given version
func (c *Client) GetVersionsSince(ctx context.Context, packageName, sinceVersion string) ([]string, error) {
	ctx, cancel := context.WithCancel(c.ctx)
	defer cancel()

	versions, err := c.GetVersions(ctx, packageName)
	if err != nil {
		return nil, err
	}

	var newerVersions []string
	for _, v := range versions {
		if versionCompare(v, sinceVersion) > 0 {
			newerVersions = append(newerVersions, v)
		}
	}

	return newerVersions, nil
}

// ClearCache clears the cache
func (c *Client) ClearCache() {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	c.cache = make(map[string]*VersionInfo)
}

// InvalidateCache invalidates cache for a specific package
func (c *Client) InvalidateCache(packageName string) {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()
	delete(c.cache, packageName)
}

// GetCacheStats returns cache statistics
func (c *Client) GetCacheStats() (hitRate float64, size int) {
	c.cacheMu.RLock()
	defer c.cacheMu.RUnlock()

	size = len(c.cache)
	if c.cacheHitCount+c.cacheMissCount > 0 {
		hitRate = float64(c.cacheHitCount) / float64(c.cacheHitCount+c.cacheMissCount)
	}
	return hitRate, size
}

// GetCached returns cached version info for a package
func (c *Client) getCached(packageName string) (*VersionInfo, bool) {
	c.cacheMu.RLock()
	defer c.cacheMu.RUnlock()

	info, ok := c.cache[packageName]
	if !ok {
		return nil, false
	}

	// Check if cache has expired
	if time.Since(info.CacheTime) > c.cacheTTL {
		delete(c.cache, packageName)
		c.cacheMu.Lock()
		c.cacheMu.Unlock()
		return nil, false
	}

	return info, true
}

// SetCacheTime sets the cache time for a package
func (c *Client) SetCacheTime(packageName string, cacheTime time.Time) {
	c.cacheMu.Lock()
	defer c.cacheMu.Unlock()

	if info, ok := c.cache[packageName]; ok {
		info.CacheTime = cacheTime
	}
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

// splitURL splits a URL into parts
func splitURL(urlStr string) []string {
	parts := []string{}
	current := ""
	for i, r := range urlStr {
		if r == '/' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(r)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}

// sortVersions sorts version strings
func sortVersions(versions []string) []string {
	// Simple bubble sort - for production, use a proper version sorting library
	sorted := make([]string, len(versions))
	copy(sorted, versions)

	for i := 0; i < len(sorted)-1; i++ {
		for j := 0; j < len(sorted)-i-1; j++ {
			if versionCompare(sorted[j], sorted[j+1]) < 0 {
				sorted[j], sorted[j+1] = sorted[j+1], sorted[j]
			}
		}
	}

	return sorted
}

// versionCompare compares two version strings
// Returns: -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2
func versionCompare(v1, v2 string) int {
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
