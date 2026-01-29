# gocommonlog

[![CI](https://github.com/alvianhanif/gocommonlog/actions/workflows/ci.yml/badge.svg)](https://github.com/alvianhanif/gocommonlog/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/alvianhanif/gocommonlog)](https://goreportcard.com/report/github.com/alvianhanif/gocommonlog)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A unified logging and alerting library for Go, supporting Slack and Lark integrations via WebClient and Webhook. Features configurable providers, alert levels, and file attachment support.

## Installation

Add to your `go.mod`:

```bash
go get github.com/alvianhanif/gocommonlog
```


## Usage

```go
package main

import (
    "github.com/alvianhanif/gocommonlog"
)

func main() {
    cfg := commonlog.Config{
        Provider:   "lark", // or "slack"
        SendMethod: commonlog.MethodWebClient,
        Token:      "app_id++app_secret", // for Lark, use "app_id++app_secret" format
        SlackToken: "xoxb-your-slack-token", // dedicated Slack token
        LarkToken: commonlog.LarkTokenConfig{ // dedicated Lark token
            AppID:     "your-app-id",
            AppSecret: "your-app-secret",
        },
        Channel:    "your_lark_channel_id",
        ProviderConfig: map[string]interface{}{
            "redis_host": "localhost", // required for Lark
            "redis_port": "6379",      // required for Lark
        },
    }
    logger := commonlog.NewLogger(cfg)

    // Send error with attachment
    if err := logger.Send(commonlog.ERROR, "System error occurred", &commonlog.Attachment{URL: "https://example.com/log.txt"}); err != nil {
        log.Printf("Failed to send alert: %v", err)
    }

    // Send info (logs only)
    logger.Send(commonlog.INFO, "Info message")

    // Send to a specific channel
    if err := logger.SendToChannel(commonlog.ERROR, "Send to another channel", nil, "", "another-channel-id"); err != nil {
        log.Printf("Failed to send alert: %v", err)
    }

    // Send to a different provider dynamically
    if err := logger.CustomSend("slack", commonlog.ERROR, "Message via Slack", nil, "", "slack-channel"); err != nil {
        log.Printf("Failed to send alert: %v", err)
    }
}
```

## Send Methods

commonlog supports two send methods: WebClient (API-based) and Webhook (simple HTTP POST).

### WebClient Usage

WebClient uses the full API with authentication tokens:

```go
cfg := commonlog.Config{
    Provider:   "lark",
    SendMethod: commonlog.MethodWebClient,
    Token:      "app_id++app_secret", // for Lark
    SlackToken: "xoxb-your-slack-token", // for Slack
    LarkToken: commonlog.LarkTokenConfig{
        AppID:     "your-app-id",
        AppSecret: "your-app-secret",
    },
    Channel:   "your_channel",
    ProviderConfig: map[string]interface{}{
        "redis_host": "localhost", // required for Lark
        "redis_port": "6379",      // required for Lark
    },
}
```

### Webhook Usage

Webhook is simpler and requires only a webhook URL:

```go
cfg := commonlog.Config{
    Provider:   "slack",
    SendMethod: commonlog.MethodWebhook,
    Token:      "https://hooks.slack.com/services/YOUR/WEBHOOK/URL",
    Channel:    "optional-channel-override", // optional
}
```

### Lark Token Caching

When using Lark, the tenant_access_token is cached in Redis. The expiry is set dynamically from the API response minus 10 minutes. You must set `redis_host` and `redis_port` in your `ProviderConfig`.

## Channel Mapping

You can configure different channels for different alert levels using a channel resolver:

```go
package main

import (
    "github.com/alvianhanif/gocommonlog"
)

func main() {
    // Create a channel resolver that maps alert levels to different channels
    resolver := &commonlog.DefaultChannelResolver{
        ChannelMap: map[int]string{
            commonlog.INFO:  "#general",
            commonlog.WARN:  "#warnings",
            commonlog.ERROR: "#alerts",
        },
        DefaultChannel: "#general",
    }

    // Create config with channel resolver
    config := commonlog.Config{
        Provider:        "slack",
        SendMethod:      commonlog.MethodWebClient,
        Token:           "xoxb-your-slack-bot-token",
        ChannelResolver: resolver,
        ServiceName:     "user-service",
        Environment:     "production",
    }

    logger := commonlog.NewLogger(config)

    // These will go to different channels based on level
    logger.Send(commonlog.INFO, "Info message")    // goes to #general
    logger.Send(commonlog.WARN, "Warning message") // goes to #warnings
    logger.Send(commonlog.ERROR, "Error message")  // goes to #alerts
}
```

### Custom Channel Resolver

You can implement custom channel resolution logic:

```go
type CustomResolver struct{}

func (r *CustomResolver) ResolveChannel(level int) string {
    switch level {
    case commonlog.ERROR:
        return "#critical-alerts"
    case commonlog.WARN:
        return "#monitoring"
    default:
        return "#general"
    }
}
```

## Configuration Options

### Common Settings

- **Provider**: `"slack"` or `"lark"`
- **SendMethod**: `MethodWebClient` (token-based authentication)
- **Channel**: Target channel or chat ID (used if no resolver)
- **ChannelResolver**: Optional resolver for dynamic channel mapping
- **ServiceName**: Name of the service sending alerts
- **Environment**: Environment (dev, staging, production)
- **Debug**: `true` to enable detailed debug logging of all internal processes

### Provider-Specific

- **Token**: API token for WebClient authentication (required)
- **SlackToken**: Dedicated Slack token (optional)
- **LarkToken**: Dedicated Lark token configuration (optional)
- **ProviderConfig**: Map of provider-specific settings (e.g., Redis config for Lark)

## Alert Levels

- **INFO**: Logs locally only
- **WARN**: Logs + sends alert
- **ERROR**: Always sends alert

## File Attachments

Provide a public URL. The library appends it to the message for simplicity.

```go
attachment := &commonlog.Attachment{URL: "https://example.com/log.txt"}
logger.Send(commonlog.ERROR, "Error with log", attachment, "")
```

## Trace Log Section

When `IncludeTrace` is set to `true`, you can pass trace information as the fourth parameter to `Send()`:

```go
trace := "goroutine 1 [running]:\nmain.main()\n    /app/main.go:15 +0x2f"
logger.Send(commonlog.ERROR, "System error occurred", nil, trace)
```

This will format the trace as a code block in the alert message.

## Testing

```bash
go test
```

## API Reference

### Types

- `Config`: Configuration struct
- `Attachment`: File attachment struct
- `Provider`: Interface for alert providers
- `LarkTokenConfig`: Lark app credentials
- `ChannelResolver`: Interface for channel resolution
- `DefaultChannelResolver`: Default channel resolver implementation

### Constants

- `MethodWebClient`: Send method (token-based authentication)
- `MethodWebhook`: Send method (simple HTTP POST)
- `INFO`, `WARN`, `ERROR`: Alert levels

### Functions

- `NewLogger(cfg Config) *Logger`: Create a new logger
- `(*Logger) Send(level int, message string, attachment *Attachment, trace string) error`: Send alert with optional attachment and trace
- `(*Logger) SendToChannel(level int, message string, attachment *Attachment, trace string, channel string) error`: Send alert to specific channel
- `(*Logger) CustomSend(provider string, level int, message string, attachment *Attachment, trace string, channel string) error`: Send alert with custom provider
