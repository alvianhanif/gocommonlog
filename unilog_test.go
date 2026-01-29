package commonlog

import (
	"testing"

	"github.com/alvianhanif/commonlog/go/types"
)

func TestNewLogger(t *testing.T) {
	cfg := types.Config{
		Provider:   "slack",
		SendMethod: types.MethodWebClient,
		Token:      "dummy-token",
		Channel:    "#test",
	}
	logger := NewLogger(cfg)
	if logger.config.Provider != "slack" {
		t.Errorf("Expected provider %s, got %s", "slack", logger.config.Provider)
	}
}

func TestNewLoggerWithLark(t *testing.T) {
	cfg := types.Config{
		Provider:   "lark",
		SendMethod: types.MethodWebhook,
		LarkToken:  types.LarkTokenConfig{AppID: "test", AppSecret: "secret"},
		Channel:    "test-channel",
	}
	logger := NewLogger(cfg)
	if logger.config.Provider != "lark" {
		t.Errorf("Expected provider %s, got %s", "lark", logger.config.Provider)
	}
}

func TestNewLoggerWithUnknownProvider(t *testing.T) {
	cfg := types.Config{
		Provider:   "unknown",
		SendMethod: types.MethodWebClient,
		Token:      "dummy-token",
		Channel:    "#test",
	}
	logger := NewLogger(cfg)
	if logger.config.Provider != "unknown" {
		t.Errorf("Expected provider %s, got %s", "unknown", logger.config.Provider)
	}
	// Should default to slack provider
	if logger.provider == nil {
		t.Error("Expected provider to be initialized")
	}
}

func TestSendInfo(t *testing.T) {
	cfg := types.Config{}
	logger := NewLogger(cfg)
	// INFO level should not send, just log
	if err := logger.Send(types.INFO, "Test info message", nil, ""); err != nil {
		t.Errorf("Expected no error for INFO level, got %v", err)
	}
}

func TestSendWarn(t *testing.T) {
	cfg := types.Config{
		Provider:   "slack",
		SendMethod: types.MethodWebhook, // Use webhook which should definitely fail with dummy URL
		Token:      "dummy-token",
		Channel:    "#test",
	}
	logger := NewLogger(cfg)
	// WARN level should attempt to send (will fail with dummy token)
	err := logger.Send(types.WARN, "Test warn message", nil, "")
	if err == nil {
		t.Error("Expected error with dummy webhook URL, but got none")
	}
}

func TestSendError(t *testing.T) {
	cfg := types.Config{
		Provider:   "slack",
		SendMethod: types.MethodWebhook, // Use webhook which should definitely fail
		Token:      "dummy-token",
		Channel:    "#test",
	}
	logger := NewLogger(cfg)
	// ERROR level should attempt to send (will fail with dummy token)
	err := logger.Send(types.ERROR, "Test error message", nil, "")
	if err == nil {
		t.Error("Expected error with dummy webhook URL, but got none")
	}
}

func TestSendToChannel(t *testing.T) {
	cfg := types.Config{
		Provider:   "slack",
		SendMethod: types.MethodWebhook,
		Token:      "dummy-token",
		Channel:    "#default",
	}
	logger := NewLogger(cfg)
	// Test sending to specific channel
	err := logger.SendToChannel(types.WARN, "Test message", nil, "", "#custom")
	if err == nil {
		t.Error("Expected error with dummy webhook URL, but got none")
	}
}

func TestSendWithAttachment(t *testing.T) {
	cfg := types.Config{
		Provider:   "slack",
		SendMethod: types.MethodWebhook,
		Token:      "dummy-token",
		Channel:    "#test",
	}
	logger := NewLogger(cfg)
	attachment := &types.Attachment{
		FileName: "test.txt",
		Content:  "test content",
	}
	err := logger.Send(types.ERROR, "Test with attachment", attachment, "")
	if err == nil {
		t.Error("Expected error with dummy webhook URL, but got none")
	}
}

func TestSendWithTrace(t *testing.T) {
	cfg := types.Config{
		Provider:   "slack",
		SendMethod: types.MethodWebhook,
		Token:      "dummy-token",
		Channel:    "#test",
	}
	logger := NewLogger(cfg)
	trace := "stack trace here"
	err := logger.Send(types.ERROR, "Test with trace", nil, trace)
	if err == nil {
		t.Error("Expected error with dummy webhook URL, but got none")
	}
}

func TestSendWithAttachmentAndTrace(t *testing.T) {
	cfg := types.Config{
		Provider:   "slack",
		SendMethod: types.MethodWebhook,
		Token:      "dummy-token",
		Channel:    "#test",
	}
	logger := NewLogger(cfg)
	attachment := &types.Attachment{
		FileName: "test.txt",
		Content:  "test content",
	}
	trace := "stack trace here"
	err := logger.Send(types.ERROR, "Test with attachment and trace", attachment, trace)
	if err == nil {
		t.Error("Expected error with dummy webhook URL, but got none")
	}
}

func TestCustomSend(t *testing.T) {
	cfg := types.Config{
		Provider:   "slack",
		SendMethod: types.MethodWebClient,
		Token:      "dummy-token",
		Channel:    "#test",
	}
	logger := NewLogger(cfg)
	err := logger.CustomSend("lark", types.ERROR, "Custom provider test", nil, "", "")
	if err == nil {
		t.Error("Expected error with dummy config, but got none")
	}
}

func TestCustomSendUnknownProvider(t *testing.T) {
	cfg := types.Config{
		Provider:   "slack",
		SendMethod: types.MethodWebhook,
		Token:      "dummy-token",
		Channel:    "#test",
	}
	logger := NewLogger(cfg)
	err := logger.CustomSend("unknown", types.ERROR, "Unknown provider test", nil, "", "")
	if err == nil {
		t.Error("Expected error with unknown provider and dummy webhook URL, but got none")
	}
}

func TestResolveChannelWithResolver(t *testing.T) {
	resolver := &types.DefaultChannelResolver{
		ChannelMap:     map[int]string{types.ERROR: "#errors", types.WARN: "#warnings"},
		DefaultChannel: "#general",
	}
	cfg := types.Config{
		Provider:        "slack",
		SendMethod:      types.MethodWebClient,
		Token:           "dummy-token",
		Channel:         "#default",
		ChannelResolver: resolver,
	}
	logger := NewLogger(cfg)

	// Test ERROR level resolution
	errorChannel := logger.resolveChannel(types.ERROR)
	if errorChannel != "#errors" {
		t.Errorf("Expected #errors, got %s", errorChannel)
	}

	// Test WARN level resolution
	warnChannel := logger.resolveChannel(types.WARN)
	if warnChannel != "#warnings" {
		t.Errorf("Expected #warnings, got %s", warnChannel)
	}

	// Test INFO level resolution (should use default)
	infoChannel := logger.resolveChannel(types.INFO)
	if infoChannel != "#general" {
		t.Errorf("Expected #general, got %s", infoChannel)
	}
}

func TestResolveChannelWithoutResolver(t *testing.T) {
	cfg := types.Config{
		Provider:   "slack",
		SendMethod: types.MethodWebClient,
		Token:      "dummy-token",
		Channel:    "#default",
	}
	logger := NewLogger(cfg)

	channel := logger.resolveChannel(types.ERROR)
	if channel != "#default" {
		t.Errorf("Expected #default, got %s", channel)
	}
}
