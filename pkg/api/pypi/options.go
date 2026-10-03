package pypi

import (
	"time"
)

// Config holds configuration options for the PyPI client
type Config struct {
	// BaseURL is the base URL for the PyPI API (default: https://pypi.org)
	BaseURL string

	// CacheTTL is the time-to-live for cached responses (default: 5 minutes)
	CacheTTL time.Duration

	// RateLimit is the rate limit for API requests (default: 10 requests/second)
	RateLimit float64

	// MaxCacheSize is the maximum number of entries in the cache (default: 1000)
	MaxCacheSize int

	// Timeout is the request timeout (default: 30 seconds)
	Timeout time.Duration

	// EnableVersionCache enables version caching (default: true)
	EnableVersionCache bool
}

// Default returns a new Config with default values
func Default() *Config {
	return &Config{
		BaseURL:        "https://pypi.org",
		CacheTTL:       5 * time.Minute,
		RateLimit:      10.0,
		MaxCacheSize:   1000,
		Timeout:        30 * time.Second,
		EnableVersionCache: true,
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

// WithVersionCache enables or disables version caching
func (c *Config) WithVersionCache(enable bool) *Config {
	c.EnableVersionCache = enable
	return c
}
