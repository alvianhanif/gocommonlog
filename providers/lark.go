package providers

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/alvianhanif/gocommonlog/cache"
	"github.com/alvianhanif/gocommonlog/types"

	redis "github.com/go-redis/redis/v8"
)

// getRedisClient returns a Redis client using host/port from cfg, env, or default
func getRedisClient(cfg types.Config) (*redis.Client, error) {
	host, ok := cfg.ProviderConfig["redis_host"].(string)
	if !ok || host == "" {
		return nil, fmt.Errorf("redis_host must be set in provider_config")
	}
	port, ok := cfg.ProviderConfig["redis_port"].(string)
	if !ok || port == "" {
		return nil, fmt.Errorf("redis_port must be set in provider_config")
	}

	// Optional configuration for ElastiCache support
	password, _ := cfg.ProviderConfig["redis_password"].(string)
	ssl, _ := cfg.ProviderConfig["redis_ssl"].(bool)
	clusterMode, _ := cfg.ProviderConfig["redis_cluster_mode"].(bool)
	db := 0
	if dbVal, ok := cfg.ProviderConfig["redis_db"]; ok {
		if dbInt, ok := dbVal.(int); ok {
			db = dbInt
		} else if dbStr, ok := dbVal.(string); ok {
			if parsed, err := strconv.Atoi(dbStr); err == nil {
				db = parsed
			}
		}
	}

	fmt.Printf("[Lark] Initializing Redis client with host: '%s', port: '%s'\n", host, port)

	if clusterMode {
		// For cluster mode, we need to use RedisCluster
		// Note: This requires additional setup and the go-redis/redis/v8 library supports clustering
		return nil, fmt.Errorf("cluster mode not yet implemented for Go version - requires RedisCluster client")
	}

	addr := host + ":" + port
	fmt.Printf("[Lark] Connecting to Redis at address: %s\n", addr)

	options := &redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	}

	// Configure TLS if SSL is enabled
	if ssl {
		options.TLSConfig = &tls.Config{
			InsecureSkipVerify: false, // Set to true only for development
		}
	}

	client := redis.NewClient(options)
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		fmt.Printf("[Lark] Failed to ping Redis at %s: %v\n", addr, err)
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}
	fmt.Printf("[Lark] Successfully connected to Redis at %s\n", addr)
	return client, nil
}

func cacheLarkToken(cfg types.Config, appID, appSecret, token string) error {
	key := "commonlog_lark_token:" + appID + ":" + appSecret
	client, err := getRedisClient(cfg)
	if err != nil {
		// Fallback to in-memory cache
		cache.GetGlobalCache().Set(key, token, 90*time.Minute)
		types.DebugLog(cfg, "Lark token cached in memory")
		return nil
	}
	return client.Set(context.Background(), key, token, 90*time.Minute).Err()
}

func cacheChatID(cfg types.Config, channelName, chatID string) error {
	key := "commonlog_lark_chat_id:" + cfg.Environment + ":" + channelName
	client, err := getRedisClient(cfg)
	if err != nil {
		// Fallback to in-memory cache (30 days expiry)
		cache.GetGlobalCache().Set(key, chatID, 30*24*time.Hour)
		types.DebugLog(cfg, "Lark chat ID cached in memory")
		return nil
	}
	return client.Set(context.Background(), key, chatID, 0).Err() // No expiry
}

func getCachedLarkToken(cfg types.Config, appID, appSecret string) (string, error) {
	key := "commonlog_lark_token:" + appID + ":" + appSecret
	client, err := getRedisClient(cfg)
	if err != nil {
		// Fallback to in-memory cache
		if token, found := cache.GetGlobalCache().Get(key); found {
			types.DebugLog(cfg, "Lark token retrieved from memory")
			return token, nil
		}
		return "", nil // No cached token
	}
	result, err := client.Get(context.Background(), key).Result()
	if err == redis.Nil {
		fmt.Printf("[Lark] No cached token found for key: %s\n", key)
		return "", nil // No cached token
	} else if err != nil {
		fmt.Printf("[Lark] Error retrieving cached token for key %s: %v\n", key, err)
		return "", err
	}
	fmt.Printf("[Lark] Retrieved cached token for key: %s\n", key)
	return result, nil
}

