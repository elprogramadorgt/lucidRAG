package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/elprogramadorgt/lucidRAG/internal/config"
	"github.com/elprogramadorgt/lucidRAG/pkg/logger"
)

// Client implements WhatsApp Cloud API operations
type Client struct {
	config *config.WhatsAppConfig
	logger *logger.Logger
	client *http.Client
}

// NewClient creates a new WhatsApp client
func NewClient(cfg *config.WhatsAppConfig, log *logger.Logger) *Client {
	return &Client{
		config: cfg,
		logger: log,
		client: &http.Client{},
	}
}

// SendMessage sends a text message via WhatsApp
func (c *Client) SendMessage(ctx context.Context, to, content string) error {
	url := fmt.Sprintf("https://graph.facebook.com/%s/%s/messages",
		c.config.APIVersion, c.config.PhoneNumberID)

	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "text",
		"text": map[string]string{
			"body": content,
		},
	}

	return c.sendRequest(ctx, url, payload)
}

// SendTemplateMessage sends a template message via WhatsApp
func (c *Client) SendTemplateMessage(ctx context.Context, to, templateName string, params map[string]string) error {
	url := fmt.Sprintf("https://graph.facebook.com/%s/%s/messages",
		c.config.APIVersion, c.config.PhoneNumberID)

	// Build template components based on params
	components := make([]map[string]interface{}, 0)
	if len(params) > 0 {
		parameters := make([]map[string]string, 0)
		for _, value := range params {
			parameters = append(parameters, map[string]string{
				"type": "text",
				"text": value,
			})
		}
		components = append(components, map[string]interface{}{
			"type":       "body",
			"parameters": parameters,
		})
	}

	payload := map[string]interface{}{
		"messaging_product": "whatsapp",
		"to":                to,
		"type":              "template",
		"template": map[string]interface{}{
			"name":       templateName,
			"language":   map[string]string{"code": "en"},
			"components": components,
		},
	}

	return c.sendRequest(ctx, url, payload)
}

// VerifyWebhook verifies the webhook subscription
func (c *Client) VerifyWebhook(verifyToken, mode, challenge string) (string, error) {
	if mode == "subscribe" && verifyToken == c.config.WebhookVerifyToken {
		c.logger.Info("Webhook verified successfully")
		return challenge, nil
	}
	c.logger.Warn("Webhook verification failed")
	return "", fmt.Errorf("invalid verification token")
}

// ProcessWebhook processes incoming webhook events
func (c *Client) ProcessWebhook(ctx context.Context, payload []byte) error {
	// Parse webhook payload
	var webhookData map[string]interface{}
	if err := json.Unmarshal(payload, &webhookData); err != nil {
		c.logger.Error("Failed to parse webhook payload: %v", err)
		return fmt.Errorf("invalid webhook payload: %w", err)
	}

	c.logger.Info("Received webhook: %+v", webhookData)

	// TODO: Process different webhook event types
	// This is a placeholder for actual webhook processing logic

	return nil
}

// sendRequest sends an HTTP request to WhatsApp API
func (c *Client) sendRequest(ctx context.Context, url string, payload interface{}) error {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode >= 400 {
		c.logger.Error("WhatsApp API error: %s", string(body))
		return fmt.Errorf("WhatsApp API returned status %d: %s", resp.StatusCode, string(body))
	}

	c.logger.Info("WhatsApp API response: %s", string(body))
	return nil
}
