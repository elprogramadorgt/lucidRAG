package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/elprogramadorgt/lucidRAG/internal/config"
	"github.com/elprogramadorgt/lucidRAG/internal/domain"
	"github.com/elprogramadorgt/lucidRAG/internal/repository"
	"github.com/elprogramadorgt/lucidRAG/pkg/logger"
)

// Client implements WhatsApp Cloud API operations
type Client struct {
	config      *config.WhatsAppConfig
	logger      *logger.Logger
	client      *http.Client
	messageRepo domain.MessageRepository
	sessionRepo domain.SessionRepository
}

// NewClient creates a new WhatsApp client
func NewClient(cfg *config.WhatsAppConfig, log *logger.Logger, messageRepo domain.MessageRepository, sessionRepo domain.SessionRepository) *Client {
	return &Client{
		config:      cfg,
		logger:      log,
		client:      &http.Client{},
		messageRepo: messageRepo,
		sessionRepo: sessionRepo,
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

	// Process webhook entries
	if entry, ok := webhookData["entry"].([]interface{}); ok && len(entry) > 0 {
		for _, e := range entry {
			if entryMap, ok := e.(map[string]interface{}); ok {
				if changes, ok := entryMap["changes"].([]interface{}); ok {
					for _, change := range changes {
						if changeMap, ok := change.(map[string]interface{}); ok {
							if value, ok := changeMap["value"].(map[string]interface{}); ok {
								c.processMessages(ctx, value)
							}
						}
					}
				}
			}
		}
	}

	return nil
}

// processMessages processes messages from webhook value
func (c *Client) processMessages(ctx context.Context, value map[string]interface{}) {
	messages, ok := value["messages"].([]interface{})
	if !ok || len(messages) == 0 {
		return
	}

	metadata, _ := value["metadata"].(map[string]interface{})
	phoneNumberID := ""
	if metadata != nil {
		if pnid, ok := metadata["phone_number_id"].(string); ok {
			phoneNumberID = pnid
		}
	}

	for _, msg := range messages {
		msgMap, ok := msg.(map[string]interface{})
		if !ok {
			continue
		}

		from, _ := msgMap["from"].(string)
		msgID, _ := msgMap["id"].(string)
		msgType, _ := msgMap["type"].(string)
		timestamp, _ := msgMap["timestamp"].(string)

		// Extract message content based on type
		content := ""
		if msgType == "text" {
			if textObj, ok := msgMap["text"].(map[string]interface{}); ok {
				if body, ok := textObj["body"].(string); ok {
					content = body
				}
			}
		}

		// Parse timestamp
		var msgTime time.Time
		if timestamp != "" {
			if ts, err := strconv.ParseInt(timestamp, 10, 64); err == nil {
				msgTime = time.Unix(ts, 0)
			}
		}
		if msgTime.IsZero() {
			msgTime = time.Now()
		}

		// Create or update session
		session, err := c.sessionRepo.GetByPhoneNumber(ctx, from)
		if err != nil {
			// Session not found, create new one
			session = &domain.ChatSession{
				UserPhoneNumber: from,
				StartedAt:       msgTime,
				LastMessageAt:   msgTime,
				IsActive:        true,
				Context:         "",
			}
			if err := c.sessionRepo.Save(ctx, session); err != nil {
				c.logger.Error("Failed to save session: %v", err)
				continue
			}
		} else {
			// Update last message time
			if err := c.sessionRepo.UpdateLastMessage(ctx, session.ID); err != nil {
				c.logger.Error("Failed to update session: %v", err)
			}
		}

		// Save message with proper session association
		message := &domain.Message{
			ID:          msgID,
			From:        from,
			To:          phoneNumberID,
			Content:     content,
			MessageType: msgType,
			Timestamp:   msgTime,
			Status:      "received",
		}

		if msgRepo, ok := c.messageRepo.(*repository.InMemoryMessageRepository); ok {
			if err := msgRepo.SaveWithSession(ctx, message, session.ID); err != nil {
				c.logger.Error("Failed to save message: %v", err)
			} else {
				c.logger.Info("Message saved: %s from %s", msgID, from)
			}
		} else {
			// Fallback to regular Save method
			if err := c.messageRepo.Save(ctx, message); err != nil {
				c.logger.Error("Failed to save message: %v", err)
			} else {
				c.logger.Info("Message saved: %s from %s", msgID, from)
			}
		}
	}
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

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Warn("Failed to read response body: %v", err)
		body = []byte{}
	}

	if resp.StatusCode >= 400 {
		c.logger.Error("WhatsApp API error: %s", string(body))
		return fmt.Errorf("WhatsApp API returned status %d: %s", resp.StatusCode, string(body))
	}

	c.logger.Info("WhatsApp API response: %s", string(body))
	return nil
}
