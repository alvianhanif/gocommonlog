package cache

import (
	"fmt"
	"sync"
	"time"
)

// Cache provides a unified interface for caching operations
type Cache interface {
	Get(key string) (string, bool)
	Set(key, value string, duration time.Duration)
	Delete(key string)
}

// InMemoryCache provides thread-safe in-memory caching with automatic cleanup
type InMemoryCache struct {
	data sync.Map // key -> cacheItem
}

type cacheItem struct {
	value  string
	expiry time.Time
}

// NewInMemoryCache creates a new in-memory cache instance
func NewInMemoryCache() *InMemoryCache {
	cache := &InMemoryCache{}
	// Start cleanup goroutine
	go cache.cleanupWorker()
	return cache
}

// Get retrieves a value from the cache
func (c *InMemoryCache) Get(key string) (string, bool) {
	value, ok := c.data.Load(key)
	if !ok {
		return "", false
	}
	item := value.(cacheItem)
	if time.Now().After(item.expiry) {
		// Expired, remove it
		c.data.Delete(key)
		return "", false
	}
	return item.value, true
}

// Set stores a value in the cache with expiration
func (c *InMemoryCache) Set(key, value string, duration time.Duration) {
	item := cacheItem{
		value:  value,
		expiry: time.Now().Add(duration),
	}
	c.data.Store(key, item)
}

// Delete removes a value from the cache
func (c *InMemoryCache) Delete(key string) {
	c.data.Delete(key)
}

func (c *InMemoryCache) cleanupWorker() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		c.cleanupExpired()
	}
}

func (c *InMemoryCache) cleanupExpired() {
	now := time.Now()
	expiredKeys := make([]string, 0)

	c.data.Range(func(key, value interface{}) bool {
		item := value.(cacheItem)
		if now.After(item.expiry) {
			expiredKeys = append(expiredKeys, key.(string))
		}
		return true
	})

	for _, key := range expiredKeys {
		c.data.Delete(key)
	}

	if len(expiredKeys) > 0 {
		fmt.Printf("[Cache] Cleaned up %d expired entries from memory cache\n", len(expiredKeys))
	}
}

// Global cache instance
var globalCache Cache = NewInMemoryCache()

// GetGlobalCache returns the global cache instance
func GetGlobalCache() Cache {
	return globalCache
}

// SetGlobalCache allows setting a custom cache implementation (useful for testing or Redis integration)
func SetGlobalCache(c Cache) {
	globalCache = c
}