func getCachedChatID(cfg types.Config, channelName string) (string, error) {
	key := "commonlog_lark_chat_id:" + cfg.Environment + ":" + channelName
	client, err := getRedisClient(cfg)
	if err != nil {
		// Fallback to in-memory cache
		if chatID, found := cache.GetGlobalCache().Get(key); found {
			types.DebugLog(cfg, "Lark chat ID retrieved from memory")
			return chatID, nil
		}
		return "", nil // No cached chat ID
	}
	result, err := client.Get(context.Background(), key).Result()
	if err == redis.Nil {
		fmt.Printf("[Lark] No cached chat_id found for channel: %s in environment: %s\n", channelName, cfg.Environment)
		return "", nil // No cached chat_id
	} else if err != nil {
		fmt.Printf("[Lark] Error retrieving cached chat_id for channel %s in environment %s: %v\n", channelName, cfg.Environment, err)
		return "", err
	}
	fmt.Printf("[Lark] Retrieved cached chat_id for channel: %s in environment: %s\n", channelName, cfg.Environment)
	return result, nil
}

// getChatIDFromChannelName fetches the chat_id for a given channel name using pagination
func getChatIDFromChannelName(cfg types.Config, token, channelName string) (string, error) {
	// Try Redis cache first
	cached, err := getCachedChatID(cfg, channelName)
	if err != nil {
		return "", fmt.Errorf("failed to get Redis client: %w", err)
	}
	if cached != "" {
		return cached, nil
	}

	baseURL := "https://open.larksuite.com/open-apis/im/v1/chats"
	headers := map[string]string{"Authorization": "Bearer " + token}

	pageToken := ""
	hasMore := true

	for hasMore {
		url := baseURL + "?page_size=10"
		if pageToken != "" {
			url += "&page_token=" + pageToken
		}

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return "", err
		}
		for k, v := range headers {
			req.Header.Set(k, v)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			return "", fmt.Errorf("lark chats API response: %d", resp.StatusCode)
		}

		var result struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
			Data struct {
				Items []struct {
					ChatID string `json:"chat_id"`
					Name   string `json:"name"`
				} `json:"items"`
				PageToken string `json:"page_token"`
				HasMore   bool   `json:"has_more"`
			} `json:"data"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return "", err
		}

		if result.Code != 0 {
			return "", fmt.Errorf("lark API error: %s", result.Msg)
		}

		// Search for the channel name in the current page
		for _, item := range result.Data.Items {
			if item.Name == channelName {
				// Cache the chat_id without expiry
				if err := cacheChatID(cfg, channelName, item.ChatID); err != nil {
					fmt.Printf("[Lark] Warning: failed to cache chat_id for channel %s: %v\n", channelName, err)
				}
				return item.ChatID, nil
			}
		}

		// Update pagination info
		pageToken = result.Data.PageToken
		hasMore = result.Data.HasMore
	}

	return "", fmt.Errorf("channel '%s' not found", channelName)
}

// LarkProvider implements Provider for Lark
type LarkProvider struct{}

func getTenantAccessToken(cfg types.Config, appID, appSecret string) (string, error) {
	// Try Redis cache first
	cached, err := getCachedLarkToken(cfg, appID, appSecret)
	if err != nil {
		return "", fmt.Errorf("failed to get Redis client: %w", err)
	}
	if cached != "" {
		return cached, nil
	}
	url := "https://open.larksuite.com/open-apis/auth/v3/tenant_access_token/internal"
	payload := map[string]string{"app_id": appID, "app_secret": appSecret}
	data, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var result struct {
		Code   int    `json:"code"`
		Msg    string `json:"msg"`
		Token  string `json:"tenant_access_token"`
		Expire int    `json:"expire"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if result.Code != 0 {
		return "", fmt.Errorf("lark token error: %s", result.Msg)
	}
	// Cache the token for (expire - 10 minutes) - optional
	expireSeconds := result.Expire - 600
	if expireSeconds <= 0 {
		expireSeconds = 60 // fallback to 1 minute if API returns too low
	}
	key := "commonlog_lark_token:" + appID + ":" + appSecret
	client, err := getRedisClient(cfg)
	if err != nil {
		// Redis not configured, skip caching but continue with token
		types.DebugLog(cfg, "Lark token caching disabled - Redis not configured")
	} else {
		err = client.Set(context.Background(), key, result.Token, time.Duration(expireSeconds)*time.Second).Err()
		if err != nil {
			fmt.Printf("[Lark] Warning: failed to cache token: %v\n", err)
			// Don't return error, just log warning and continue
		}
	}
	return result.Token, nil
}

