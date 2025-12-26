package conversation

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	convDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/conversation"
	"github.com/elprogramadorgt/lucidRAG/pkg/logger"
	"github.com/gin-gonic/gin"
)

type mockConversationService struct {
	listConversationsFunc func(ctx context.Context, userCtx convDomain.UserContext, limit, offset int) ([]convDomain.Conversation, int64, error)
	getConversationFunc   func(ctx context.Context, userCtx convDomain.UserContext, id string) (*convDomain.Conversation, error)
	getMessagesFunc       func(ctx context.Context, userCtx convDomain.UserContext, conversationID string, limit, offset int) ([]convDomain.Message, int64, error)
}

func (m *mockConversationService) ListConversations(ctx context.Context, userCtx convDomain.UserContext, limit, offset int) ([]convDomain.Conversation, int64, error) {
	if m.listConversationsFunc != nil {
		return m.listConversationsFunc(ctx, userCtx, limit, offset)
	}
	return []convDomain.Conversation{}, 0, nil
}

func (m *mockConversationService) GetConversation(ctx context.Context, userCtx convDomain.UserContext, id string) (*convDomain.Conversation, error) {
	if m.getConversationFunc != nil {
		return m.getConversationFunc(ctx, userCtx, id)
	}
	return nil, errors.New("not found")
}

func (m *mockConversationService) GetMessages(ctx context.Context, userCtx convDomain.UserContext, conversationID string, limit, offset int) ([]convDomain.Message, int64, error) {
	if m.getMessagesFunc != nil {
		return m.getMessagesFunc(ctx, userCtx, conversationID, limit, offset)
	}
	return []convDomain.Message{}, 0, nil
}

func (m *mockConversationService) CreateConversation(ctx context.Context, conv *convDomain.Conversation) error {
	return nil
}

func (m *mockConversationService) CreateMessage(ctx context.Context, msg *convDomain.Message) error {
	return nil
}

func (m *mockConversationService) GetConversationByPhone(ctx context.Context, userID, phoneNumber string) (*convDomain.Conversation, error) {
	return nil, nil
}

func (m *mockConversationService) GetOrCreateConversation(ctx context.Context, userID, phoneNumber, contactName string) (*convDomain.Conversation, error) {
	return nil, nil
}

func (m *mockConversationService) SaveIncomingMessage(ctx context.Context, phoneNumber, contactName, whatsappMsgID, content, msgType string) (*convDomain.Message, error) {
	return nil, nil
}

func (m *mockConversationService) SaveOutgoingMessage(ctx context.Context, conversationID, content, ragAnswer string) (*convDomain.Message, error) {
	return nil, nil
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func createTestHandler(mockSvc *mockConversationService) *Handler {
	log := logger.New(logger.Options{Level: "error"})
	return NewHandler(mockSvc, log)
}

func TestListConversations(t *testing.T) {
	mockSvc := &mockConversationService{
		listConversationsFunc: func(ctx context.Context, userCtx convDomain.UserContext, limit, offset int) ([]convDomain.Conversation, int64, error) {
			return []convDomain.Conversation{
				{ID: "conv-1", PhoneNumber: "+1234567890"},
				{ID: "conv-2", PhoneNumber: "+0987654321"},
			}, 2, nil
		},
	}
	handler := createTestHandler(mockSvc)

	router := setupTestRouter()
	router.GET("/conversations", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Set("user_role", "user")
		handler.ListConversations(c)
	})

	req, _ := http.NewRequest("GET", "/conversations", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	convs, ok := result["conversations"].([]interface{})
	if !ok {
		t.Fatal("Expected conversations array in response")
	}
	if len(convs) != 2 {
		t.Errorf("Expected 2 conversations, got %d", len(convs))
	}
}

func TestListConversationsWithPagination(t *testing.T) {
	var capturedLimit, capturedOffset int
	mockSvc := &mockConversationService{
		listConversationsFunc: func(ctx context.Context, userCtx convDomain.UserContext, limit, offset int) ([]convDomain.Conversation, int64, error) {
			capturedLimit = limit
			capturedOffset = offset
			return []convDomain.Conversation{}, 0, nil
		},
	}
	handler := createTestHandler(mockSvc)

	router := setupTestRouter()
	router.GET("/conversations", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Set("user_role", "user")
		handler.ListConversations(c)
	})

	req, _ := http.NewRequest("GET", "/conversations?limit=10&offset=5", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if capturedLimit != 10 {
		t.Errorf("Expected limit 10, got %d", capturedLimit)
	}
	if capturedOffset != 5 {
		t.Errorf("Expected offset 5, got %d", capturedOffset)
	}
}

func TestListConversationsError(t *testing.T) {
	mockSvc := &mockConversationService{
		listConversationsFunc: func(ctx context.Context, userCtx convDomain.UserContext, limit, offset int) ([]convDomain.Conversation, int64, error) {
			return nil, 0, errors.New("database error")
		},
	}
	handler := createTestHandler(mockSvc)

	router := setupTestRouter()
	router.GET("/conversations", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Set("user_role", "user")
		handler.ListConversations(c)
	})

	req, _ := http.NewRequest("GET", "/conversations", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", resp.Code)
	}
}

