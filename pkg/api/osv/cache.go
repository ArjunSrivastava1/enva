package osv

import (
	"sync"
	"time"
)

// LRUCache is a simple LRU cache for OSV vulnerability data
type LRUCache struct {
	mu        sync.RWMutex
	cache     map[string]*VulnerabilityCache
	accessMap map[string]time.Time
	maxSize   int
}

// NewLRUCache creates a new LRU cache
func NewLRUCache(maxSize int) *LRUCache {
	return &LRUCache{
		cache:     make(map[string]*VulnerabilityCache),
		accessMap: make(map[string]time.Time),
		maxSize:   maxSize,
	}
}

// Get retrieves a value from the cache
func (c *LRUCache) Get(key string) (*VulnerabilityCache, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	cache, ok := c.cache[key]
	if !ok {
		return nil, false
	}

	// Update access time
	c.accessMap[key] = time.Now()

	return cache, true
}

// Put stores a value in the cache
func (c *LRUCache) Put(key string, value *VulnerabilityCache) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove if exceeds max size
	if len(c.cache) >= c.maxSize {
		c.evictOldest()
	}

	c.cache[key] = value
	c.accessMap[key] = time.Now()
}

// Delete removes a value from the cache
func (c *LRUCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cache, key)
	delete(c.accessMap, key)
}

// Clear removes all entries from the cache
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*VulnerabilityCache)
	c.accessMap = make(map[string]time.Time)
}

// Size returns the current cache size
func (c *LRUCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.cache)
}

// evictOldest evicts the oldest entry from the cache
func (c *LRUCache) evictOldest() {
	var oldestKey string
	var oldestAccess time.Time

	c.mu.RLock()
	for key, access := range c.accessMap {
		if oldestAccess.IsZero() || access.Before(oldestAccess) {
			oldestKey = key
			oldestAccess = access
		}
	}
	c.mu.RUnlock()

	if oldestKey != "" {
		c.mu.Lock()
		delete(c.cache, oldestKey)
		delete(c.accessMap, oldestKey)
		c.mu.Unlock()
	}
}

// TimeBoundedCache is a cache with time-based expiration
type TimeBoundedCache struct {
	mu        sync.RWMutex
	cache     map[string]*VulnerabilityCache
	expiry    map[string]time.Time
	ttl       time.Duration
	maxSize   int
}

// NewTimeBoundedCache creates a new time-bounded cache
func NewTimeBoundedCache(ttl time.Duration, maxSize int) *TimeBoundedCache {
	return &TimeBoundedCache{
		cache:   make(map[string]*VulnerabilityCache),
		expiry:  make(map[string]time.Time),
		ttl:     ttl,
		maxSize: maxSize,
	}
}

// Get retrieves a value from the cache if not expired
func (c *TimeBoundedCache) Get(key string) (*VulnerabilityCache, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cache, ok := c.cache[key]
	if !ok {
		return nil, false
	}

	// Check if expired
	if time.Since(c.expiry[key]) > c.ttl {
		c.mu.Lock()
		delete(c.cache, key)
		delete(c.expiry, key)
		c.mu.Unlock()
		return nil, false
	}

	return cache, true
}

// Put stores a value in the cache with expiration
func (c *TimeBoundedCache) Put(key string, value *VulnerabilityCache) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove if exceeds max size
	if len(c.cache) >= c.maxSize {
		c.evictOldest()
	}

	c.cache[key] = value
	c.expiry[key] = time.Now().Add(c.ttl)
}

// Delete removes a value from the cache
func (c *TimeBoundedCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cache, key)
	delete(c.expiry, key)
}

// Clear removes all entries from the cache
func (c *TimeBoundedCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*VulnerabilityCache)
	c.expiry = make(map[string]time.Time)
}

// Size returns the current cache size (excluding expired entries)
func (c *TimeBoundedCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()
	count := 0
	for _, expiry := range c.expiry {
		if now.Sub(expiry) < c.ttl {
			count++
		}
	}
	return count
}

// evictOldest evicts the oldest entry from the cache
func (c *TimeBoundedCache) evictOldest() {
	// Find and remove the oldest entry
	var oldestKey string
	var oldestExpiry time.Time

	c.mu.RLock()
	for key, expiry := range c.expiry {
		if oldestExpiry.IsZero() || expiry.Before(oldestExpiry) {
			oldestKey = key
			oldestExpiry = expiry
		}
	}
	c.mu.RUnlock()

	if oldestKey != "" {
		c.mu.Lock()
		delete(c.cache, oldestKey)
		delete(c.expiry, oldestKey)
		c.mu.Unlock()
	}
}
