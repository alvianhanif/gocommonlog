# Cache Package

This package provides caching utilities for the commonlog library providers.

## Features

- **Thread-safe in-memory caching** with automatic cleanup of expired entries
- **Unified Cache interface** for easy swapping between different cache implementations
- **Background cleanup** to prevent memory leaks
- **Global cache instance** for easy access across providers

## Usage

### Basic Usage

```go
import "github.com/alvianhanif/gocommonlog/cache"

// Get the global cache instance
cache := cache.GetGlobalCache()

// Set a value with expiration (30 minutes)
cache.Set("my_key", "my_value", 30*time.Minute)

// Get a value
if value, found := cache.Get("my_key"); found {
    fmt.Printf("Found: %s\n", value)
} else {
    fmt.Println("Not found or expired")
}

// Delete a value
cache.Delete("my_key")
```

## Cache Interface

The `Cache` interface allows for different cache implementations:

```go
type Cache interface {
    Get(key string) (string, bool)
    Set(key, value string, duration time.Duration)
    Delete(key string)
}
```

## Custom Cache Implementation

You can implement custom cache backends (Redis, database, etc.) by implementing the `Cache` interface and setting it as the global cache:

```go
// Set a custom cache implementation
cache.SetGlobalCache(myCustomCache)
```

## Automatic Cleanup

The in-memory cache automatically cleans up expired entries every 5 minutes in a background goroutine. This prevents memory leaks while maintaining performance.</content>
<parameter name="filePath">/Users/pid-alvian/Documents/alvian/gocommonlog/cache/README.md
