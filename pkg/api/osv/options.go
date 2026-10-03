package osv

import (
	"time"
)

// Config holds configuration options for the OSV client
type Config struct {
	// BaseURL is the base URL for the OSV API (default: https://api.osv.dev)
	BaseURL string

	// CacheTTL is the time-to-live for cached responses (default: 5 minutes)
	CacheTTL time.Duration

	// RateLimit is the rate limit for API requests (default: 10 requests/second)
	RateLimit float64

	// MaxCacheSize is the maximum number of entries in the cache (default: 1000)
	MaxCacheSize int

	// Timeout is the request timeout (default: 30 seconds)
	Timeout time.Duration

	// EnableVulnerabilityCache enables vulnerability caching (default: true)
	EnableVulnerabilityCache bool

	// DefaultEcosystem is the default ecosystem for package queries (default: "pypi")
	DefaultEcosystem string

	// MaxConcurrentRequests limits the number of concurrent API requests (default: 5)
	MaxConcurrentRequests int

	// EnableCVEMapping enables automatic CVE ID extraction from references (default: true)
	EnableCVEMapping bool

	// EnableSeverityMapping enables automatic severity mapping (default: true)
	EnableSeverityMapping bool
}

// Default returns a new Config with default values
func Default() *Config {
	return &Config{
		BaseURL:             "https://api.osv.dev",
		CacheTTL:            5 * time.Minute,
		RateLimit:           10.0,
		MaxCacheSize:        1000,
		Timeout:             30 * time.Second,
		EnableVulnerabilityCache: true,
		DefaultEcosystem:    "pypi",
		MaxConcurrentRequests: 5,
		EnableCVEMapping:    true,
		EnableSeverityMapping: true,
	}
}

// WithBaseURL sets the base URL
func (c *Config) WithBaseURL(baseURL string) *Config {
	c.BaseURL = baseURL
	return c
}

// WithCacheTTL sets the cache TTL
func (c *Config) WithCacheTTL(cacheTTL time.Duration) *Config {
	c.CacheTTL = cacheTTL
	return c
}

// WithRateLimit sets the rate limit
func (c *Config) WithRateLimit(rateLimit float64) *Config {
	c.RateLimit = rateLimit
	return c
}

// WithMaxCacheSize sets the maximum cache size
func (c *Config) WithMaxCacheSize(maxCacheSize int) *Config {
	c.MaxCacheSize = maxCacheSize
	return c
}

// WithTimeout sets the request timeout
func (c *Config) WithTimeout(timeout time.Duration) *Config {
	c.Timeout = timeout
	return c
}

// WithVulnerabilityCache enables or disables vulnerability caching
func (c *Config) WithVulnerabilityCache(enable bool) *Config {
	c.EnableVulnerabilityCache = enable
	return c
}

// WithDefaultEcosystem sets the default ecosystem for package queries
func (c *Config) WithDefaultEcosystem(ecosystem string) *Config {
	c.DefaultEcosystem = ecosystem
	return c
}

// WithMaxConcurrentRequests sets the maximum number of concurrent API requests
func (c *Config) WithMaxConcurrentRequests(maxConcurrent int) *Config {
	c.MaxConcurrentRequests = maxConcurrent
	return c
}

// WithCVEMapping enables or disables automatic CVE ID extraction
func (c *Config) WithCVEMapping(enable bool) *Config {
	c.EnableCVEMapping = enable
	return c
}

// WithSeverityMapping enables or disables automatic severity mapping
func (c *Config) WithSeverityMapping(enable bool) *Config {
	c.EnableSeverityMapping = enable
	return c
}
