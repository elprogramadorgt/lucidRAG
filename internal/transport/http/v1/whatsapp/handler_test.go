package whatsapp

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	conversationDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/conversation"
	documentDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/document"
	whatsappDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/whatsapp"
	"github.com/elprogramadorgt/lucidRAG/internal/transport/http/v1/whatsapp/dto"
	"github.com/elprogramadorgt/lucidRAG/pkg/logger"
	"github.com/gin-gonic/gin"
)

type mockWhatsAppService struct {
	verifyWebhookFunc func(req whatsappDomain.HookInput, expectedToken string) (string, error)
}

func (m *mockWhatsAppService) VerifyWebhook(req whatsappDomain.HookInput, expectedToken string) (string, error) {
	if m.verifyWebhookFunc != nil {
		return m.verifyWebhookFunc(req, expectedToken)
	}
	return req.Challenge, nil
}

type mockConversationService struct {
	saveIncomingMessageFunc func(ctx context.Context, phoneNumber, contactName, whatsappMsgID, content, msgType string) (*conversationDomain.Message, error)
	saveOutgoingMessageFunc func(ctx context.Context, conversationID, content, ragAnswer string) (*conversationDomain.Message, error)
}

func (m *mockConversationService) GetOrCreateConversation(ctx context.Context, userID, phoneNumber, contactName string) (*conversationDomain.Conversation, error) {
	return nil, nil
}

func (m *mockConversationService) ListConversations(ctx context.Context, userCtx conversationDomain.UserContext, limit, offset int) ([]conversationDomain.Conversation, int64, error) {
	return nil, 0, nil
}

func (m *mockConversationService) GetConversation(ctx context.Context, userCtx conversationDomain.UserContext, id string) (*conversationDomain.Conversation, error) {
	return nil, nil
}

func (m *mockConversationService) SaveIncomingMessage(ctx context.Context, phoneNumber, contactName, whatsappMsgID, content, msgType string) (*conversationDomain.Message, error) {
	if m.saveIncomingMessageFunc != nil {
		return m.saveIncomingMessageFunc(ctx, phoneNumber, contactName, whatsappMsgID, content, msgType)
	}
	return &conversationDomain.Message{ID: "msg-123", ConversationID: "conv-123"}, nil
}

func (m *mockConversationService) SaveOutgoingMessage(ctx context.Context, conversationID, content, ragAnswer string) (*conversationDomain.Message, error) {
	if m.saveOutgoingMessageFunc != nil {
		return m.saveOutgoingMessageFunc(ctx, conversationID, content, ragAnswer)
	}
	return &conversationDomain.Message{ID: "msg-456"}, nil
}

func (m *mockConversationService) GetMessages(ctx context.Context, userCtx conversationDomain.UserContext, conversationID string, limit, offset int) ([]conversationDomain.Message, int64, error) {
	return nil, 0, nil
}

type mockDocumentService struct {
	queryRAGFunc func(ctx context.Context, query documentDomain.RAGQuery) (*documentDomain.RAGResponse, error)
}

func (m *mockDocumentService) CreateDocument(ctx context.Context, userCtx documentDomain.UserContext, doc *documentDomain.Document) (string, error) {
	return "", nil
}

func (m *mockDocumentService) GetDocument(ctx context.Context, userCtx documentDomain.UserContext, id string) (*documentDomain.Document, error) {
	return nil, nil
}

func (m *mockDocumentService) ListDocuments(ctx context.Context, userCtx documentDomain.UserContext, limit, offset int) ([]documentDomain.Document, int64, error) {
	return nil, 0, nil
}

func (m *mockDocumentService) UpdateDocument(ctx context.Context, userCtx documentDomain.UserContext, doc *documentDomain.Document) error {
	return nil
}

func (m *mockDocumentService) DeleteDocument(ctx context.Context, userCtx documentDomain.UserContext, id string) error {
	return nil
}

