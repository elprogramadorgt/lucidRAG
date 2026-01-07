package whatsapp

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	conversationDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/conversation"
	ragDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/rag"
	whatsappDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/whatsapp"
	"github.com/elprogramadorgt/lucidRAG/pkg/logger"
	"github.com/gin-gonic/gin"
)

type mockWhatsAppService struct {
	challenge string
	err       error
}

func (m *mockWhatsAppService) VerifyWebhook(_ whatsappDomain.HookInput, _ string) (string, error) {
	return m.challenge, m.err
}

type mockConversationService struct {
	conversation *conversationDomain.Conversation
	message      *conversationDomain.Message
	messages     []conversationDomain.Message
	count        int64
	err          error
}

func (m *mockConversationService) GetOrCreateConversation(_ context.Context, _, _, _ string) (*conversationDomain.Conversation, error) {
	return m.conversation, m.err
}

func (m *mockConversationService) ListConversations(_ context.Context, _ conversationDomain.UserContext, _, _ int) ([]conversationDomain.Conversation, int64, error) {
	return nil, m.count, m.err
}

func (m *mockConversationService) GetConversation(_ context.Context, _ conversationDomain.UserContext, _ string) (*conversationDomain.Conversation, error) {
	return m.conversation, m.err
}

func (m *mockConversationService) SaveIncomingMessage(_ context.Context, _, _, _, _, _ string) (*conversationDomain.Message, error) {
	return m.message, m.err
}

func (m *mockConversationService) SaveOutgoingMessage(_ context.Context, _, _, _ string) (*conversationDomain.Message, error) {
	return m.message, m.err
}

func (m *mockConversationService) GetMessages(_ context.Context, _ conversationDomain.UserContext, _ string, _, _ int) ([]conversationDomain.Message, int64, error) {
	return m.messages, m.count, m.err
}

type mockRAGService struct {
	response *ragDomain.Response
	err      error
}

func (m *mockRAGService) Query(_ context.Context, _ ragDomain.Query) (*ragDomain.Response, error) {
	return m.response, m.err
}

func (m *mockRAGService) IndexDocument(_ context.Context, _, _ string) error {
	return m.err
}

func (m *mockRAGService) DeleteDocumentChunks(_ context.Context, _ string) error {
	return m.err
}

func setupRouter() (*gin.Engine, *logger.Logger) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	log := logger.New(logger.Options{Level: "error"})
	return r, log
}

func TestNewHandler(t *testing.T) {
	_, log := setupRouter()

	h := NewHandler(HandlerConfig{
		WhatsAppSvc:        &mockWhatsAppService{},
		WebhookVerifyToken: "test-token",
		Log:                log,
	})

	if h == nil {
		t.Fatal("Expected handler to be created")
	}
}

func TestHandleWebhookVerificationSuccess(t *testing.T) {
	r, log := setupRouter()

	waSvc := &mockWhatsAppService{
		challenge: "test-challenge",
	}

	h := NewHandler(HandlerConfig{
		WhatsAppSvc:        waSvc,
		WebhookVerifyToken: "valid-token",
		Log:                log,
	})

	r.GET("/webhook", h.HandleWebhookVerification)

	req := httptest.NewRequest(http.MethodGet, "/webhook?hub.mode=subscribe&hub.verify_token=valid-token&hub.challenge=test-challenge", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleWebhookVerificationMissingParams(t *testing.T) {
	r, log := setupRouter()

	h := NewHandler(HandlerConfig{
		WhatsAppSvc:        &mockWhatsAppService{},
		WebhookVerifyToken: "valid-token",
		Log:                log,
	})

	r.GET("/webhook", h.HandleWebhookVerification)

	req := httptest.NewRequest(http.MethodGet, "/webhook?hub.mode=subscribe", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleWebhookVerificationInvalidToken(t *testing.T) {
	r, log := setupRouter()

	waSvc := &mockWhatsAppService{
		err: errors.New("invalid token"),
	}

	h := NewHandler(HandlerConfig{
		WhatsAppSvc:        waSvc,
		WebhookVerifyToken: "valid-token",
		Log:                log,
	})

	r.GET("/webhook", h.HandleWebhookVerification)

	req := httptest.NewRequest(http.MethodGet, "/webhook?hub.mode=subscribe&hub.verify_token=wrong-token&hub.challenge=test", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestHandleIncomingMessageSuccess(t *testing.T) {
	r, log := setupRouter()

	h := NewHandler(HandlerConfig{
		WhatsAppSvc:        &mockWhatsAppService{},
		WebhookVerifyToken: "test-token",
		Log:                log,
	})

	r.POST("/webhook", h.HandleIncomingMessage)

	body := `{
		"object": "whatsapp_business_account",
		"entry": [{
			"id": "entry-1",
			"changes": [{
				"value": {
					"messaging_product": "whatsapp",
					"contacts": [{"wa_id": "1234567890", "profile": {"name": "John"}}],
					"messages": [{"from": "1234567890", "id": "msg-1", "type": "text", "text": {"body": "Hello"}}]
				}
			}]
		}]
	}`

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandleIncomingMessageInvalidJSON(t *testing.T) {
	r, log := setupRouter()

	h := NewHandler(HandlerConfig{
		WhatsAppSvc:        &mockWhatsAppService{},
		WebhookVerifyToken: "test-token",
		Log:                log,
	})

	r.POST("/webhook", h.HandleIncomingMessage)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleIncomingMessageNonWhatsApp(t *testing.T) {
	r, log := setupRouter()

	h := NewHandler(HandlerConfig{
		WhatsAppSvc:        &mockWhatsAppService{},
		WebhookVerifyToken: "test-token",
		Log:                log,
	})

	r.POST("/webhook", h.HandleIncomingMessage)

	body := `{"object": "other", "entry": []}`
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestHandlerConfigZeroValue(t *testing.T) {
	var cfg HandlerConfig

	if cfg.WhatsAppSvc != nil {
		t.Error("Expected nil WhatsAppSvc in zero value")
	}
	if cfg.ConversationSvc != nil {
		t.Error("Expected nil ConversationSvc in zero value")
	}
	if cfg.RAGSvc != nil {
		t.Error("Expected nil RAGSvc in zero value")
	}
	if cfg.WebhookVerifyToken != "" {
		t.Error("Expected empty WebhookVerifyToken in zero value")
	}
	if cfg.Log != nil {
		t.Error("Expected nil Log in zero value")
	}
}
