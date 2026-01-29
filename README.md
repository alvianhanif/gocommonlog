# commonlog (Go)

A unified logging and alerting library for Go, supporting Slack and Lark integrations via WebClient and Webhook. Features configurable providers, alert levels, and file attachment support.

## Installation

Add to your `go.mod`:

```bash
go get github.com/alvianhanif/commonlog/go
```


## Usage

```go
package main

import (
    "github.com/alvianhanif/commonlog/go"
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
        RedisHost:  "localhost", // required for Lark
        RedisPort:  "6379",      // required for Lark
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
    RedisHost: "localhost", // required for Lark
    RedisPort: "6379",
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

When using Lark, the tenant_access_token is cached in Redis. The expiry is set dynamically from the API response minus 10 minutes. You must set `RedisHost` and `RedisPort` in your config.

## Channel Mapping

You can configure different channels for different alert levels using a channel resolver:

```go
package main

import (
    "github.com/alvianhanif/commonlog/go"
    "github.com/alvianhanif/commonlog/go/types"
)

func main() {
    // Create a channel resolver that maps alert levels to different channels
    resolver := &types.DefaultChannelResolver{
        ChannelMap: map[int]string{
            types.INFO:  "#general",
            types.WARN:  "#warnings",
            types.ERROR: "#alerts",
        },
        DefaultChannel: "#general",
    }

    // Create config with channel resolver
    config := types.Config{
        Provider:        "slack",
        SendMethod:      types.MethodWebClient,
        Token:           "xoxb-your-slack-bot-token",
        ChannelResolver: resolver,
        ServiceName:     "user-service",
        Environment:     "production",
    }

    logger := commonlog.NewLogger(config)

    // These will go to different channels based on level
    logger.Send(types.INFO, "Info message")    // goes to #general
    logger.Send(types.WARN, "Warning message") // goes to #warnings
    logger.Send(types.ERROR, "Error message")  // goes to #alerts
}
```

### Custom Channel Resolver

You can implement custom channel resolution logic:

```go
type CustomResolver struct{}

func (r *CustomResolver) ResolveChannel(level int) string {
    switch level {
    case types.ERROR:
        return "#critical-alerts"
    case types.WARN:
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
cd go
go test
```

## API Reference

### Types

- `Config`: Configuration struct
- `Attachment`: File attachment struct
- `Provider`: Interface for alert providers

### Constants

- `MethodWebClient`: Send method (token-based authentication)
- `INFO`, `WARN`, `ERROR`: Alert levels

### Functions

- `NewLogger(cfg Config) *Logger`: Create a new logger
- `(*Logger) Send(level int, message string, attachment *Attachment, trace string)`: Send alert with optional trace