func (m *mockDocumentService) QueryRAG(ctx context.Context, query documentDomain.RAGQuery) (*documentDomain.RAGResponse, error) {
	if m.queryRAGFunc != nil {
		return m.queryRAGFunc(ctx, query)
	}
	return &documentDomain.RAGResponse{Answer: "test answer"}, nil
}

func createTestLogger() *logger.Logger {
	return logger.New(logger.Options{Level: "error"})
}

func setupTestHandler(whatsappSvc whatsappDomain.Service, convSvc conversationDomain.Service, docSvc documentDomain.Service) *Handler {
	return NewHandler(HandlerConfig{
		WhatsAppSvc:        whatsappSvc,
		ConversationSvc:    convSvc,
		DocumentSvc:        docSvc,
		WebhookVerifyToken: "test-token",
		Log:                createTestLogger(),
	})
}

func setupTestRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/webhook", handler.HandleWebhookVerification)
	r.POST("/webhook", handler.HandleIncomingMessage)
	return r
}

func TestNewHandler(t *testing.T) {
	whatsappSvc := &mockWhatsAppService{}
	convSvc := &mockConversationService{}
	docSvc := &mockDocumentService{}

	handler := setupTestHandler(whatsappSvc, convSvc, docSvc)

	if handler == nil {
		t.Fatal("NewHandler returned nil")
	}
	if handler.svc == nil {
		t.Error("Handler whatsapp service is nil")
	}
	if handler.convSvc == nil {
		t.Error("Handler conversation service is nil")
	}
	if handler.docSvc == nil {
		t.Error("Handler document service is nil")
	}
	if handler.webhookVerifyToken != "test-token" {
		t.Errorf("Expected webhook token 'test-token', got %q", handler.webhookVerifyToken)
	}
}

func TestHandleWebhookVerification_Success(t *testing.T) {
	whatsappSvc := &mockWhatsAppService{
		verifyWebhookFunc: func(req whatsappDomain.HookInput, expectedToken string) (string, error) {
			if req.Mode != "subscribe" {
				t.Errorf("Expected mode 'subscribe', got %q", req.Mode)
			}
			if expectedToken != "test-token" {
				t.Errorf("Expected token 'test-token', got %q", expectedToken)
			}
			return req.Challenge, nil
		},
	}

	handler := setupTestHandler(whatsappSvc, nil, nil)
	router := setupTestRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/webhook?hub.mode=subscribe&hub.verify_token=test-token&hub.challenge=challenge123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response dto.HookVerificationResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	if response.Challenge != "challenge123" {
		t.Errorf("Expected challenge 'challenge123', got %q", response.Challenge)
	}
}

