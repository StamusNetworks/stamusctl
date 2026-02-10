package completion

import (
	"sync"
	"time"
)

// cacheEntry holds a cached value with its expiration time
type cacheEntry struct {
	value      []string
	expiration time.Time
}

// Cache provides a TTL-based cache for completion results
type Cache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
}

// globalCache is the singleton cache instance
var globalCache = &Cache{
	entries: make(map[string]cacheEntry),
}

// GetCache returns the global cache instance
func GetCache() *Cache {
	return globalCache
}

// Get retrieves a cached value if it exists and is not expired
func (c *Cache) Get(key string) ([]string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.expiration) {
		return nil, false
	}

	return entry.value, true
}

// Set stores a value in the cache with the given TTL
func (c *Cache) Set(key string, value []string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = cacheEntry{
		value:      value,
		expiration: time.Now().Add(ttl),
	}
}

// Clear removes all entries from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]cacheEntry)
}

// Cleanup removes expired entries from the cache
func (c *Cache) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.After(entry.expiration) {
			delete(c.entries, key)
		}
	}
}

// TTL constants for different completion types
const (
	// ConfigKeysTTL is the TTL for config keys cache (short for freshness)
	ConfigKeysTTL = 5 * time.Second
	// BackupsTTL is the TTL for backups list cache
	BackupsTTL = 30 * time.Second
	// ConfigsTTL is the TTL for configs list cache
	ConfigsTTL = 30 * time.Second
)
