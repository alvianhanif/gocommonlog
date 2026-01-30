package gocommonlog

import (
	"log"

	"github.com/alvianhanif/gocommonlog/providers"
	"github.com/alvianhanif/gocommonlog/types"
)

// ====================
// Main Logger
// ====================

// createProvider creates a provider instance by name
func createProvider(providerName string) types.Provider {
	switch providerName {
	case "slack":
		return &providers.SlackProvider{}
	case "lark":
		return &providers.LarkProvider{}
	default:
		return &providers.SlackProvider{}
	}
}

// Logger is the main struct
type Logger struct {
	config   types.Config
	provider types.Provider
}

// NewLogger creates a new Logger with the appropriate provider
func NewLogger(cfg types.Config) *Logger {
	// Populate ProviderConfig with top-level fields for backward compatibility
	if cfg.ProviderConfig == nil {
		cfg.ProviderConfig = make(map[string]interface{})
	}
	if cfg.Provider != "" {
		cfg.ProviderConfig["provider"] = cfg.Provider
	}
	if cfg.Token != "" {
		cfg.ProviderConfig["token"] = cfg.Token
	}
	if cfg.SlackToken != "" {
		cfg.ProviderConfig["slack_token"] = cfg.SlackToken
	}
	if cfg.LarkToken.AppID != "" || cfg.LarkToken.AppSecret != "" {
		cfg.ProviderConfig["lark_token"] = cfg.LarkToken
	}

	if _, ok := cfg.ProviderConfig["provider"]; !ok {
		cfg.ProviderConfig["provider"] = "slack"  // default
	}

	providerName, ok := cfg.ProviderConfig["provider"].(string)
	if !ok {
		providerName = "slack"  // fallback
	}
	provider := createProvider(providerName)
	logger := &Logger{config: cfg, provider: provider}

	types.DebugLog(cfg, "Created new logger with provider: %s, send method: %s, debug: %t",
		providerName, cfg.SendMethod, cfg.Debug)

	return logger
}

// resolveChannel resolves the channel for the given alert level
func (l *Logger) resolveChannel(level int) string {
	if l.config.ChannelResolver != nil {
		return l.config.ChannelResolver.ResolveChannel(level)
	}
	return l.config.Channel
}

// Send sends a message with alert level, optional attachment, and optional trace log
func (l *Logger) Send(level int, message string, attachment *types.Attachment, trace string) error {
	return l.SendToChannel(level, message, attachment, trace, "")
}

// SendToChannel sends a message to a specific channel, overriding the default/channel resolver
func (l *Logger) SendToChannel(level int, message string, attachment *types.Attachment, trace string, channel string) error {
	types.DebugLog(l.config, "SendToChannel called with level: %d, message length: %d, channel: %s, has attachment: %t, has trace: %t",
		level, len(message), channel, attachment != nil, trace != "")

	if level == types.INFO {
		log.Printf("[INFO] %s", message)
		types.DebugLog(l.config, "INFO level message logged locally, skipping provider send")
		return nil
	}

	resolvedChannel := channel
	if resolvedChannel == "" {
		resolvedChannel = l.resolveChannel(level)
		types.DebugLog(l.config, "Resolved channel using resolver: %s", resolvedChannel)
	} else {
		types.DebugLog(l.config, "Using provided channel: %s", resolvedChannel)
	}

	sendConfig := l.config
	sendConfig.Channel = resolvedChannel

	if trace != "" {
		types.DebugLog(l.config, "Processing trace attachment, trace length: %d", len(trace))
		traceAttachment := &types.Attachment{
			FileName: "trace.log",
			Content:  trace,
		}
		if attachment != nil {
			if attachment.Content != "" {
				attachment.Content += "\n\n--- Trace Log ---\n" + trace
				types.DebugLog(l.config, "Appended trace to existing attachment content")
			} else {
				attachment.Content = trace
				attachment.FileName = "trace.log"
				types.DebugLog(l.config, "Set trace as attachment content")
			}
		} else {
			attachment = traceAttachment
			types.DebugLog(l.config, "Created new trace attachment")
		}
	}

	types.DebugLog(l.config, "Calling provider.SendToChannel with resolved channel: %s", resolvedChannel)
	err := l.provider.SendToChannel(level, message, attachment, sendConfig, resolvedChannel)
	if err != nil {
		types.DebugLog(l.config, "Provider.SendToChannel failed: %v", err)
	} else {
		types.DebugLog(l.config, "Provider.SendToChannel completed successfully")
	}
	return err
}

// CustomSend sends a message with a custom provider, allowing override of the default provider
func (l *Logger) CustomSend(provider string, level int, message string, attachment *types.Attachment, trace string, channel string) error {
	types.DebugLog(l.config, "CustomSend called with custom provider: %s, level: %d, message length: %d",
		provider, level, len(message))

	customProvider := createProvider(provider)
	if customProvider == nil {
		log.Printf("[ERROR] Unknown provider: %s, defaulting to slack", provider)
		customProvider = createProvider("slack")
		types.DebugLog(l.config, "Unknown provider '%s', defaulted to slack", provider)
	} else {
		types.DebugLog(l.config, "Created custom provider: %s", provider)
	}

	if level == types.INFO {
		log.Printf("[INFO] %s", message)
		types.DebugLog(l.config, "INFO level message logged locally for custom provider, skipping send")
		return nil
	}

	resolvedChannel := channel
	if resolvedChannel == "" {
		resolvedChannel = l.resolveChannel(level)
		types.DebugLog(l.config, "Resolved channel for custom send: %s", resolvedChannel)
	}

	sendConfig := l.config
	sendConfig.Channel = resolvedChannel

	if trace != "" {
		types.DebugLog(l.config, "Processing trace for custom send, trace length: %d", len(trace))
		traceAttachment := &types.Attachment{
			FileName: "trace.log",
			Content:  trace,
		}
		if attachment != nil {
			if attachment.Content != "" {
				attachment.Content += "\n\n--- Trace Log ---\n" + trace
			} else {
				attachment.Content = trace
				attachment.FileName = "trace.log"
			}
		} else {
			attachment = traceAttachment
		}
	}

	types.DebugLog(l.config, "Calling custom provider.SendToChannel with provider: %s, channel: %s", provider, resolvedChannel)
	err := customProvider.SendToChannel(level, message, attachment, sendConfig, resolvedChannel)
	if err != nil {
		types.DebugLog(l.config, "Custom provider.SendToChannel failed: %v", err)
	} else {
		types.DebugLog(l.config, "Custom provider.SendToChannel completed successfully")
	}
	return err
}
