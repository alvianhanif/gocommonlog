// Package commonlog provides a unified logging and alerting library
// supporting multiple providers like Slack and Lark with various send methods.
package types

import (
	"log"
	"os"
)

// AlertLevel defines the severity of the alert
const (
	INFO = iota
	WARN
	ERROR
)

// DebugLogger provides centralized debug logging
var DebugLogger = log.New(os.Stdout, "[COMMONLOG DEBUG] ", log.LstdFlags|log.Lshortfile)

// DebugLog logs debug information if debug mode is enabled
func DebugLog(cfg Config, format string, args ...interface{}) {
	if cfg.Debug {
		DebugLogger.Printf(format, args...)
	}
}

// SendMethod defines supported sending methods
const (
	MethodWebClient = "webclient"
	MethodWebhook   = "webhook"
)

// ChannelResolver defines an interface for resolving channels based on alert levels
type ChannelResolver interface {
	ResolveChannel(level int) string
}

// DefaultChannelResolver provides a simple map-based channel resolution
type DefaultChannelResolver struct {
	ChannelMap     map[int]string
	DefaultChannel string
}

func (r *DefaultChannelResolver) ResolveChannel(level int) string {
	if channel, exists := r.ChannelMap[level]; exists {
		return channel
	}
	return r.DefaultChannel
}

// Config holds configuration for the library
type Config struct {
	Provider        string          // "slack" or "lark"
	SendMethod      string          // "webclient", "webhook", "http"
	Token           string          // API token for SDK/webclient
	SlackToken      string          // Slack-specific token
	LarkToken       LarkTokenConfig // Lark-specific token configuration
	Channel         string          // Default channel or chat ID (used if no resolver)
	ChannelResolver ChannelResolver // Optional resolver for dynamic channel mapping
	ServiceName     string          // Name of the service sending alerts
	Environment     string          // Environment (dev, staging, production)
	RedisHost       string          // Redis host for token caching
	RedisPort       string          // Redis port for token caching
	Debug           bool            // Enable debug logging for all processes
}

// LarkTokenConfig holds Lark app credentials
type LarkTokenConfig struct {
	AppID     string
	AppSecret string
}

// Attachment represents a file attachment
type Attachment struct {
	URL      string // Public URL for external files
	FileName string // Optional file name
	Content  string // Inline content for text attachments
}

// Provider interface for alert providers
type Provider interface {
	Send(level int, message string, attachment *Attachment, cfg Config) error
	SendToChannel(level int, message string, attachment *Attachment, cfg Config, channel string) error
}
