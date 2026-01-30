# Redis Setup Guide

This guide covers Redis setup for both local development and production deployments with AWS ElastiCache.

## Local Development Setup

### Install Redis

**macOS (using Homebrew):**
```bash
brew install redis
```

**Ubuntu/Debian:**
```bash
sudo apt update
sudo apt install redis-server
```

**CentOS/RHEL:**
```bash
sudo yum install redis
```

**Windows:**
Download from [https://redis.io/download](https://redis.io/download) and follow installation instructions.

### Start Redis

**macOS:**
```bash
brew services start redis
```
Or run manually: `redis-server`

**Linux:**
```bash
sudo systemctl start redis
```
Or run manually: `redis-server`

### Verify Redis is running

```bash
redis-cli ping
```

Should return `PONG`

## AWS ElastiCache Configuration

For production deployments using AWS ElastiCache, configure additional security and connection options:

### Python Configuration

```python
from pycommonlog import Config

config = Config(
    provider="lark",
    # ... other config ...
    provider_config={
        "redis_host": "your-elasticache-endpoint.cache.amazonaws.com",
        "redis_port": 6379,
        "redis_password": "your-auth-token",  # Required for ElastiCache with AUTH enabled
        "redis_ssl": True,                   # Enable SSL/TLS encryption
        "redis_cluster_mode": False,         # Set to True for ElastiCache cluster mode
        "redis_db": 0,                       # Redis database number (usually 0)
    }
)
```

### Go Configuration

```go
import "github.com/alvianhanif/gocommonlog"

cfg := commonlog.Config{
    Provider:   "lark",
    // ... other config ...
    ProviderConfig: map[string]interface{}{
        "redis_host": "your-elasticache-endpoint.cache.amazonaws.com",
        "redis_port": "6379",
        "redis_password": "your-auth-token",  // Required for ElastiCache with AUTH enabled
        "redis_ssl": true,                   // Enable SSL/TLS encryption
        "redis_cluster_mode": false,         // Set to true for ElastiCache cluster mode (not yet implemented)
        "redis_db": 0,                       // Redis database number (usually 0)
    },
}
```

## ElastiCache Security Considerations

- **VPC Access:** Ensure your application has network access to the ElastiCache VPC/subnet
- **Security Groups:** Configure security groups to allow inbound connections on port 6379 (or your custom port)
- **Encryption:** Set `"redis_ssl": true` for encryption in transit
- **Authentication:** Set `"redis_password"` to your ElastiCache AUTH token
- **Cluster Mode:** Currently not supported in the Go version (requires RedisCluster client)

## Dependencies

### Python
For cluster mode support, install the additional dependency:
```bash
pip install redis-py-cluster
```

### Go
Cluster mode support is not yet implemented in the Go version (requires RedisCluster client).

## Troubleshooting

If Redis is not available, Lark token caching will be disabled and tokens will be fetched on every request. Check your application logs for connection errors.

## Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `redis_host` | string | - | Redis server hostname (required) |
| `redis_port` | string | - | Redis server port (required) |
| `redis_password` | string | - | Redis AUTH password |
| `redis_ssl` | bool | false | Enable SSL/TLS encryption |
| `redis_cluster_mode` | bool | false | Enable Redis cluster mode (Python only) |
| `redis_db` | int | 0 | Redis database number |