func (p *LarkProvider) Send(level int, message string, attachment *types.Attachment, cfg types.Config) error {
	return p.SendToChannel(level, message, attachment, cfg, cfg.Channel)
}

func (p *LarkProvider) SendToChannel(level int, message string, attachment *types.Attachment, cfg types.Config, channel string) error {
	types.DebugLog(cfg, "LarkProvider.SendToChannel called with level: %d, send method: %s, channel: %s",
		level, cfg.SendMethod, channel)

	cfgCopy := cfg
	cfgCopy.Channel = channel
	switch cfgCopy.SendMethod {
	case types.MethodWebClient:
		types.DebugLog(cfg, "Using Lark webclient method")
		return p.sendLarkWebClient(message, attachment, cfgCopy)
	case types.MethodWebhook:
		types.DebugLog(cfg, "Using Lark webhook method")
		return p.sendLarkWebhook(message, attachment, cfgCopy)
	default:
		err := fmt.Errorf("unknown send method for Lark: %s", cfgCopy.SendMethod)
		types.DebugLog(cfg, "Error: %v", err)
		return err
	}
}

// formatMessage formats the alert message with optional attachment and returns title and content separately
func (p *LarkProvider) formatMessage(message string, attachment *types.Attachment, cfg types.Config) (string, string) {
	// Extract title from service and environment
	title := "Alert"
	if cfg.ServiceName != "" && cfg.Environment != "" {
		title = fmt.Sprintf("%s - %s", cfg.ServiceName, cfg.Environment)
	} else if cfg.ServiceName != "" {
		title = cfg.ServiceName
	} else if cfg.Environment != "" {
		title = cfg.Environment
	}

	// Format message content without the header
	formatted := message

	if attachment != nil {
		if attachment.Content != "" {
			// Inline content - show as expandable code block
			filename := attachment.FileName
			if filename == "" {
				filename = "Trace Logs"
			}
			formatted += fmt.Sprintf("\n\n**%s:**\n```\n%s\n```", filename, attachment.Content)
		}
		if attachment.URL != "" {
			// External URL attachment
			formatted += fmt.Sprintf("\n\n**Attachment:** %s", attachment.URL)
		}
	}

	return title, formatted
}