func TestGetConversation(t *testing.T) {
	mockSvc := &mockConversationService{
		getConversationFunc: func(ctx context.Context, userCtx convDomain.UserContext, id string) (*convDomain.Conversation, error) {
			return &convDomain.Conversation{
				ID:          id,
				PhoneNumber: "+1234567890",
				ContactName: "John Doe",
			}, nil
		},
	}
	handler := createTestHandler(mockSvc)

	router := setupTestRouter()
	router.GET("/conversations/:id", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Set("user_role", "user")
		handler.GetConversation(c)
	})

	req, _ := http.NewRequest("GET", "/conversations/conv-123", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.Code)
	}
}

func TestGetConversationMissingID(t *testing.T) {
	mockSvc := &mockConversationService{}
	handler := createTestHandler(mockSvc)

	// Test handler directly with empty ID in context
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Params = gin.Params{{Key: "id", Value: ""}}
	ctx.Set("user_id", "user-123")
	ctx.Set("user_role", "user")
	ctx.Request, _ = http.NewRequest("GET", "/conversations/", nil)

	handler.GetConversation(ctx)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestGetMessages(t *testing.T) {
	mockSvc := &mockConversationService{
		getMessagesFunc: func(ctx context.Context, userCtx convDomain.UserContext, conversationID string, limit, offset int) ([]convDomain.Message, int64, error) {
			return []convDomain.Message{
				{ID: "msg-1", Content: "Hello"},
				{ID: "msg-2", Content: "World"},
			}, 2, nil
		},
	}
	handler := createTestHandler(mockSvc)

	router := setupTestRouter()
	router.GET("/conversations/:id/messages", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Set("user_role", "user")
		handler.GetMessages(c)
	})

	req, _ := http.NewRequest("GET", "/conversations/conv-123/messages", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	msgs, ok := result["messages"].([]interface{})
	if !ok {
		t.Fatal("Expected messages array in response")
	}
	if len(msgs) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(msgs))
	}
}

func TestGetUserContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Set("user_id", "user-123")
	ctx.Set("user_role", "admin")

	userCtx := getUserContext(ctx)

	if userCtx.UserID != "user-123" {
		t.Errorf("Expected UserID 'user-123', got '%s'", userCtx.UserID)
	}
	if !userCtx.IsAdmin {
		t.Error("Expected IsAdmin to be true for admin role")
	}
}

func TestGetUserContextNonAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Set("user_id", "user-123")
	ctx.Set("user_role", "user")

	userCtx := getUserContext(ctx)

	if userCtx.IsAdmin {
		t.Error("Expected IsAdmin to be false for user role")
	}
}
