package cache

import (
	"testing"
	"time"
)

func TestInMemoryCache_SetAndGet(t *testing.T) {
	cache := NewInMemoryCache()

	// Test basic set and get
	cache.Set("test_key", "test_value", time.Minute)
	value, found := cache.Get("test_key")
	if !found {
		t.Error("Expected to find cached value")
	}
	if value != "test_value" {
		t.Errorf("Expected 'test_value', got '%s'", value)
	}
}

func TestInMemoryCache_GetNonexistent(t *testing.T) {
	cache := NewInMemoryCache()

	value, found := cache.Get("nonexistent")
	if found {
		t.Error("Expected not to find nonexistent key")
	}
	if value != "" {
		t.Errorf("Expected empty string for nonexistent key, got '%s'", value)
	}
}

func TestGlobalCache(t *testing.T) {
	// Test getting global cache
	cache1 := GetGlobalCache()
	cache2 := GetGlobalCache()
	if cache1 != cache2 {
		t.Error("Expected global cache to be singleton")
	}
}
