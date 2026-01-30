package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alvianhanif/gocommonlog/types"
)

// SlackProvider implements Provider for Slack
type SlackProvider struct{}

func (p *SlackProvider) Send(level int, message string, attachment *types.Attachment, cfg types.Config) error {
	return p.SendToChannel(level, message, attachment, cfg, cfg.Channel)
}

func (p *SlackProvider) SendToChannel(level int, message string, attachment *types.Attachment, cfg types.Config, channel string) error {
	types.DebugLog(cfg, "SlackProvider.SendToChannel called with level: %d, send method: %s, channel: %s",
		level, cfg.SendMethod, channel)

	cfgCopy := cfg
	cfgCopy.Channel = channel
	switch cfgCopy.SendMethod {
	case types.MethodWebClient:
		types.DebugLog(cfg, "Using Slack webclient method")
		return p.sendSlackWebClient(message, attachment, cfgCopy)
	case types.MethodWebhook:
		types.DebugLog(cfg, "Using Slack webhook method")
		return p.sendSlackWebhook(message, attachment, cfgCopy)
	default:
		err := fmt.Errorf("unknown send method for Slack: %s", cfgCopy.SendMethod)
		types.DebugLog(cfg, "Error: %v", err)
		return err
	}
}

// formatMessage formats the alert message with optional attachment
func (p *SlackProvider) formatMessage(message string, attachment *types.Attachment, cfg types.Config) string {
	formatted := ""

	// Add service and environment header
	if cfg.ServiceName != "" && cfg.Environment != "" {
		formatted += fmt.Sprintf("*[%s - %s]*\n", cfg.ServiceName, cfg.Environment)
	} else if cfg.ServiceName != "" {
		formatted += fmt.Sprintf("*[%s]*\n", cfg.ServiceName)
	} else if cfg.Environment != "" {
		formatted += fmt.Sprintf("*[%s]*\n", cfg.Environment)
	}

	formatted += message

	if attachment != nil {
		if attachment.Content != "" {
			// Inline content - show as expandable code block
			filename := attachment.FileName
			if filename == "" {
				filename = "Trace Logs"
			}
			formatted += fmt.Sprintf("\n\n*%s:*\n```\n%s\n```", filename, attachment.Content)
		}
		if attachment.URL != "" {
			// External URL attachment
			formatted += fmt.Sprintf("\n\n*Attachment:* %s", attachment.URL)
		}
	}

	return formatted
}

func (p *SlackProvider) sendSlackWebhook(message string, attachment *types.Attachment, cfg types.Config) error {
	types.DebugLog(cfg, "sendSlackWebhook: formatting message and preparing webhook request")
	formattedMessage := p.formatMessage(message, attachment, cfg)

	// For webhook, the token field contains the webhook URL
	webhookURL := cfg.ProviderConfig["token"].(string)
	if webhookURL == "" {
		err := fmt.Errorf("webhook URL is required for Slack webhook method")
		types.DebugLog(cfg, "Error: %v", err)
		return err
	}
	types.DebugLog(cfg, "sendSlackWebhook: using webhook URL (length: %d), channel: %s", len(webhookURL), cfg.Channel)

	payload := map[string]interface{}{
		"text": formattedMessage,
	}
	// If channel is specified, include it in the payload
	if cfg.Channel != "" {
		payload["channel"] = cfg.Channel
	}

	data, _ := json.Marshal(payload)
	types.DebugLog(cfg, "sendSlackWebhook: payload prepared, size: %d bytes", len(data))

	req, _ := http.NewRequest("POST", webhookURL, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	types.DebugLog(cfg, "sendSlackWebhook: sending HTTP request to webhook URL")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		types.DebugLog(cfg, "sendSlackWebhook: HTTP request failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	// Log response data
	respData := new(bytes.Buffer)
	respData.ReadFrom(resp.Body)
	types.DebugLog(cfg, "sendSlackWebhook: response status: %d, body length: %d, body: %s", resp.StatusCode, respData.Len(), respData.String())

	if resp.StatusCode != 200 {
		err := fmt.Errorf("slack webhook response: %d", resp.StatusCode)
		types.DebugLog(cfg, "sendSlackWebhook: error response: %v", err)
		return err
	}
	types.DebugLog(cfg, "sendSlackWebhook: webhook sent successfully")
	return nil
}

func (p *SlackProvider) sendSlackWebClient(message string, attachment *types.Attachment, cfg types.Config) error {
	types.DebugLog(cfg, "sendSlackWebClient: formatting message and preparing API request")
	formattedMessage := p.formatMessage(message, attachment, cfg)

	// Use SlackToken if available, otherwise fall back to Token
	token := cfg.ProviderConfig["token"].(string)
	if slackToken, ok := cfg.ProviderConfig["slack_token"].(string); ok && slackToken != "" {
		token = slackToken
		types.DebugLog(cfg, "sendSlackWebClient: using SlackToken (length: %d)", len(token))
	} else {
		types.DebugLog(cfg, "sendSlackWebClient: using Token (length: %d)", len(token))
	}

	url := "https://slack.com/api/chat.postMessage"
	headers := map[string]string{"Authorization": "Bearer " + token, "Content-Type": "application/json; charset=utf-8"}
	payload := map[string]interface{}{
		"channel": cfg.Channel,
		"text":    formattedMessage,
	}
	data, _ := json.Marshal(payload)
	types.DebugLog(cfg, "sendSlackWebClient: sending to channel: %s, payload size: %d bytes", cfg.Channel, len(data))

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(data))
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	types.DebugLog(cfg, "sendSlackWebClient: sending HTTP request to Slack API")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		types.DebugLog(cfg, "sendSlackWebClient: HTTP request failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	// Log response data
	respData := new(bytes.Buffer)
	respData.ReadFrom(resp.Body)
	types.DebugLog(cfg, "sendSlackWebClient: response status: %d, body length: %d, body: %s", resp.StatusCode, respData.Len(), respData.String())

	if resp.StatusCode != 200 {
		err := fmt.Errorf("slack WebClient response: %d", resp.StatusCode)
		types.DebugLog(cfg, "sendSlackWebClient: error response: %v", err)
		return err
	}
	types.DebugLog(cfg, "sendSlackWebClient: message sent successfully")
	return nil
}