func (p *LarkProvider) sendLarkWebClient(message string, attachment *types.Attachment, cfg types.Config) error {
	types.DebugLog(cfg, "sendLarkWebClient: formatting message and preparing API request")
	title, formattedMessage := p.formatMessage(message, attachment, cfg)
	token := cfg.Token

	types.DebugLog(cfg, "sendLarkWebClient: sending to channel '%s'", cfg.Channel)

	// Use LarkToken if available, otherwise fall back to Token parsing
	var appID, appSecret string
	if cfg.LarkToken.AppID != "" && cfg.LarkToken.AppSecret != "" {
		appID = cfg.LarkToken.AppID
		appSecret = cfg.LarkToken.AppSecret
		types.DebugLog(cfg, "sendLarkWebClient: fetching tenant access token for appID (length: %d)", len(appID))
		fetched, err := getTenantAccessToken(cfg, appID, appSecret)
		if err != nil {
			types.DebugLog(cfg, "sendLarkWebClient: error fetching tenant access token: %v", err)
			return err
		}
		token = fetched
		types.DebugLog(cfg, "sendLarkWebClient: tenant access token fetched successfully")
	}

	// Get chat_id from channel name
	types.DebugLog(cfg, "sendLarkWebClient: resolving chat_id for channel '%s'", cfg.Channel)
	chatID, err := getChatIDFromChannelName(cfg, token, cfg.Channel)
	if err != nil {
		types.DebugLog(cfg, "sendLarkWebClient: failed to get chat_id for channel '%s': %v", cfg.Channel, err)
		return fmt.Errorf("failed to get chat_id for channel '%s': %v", cfg.Channel, err)
	}
	types.DebugLog(cfg, "sendLarkWebClient: resolved chat_id (length: %d)", len(chatID))

	url := "https://open.larksuite.com/open-apis/im/v1/messages?receive_id_type=chat_id"
	headers := map[string]string{"Authorization": "Bearer " + token, "Content-Type": "application/json"}

	payload := map[string]interface{}{
		"receive_id": chatID,
		"msg_type":   "post",
		"content": map[string]interface{}{
			"post": map[string]interface{}{
				"zh_cn": map[string]interface{}{
					"title": title,
					"content": []interface{}{
						[]interface{}{
							map[string]interface{}{
								"tag":  "text",
								"text": formattedMessage,
							},
						},
					},
				},
			},
		},
	}
	data, _ := json.Marshal(payload)

	types.DebugLog(cfg, "sendLarkWebClient: sending HTTP request to Lark API, payload size: %d bytes, payload: %s", len(data), string(data))
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(data))
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		types.DebugLog(cfg, "sendLarkWebClient: HTTP request failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	// Log response data
	respBody := new(bytes.Buffer)
	_, copyErr := respBody.ReadFrom(resp.Body)
	if copyErr != nil {
		types.DebugLog(cfg, "sendLarkWebClient: error reading response body: %v", copyErr)
	} else {
		types.DebugLog(cfg, "sendLarkWebClient: response status: %d, body length: %d, body: %s", resp.StatusCode, respBody.Len(), respBody.String())
	}

	if resp.StatusCode != 200 {
		err := fmt.Errorf("lark WebClient response: %d", resp.StatusCode)
		types.DebugLog(cfg, "sendLarkWebClient: error response: %v", err)
		return err
	}
	types.DebugLog(cfg, "sendLarkWebClient: message sent successfully to channel '%s'", cfg.Channel)
	return nil
}

func (p *LarkProvider) sendLarkWebhook(message string, attachment *types.Attachment, cfg types.Config) error {
	types.DebugLog(cfg, "sendLarkWebhook: formatting message and preparing webhook request")
	title, formattedMessage := p.formatMessage(message, attachment, cfg)

	// For webhook, the token field contains the webhook URL
	webhookURL := cfg.Token
	if webhookURL == "" {
		err := fmt.Errorf("webhook URL is required for Lark webhook method")
		types.DebugLog(cfg, "Error: %v", err)
		return err
	}
	types.DebugLog(cfg, "sendLarkWebhook: using webhook URL (length: %d)", len(webhookURL))

	payload := map[string]interface{}{
		"msg_type": "post",
		"content": map[string]interface{}{
			"post": map[string]interface{}{
				"zh_cn": map[string]interface{}{
					"title": title,
					"content": []interface{}{
						[]interface{}{
							map[string]interface{}{
								"tag":  "text",
								"text": formattedMessage,
							},
						},
					},
				},
			},
		},
	}

	data, _ := json.Marshal(payload)
	types.DebugLog(cfg, "sendLarkWebhook: payload prepared, size: %d bytes, payload: %s", len(data), string(data))

	req, _ := http.NewRequest("POST", webhookURL, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	types.DebugLog(cfg, "sendLarkWebhook: sending HTTP request to webhook URL")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		types.DebugLog(cfg, "sendLarkWebhook: HTTP request failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	// Log response data
	respBody := new(bytes.Buffer)
	_, copyErr := respBody.ReadFrom(resp.Body)
	if copyErr != nil {
		types.DebugLog(cfg, "sendLarkWebhook: error reading response body: %v", copyErr)
	} else {
		types.DebugLog(cfg, "sendLarkWebhook: response status: %d, body length: %d, body: %s", resp.StatusCode, respBody.Len(), respBody.String())
	}

	if resp.StatusCode != 200 {
		err := fmt.Errorf("lark webhook response: %d", resp.StatusCode)
		types.DebugLog(cfg, "sendLarkWebhook: error response: %v", err)
		return err
	}
	types.DebugLog(cfg, "sendLarkWebhook: webhook sent successfully")
	return nil
}