func TestHandleWebhookVerification_MissingParams(t *testing.T) {
	handler := setupTestHandler(&mockWhatsAppService{}, nil, nil)
	router := setupTestRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/webhook?hub.mode=subscribe", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleWebhookVerification_InvalidToken(t *testing.T) {
	whatsappSvc := &mockWhatsAppService{
		verifyWebhookFunc: func(req whatsappDomain.HookInput, expectedToken string) (string, error) {
			return "", errors.New("invalid token")
		},
	}

	handler := setupTestHandler(whatsappSvc, nil, nil)
	router := setupTestRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/webhook?hub.mode=subscribe&hub.verify_token=wrong-token&hub.challenge=challenge123", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestHandleIncomingMessage_Success(t *testing.T) {
	convSvc := &mockConversationService{}
	docSvc := &mockDocumentService{
		queryRAGFunc: func(ctx context.Context, query documentDomain.RAGQuery) (*documentDomain.RAGResponse, error) {
			return &documentDomain.RAGResponse{
				Answer:          "Test answer",
				ConfidenceScore: 0.9,
			}, nil
		},
	}

	handler := setupTestHandler(&mockWhatsAppService{}, convSvc, docSvc)
	router := setupTestRouter(handler)

	payload := dto.WebhookPayload{
		Object: "whatsapp_business_account",
		Entry: []dto.Entry{
			{
				ID: "entry1",
				Changes: []dto.Change{
					{
						Value: dto.ChangeValue{
							MessagingProduct: "whatsapp",
							Contacts: []dto.Contact{
								{WaID: "1234567890", Profile: dto.Profile{Name: "Test User"}},
							},
							Messages: []dto.Message{
								{
									From: "1234567890",
									ID:   "msg1",
									Type: "text",
									Text: &dto.TextMessage{Body: "Hello"},
								},
							},
						},
					},
				},
			},
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	if response["status"] != "received" {
		t.Errorf("Expected status 'received', got %q", response["status"])
	}
}

func TestHandleIncomingMessage_InvalidJSON(t *testing.T) {
	handler := setupTestHandler(&mockWhatsAppService{}, nil, nil)
	router := setupTestRouter(handler)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleIncomingMessage_NonWhatsAppObject(t *testing.T) {
	handler := setupTestHandler(&mockWhatsAppService{}, nil, nil)
	router := setupTestRouter(handler)

	payload := dto.WebhookPayload{
		Object: "other_type",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	if response["status"] != "ignored" {
		t.Errorf("Expected status 'ignored', got %q", response["status"])
	}
}

func TestHandleIncomingMessage_NonTextMessage(t *testing.T) {
	convSvc := &mockConversationService{
		saveIncomingMessageFunc: func(ctx context.Context, phoneNumber, contactName, whatsappMsgID, content, msgType string) (*conversationDomain.Message, error) {
			t.Error("SaveIncomingMessage should not be called for non-text messages")
			return nil, nil
		},
	}

	handler := setupTestHandler(&mockWhatsAppService{}, convSvc, &mockDocumentService{})
	router := setupTestRouter(handler)

	payload := dto.WebhookPayload{
		Object: "whatsapp_business_account",
		Entry: []dto.Entry{
			{
				Changes: []dto.Change{
					{
						Value: dto.ChangeValue{
							Messages: []dto.Message{
								{
									From: "1234567890",
									ID:   "msg1",
									Type: "image",
								},
							},
						},
					},
				},
			},
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandleIncomingMessage_NoConversationService(t *testing.T) {
	handler := setupTestHandler(&mockWhatsAppService{}, nil, nil)
	router := setupTestRouter(handler)

	payload := dto.WebhookPayload{
		Object: "whatsapp_business_account",
		Entry: []dto.Entry{
			{
				Changes: []dto.Change{
					{
						Value: dto.ChangeValue{
							Messages: []dto.Message{
								{
									From: "1234567890",
									ID:   "msg1",
									Type: "text",
									Text: &dto.TextMessage{Body: "Hello"},
								},
							},
						},
					},
				},
			},
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandleIncomingMessage_SaveMessageError(t *testing.T) {
	convSvc := &mockConversationService{
		saveIncomingMessageFunc: func(ctx context.Context, phoneNumber, contactName, whatsappMsgID, content, msgType string) (*conversationDomain.Message, error) {
			return nil, errors.New("database error")
		},
	}

	handler := setupTestHandler(&mockWhatsAppService{}, convSvc, &mockDocumentService{})
	router := setupTestRouter(handler)

	payload := dto.WebhookPayload{
		Object: "whatsapp_business_account",
		Entry: []dto.Entry{
			{
				Changes: []dto.Change{
					{
						Value: dto.ChangeValue{
							Messages: []dto.Message{
								{
									From: "1234567890",
									ID:   "msg1",
									Type: "text",
									Text: &dto.TextMessage{Body: "Hello"},
								},
							},
						},
					},
				},
			},
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d (errors are logged, not returned)", http.StatusOK, w.Code)
	}
}

func TestHandleIncomingMessage_NoDocumentService(t *testing.T) {
	convSvc := &mockConversationService{}
	handler := setupTestHandler(&mockWhatsAppService{}, convSvc, nil)
	router := setupTestRouter(handler)

	payload := dto.WebhookPayload{
		Object: "whatsapp_business_account",
		Entry: []dto.Entry{
			{
				Changes: []dto.Change{
					{
						Value: dto.ChangeValue{
							Messages: []dto.Message{
								{
									From: "1234567890",
									ID:   "msg1",
									Type: "text",
									Text: &dto.TextMessage{Body: "Hello"},
								},
							},
						},
					},
				},
			},
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandleIncomingMessage_RAGError(t *testing.T) {
	convSvc := &mockConversationService{}
	docSvc := &mockDocumentService{
		queryRAGFunc: func(ctx context.Context, query documentDomain.RAGQuery) (*documentDomain.RAGResponse, error) {
			return nil, errors.New("RAG error")
		},
	}

	handler := setupTestHandler(&mockWhatsAppService{}, convSvc, docSvc)
	router := setupTestRouter(handler)

	payload := dto.WebhookPayload{
		Object: "whatsapp_business_account",
		Entry: []dto.Entry{
			{
				Changes: []dto.Change{
					{
						Value: dto.ChangeValue{
							Messages: []dto.Message{
								{
									From: "1234567890",
									ID:   "msg1",
									Type: "text",
									Text: &dto.TextMessage{Body: "Hello"},
								},
							},
						},
					},
				},
			},
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandleIncomingMessage_SaveOutgoingError(t *testing.T) {
	convSvc := &mockConversationService{
		saveOutgoingMessageFunc: func(ctx context.Context, conversationID, content, ragAnswer string) (*conversationDomain.Message, error) {
			return nil, errors.New("save error")
		},
	}
	docSvc := &mockDocumentService{}

	handler := setupTestHandler(&mockWhatsAppService{}, convSvc, docSvc)
	router := setupTestRouter(handler)

	payload := dto.WebhookPayload{
		Object: "whatsapp_business_account",
		Entry: []dto.Entry{
			{
				Changes: []dto.Change{
					{
						Value: dto.ChangeValue{
							Messages: []dto.Message{
								{
									From: "1234567890",
									ID:   "msg1",
									Type: "text",
									Text: &dto.TextMessage{Body: "Hello"},
								},
							},
						},
					},
				},
			},
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestHandleIncomingMessage_ContactNameExtraction(t *testing.T) {
	var capturedContactName string
	convSvc := &mockConversationService{
		saveIncomingMessageFunc: func(ctx context.Context, phoneNumber, contactName, whatsappMsgID, content, msgType string) (*conversationDomain.Message, error) {
			capturedContactName = contactName
			return &conversationDomain.Message{ID: "msg-123", ConversationID: "conv-123"}, nil
		},
	}

	handler := setupTestHandler(&mockWhatsAppService{}, convSvc, &mockDocumentService{})
	router := setupTestRouter(handler)

	payload := dto.WebhookPayload{
		Object: "whatsapp_business_account",
		Entry: []dto.Entry{
			{
				Changes: []dto.Change{
					{
						Value: dto.ChangeValue{
							Contacts: []dto.Contact{
								{WaID: "1234567890", Profile: dto.Profile{Name: "John Doe"}},
								{WaID: "0987654321", Profile: dto.Profile{Name: "Jane Doe"}},
							},
							Messages: []dto.Message{
								{
									From: "1234567890",
									ID:   "msg1",
									Type: "text",
									Text: &dto.TextMessage{Body: "Hello"},
								},
							},
						},
					},
				},
			},
		},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if capturedContactName != "John Doe" {
		t.Errorf("Expected contact name 'John Doe', got %q", capturedContactName)
	}
}

func TestHandleIncomingMessage_EmptyPayload(t *testing.T) {
	handler := setupTestHandler(&mockWhatsAppService{}, nil, nil)
	router := setupTestRouter(handler)

	payload := dto.WebhookPayload{
		Object: "whatsapp_business_account",
		Entry:  []dto.Entry{},
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}
