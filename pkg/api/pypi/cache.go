package pypi

import (
	"sync"
	"time"
)

// LRUCache is a simple LRU cache for PyPI API responses
type LRUCache struct {
	mu      sync.RWMutex
	cache   map[string]*VersionInfo
	maxSize int
}

// NewLRUCache creates a new LRU cache
func NewLRUCache(maxSize int) *LRUCache {
	return &LRUCache{
		cache:   make(map[string]*VersionInfo),
		maxSize: maxSize,
	}
}

// Get retrieves a value from the cache
func (c *LRUCache) Get(key string) (*VersionInfo, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	info, ok := c.cache[key]
	if !ok {
		return nil, false
	}

	// Update access time (LRU eviction)
	c.updateAccessTime(key)

	return info, true
}

// Put stores a value in the cache
func (c *LRUCache) Put(key string, value *VersionInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove if exceeds max size
	if len(c.cache) >= c.maxSize {
		c.evictOldest()
	}

	c.cache[key] = value
	c.updateAccessTime(key)
}

// Delete removes a value from the cache
func (c *LRUCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cache, key)
}

// Clear removes all entries from the cache
func (c *LRUCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*VersionInfo)
}

// Size returns the current cache size
func (c *LRUCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.cache)
}

// updateAccessTime updates the access time for a key
func (c *LRUCache) updateAccessTime(key string) {
	// In a real implementation, this would track access time for LRU
	// For now, we just keep the entry
}

// evictOldest evicts the oldest entry from the cache
func (c *LRUCache) evictOldest() {
	// For a simple implementation, we just remove arbitrary entries
	// In a real LRU implementation, we'd track access times
	// For now, we'll just remove the first key we find
	var firstKey string
	for k := range c.cache {
		firstKey = k
		break
	}

	if firstKey != "" {
		delete(c.cache, firstKey)
	}
}

// TimeBoundedCache is a cache with time-based expiration
type TimeBoundedCache struct {
	mu        sync.RWMutex
	cache     map[string]*VersionInfoInfo
	expiry    map[string]time.Time
	ttl       time.Duration
}

// NewTimeBoundedCache creates a new time-bounded cache
func NewTimeBoundedCache(ttl time.Duration, maxSize int) *TimeBoundedCache {
	return &TimeBoundedCache{
		cache:   make(map[string]*VersionInfo),
		expiry:  make(map[string]time.Time),
		ttl:     ttl,
	}
}

// Get retrieves a value from the cache if not expired
func (c *TimeBoundedCache) Get(key string) (*VersionInfo, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	info, ok := c.cache[key]
	if !ok {
		return nil, false
	}

	// Check if expired
	if !ok {
		return nil, false
	}

	if time.Since(c.expiry[key]) > c.ttl {
		c.mu.Lock()
		delete(c.cache, key)
		delete(c.expiry, key)
		c.mu.Unlock()
		return nil, false
	}

	return info, true
}

// Put stores a value in the cache with expiration
func (c *TimeBoundedCache) Put(key string, value *VersionInfo) {
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

	c.cache = make(map[string]*VersionInfo)
	c.expiry = make(map[string]time.Time)
}

// Size returns the current cache size (excluding expired entries)
func (c *TimeBoundedCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()
	count := 0
	for key, expiry := range c.expiry {
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

// VersionCacheInfo is a wrapper for VersionInfo with cache metadata
type VersionCacheInfo struct {
	VersionInfo *VersionInfo
	CacheTime   time.Time
	ExpireTime  time.Time
}
