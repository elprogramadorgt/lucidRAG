package conversation

import (
	"context"
	"errors"
	"testing"
	"time"

	conversationDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/conversation"
)

// mockConversationRepo is a mock implementation of ConversationRepository
type mockConversationRepo struct {
	conversations map[string]*conversationDomain.Conversation
	phoneIndex    map[string]*conversationDomain.Conversation
}

func newMockConversationRepo() *mockConversationRepo {
	return &mockConversationRepo{
		conversations: make(map[string]*conversationDomain.Conversation),
		phoneIndex:    make(map[string]*conversationDomain.Conversation),
	}
}

func (m *mockConversationRepo) Create(ctx context.Context, conv *conversationDomain.Conversation) (string, error) {
	id := "conv_" + conv.PhoneNumber
	conv.ID = id
	conv.CreatedAt = time.Now()
	conv.UpdatedAt = time.Now()
	m.conversations[id] = conv
	m.phoneIndex[conv.PhoneNumber] = conv
	return id, nil
}

func (m *mockConversationRepo) GetByID(ctx context.Context, id string) (*conversationDomain.Conversation, error) {
	conv, exists := m.conversations[id]
	if !exists {
		return nil, nil
	}
	return conv, nil
}

func (m *mockConversationRepo) GetByPhoneNumber(ctx context.Context, phoneNumber string) (*conversationDomain.Conversation, error) {
	conv, exists := m.phoneIndex[phoneNumber]
	if !exists {
		return nil, nil
	}
	return conv, nil
}

func (m *mockConversationRepo) List(ctx context.Context, limit, offset int) ([]conversationDomain.Conversation, error) {
	convs := make([]conversationDomain.Conversation, 0, len(m.conversations))
	for _, conv := range m.conversations {
		convs = append(convs, *conv)
	}
	return convs, nil
}

func (m *mockConversationRepo) ListByUser(ctx context.Context, userID string, limit, offset int) ([]conversationDomain.Conversation, error) {
	convs := make([]conversationDomain.Conversation, 0)
	for _, conv := range m.conversations {
		if conv.UserID == userID {
			convs = append(convs, *conv)
		}
	}
	return convs, nil
}

func (m *mockConversationRepo) Count(ctx context.Context) (int64, error) {
	return int64(len(m.conversations)), nil
}

func (m *mockConversationRepo) CountByUser(ctx context.Context, userID string) (int64, error) {
	count := int64(0)
	for _, conv := range m.conversations {
		if conv.UserID == userID {
			count++
		}
	}
	return count, nil
}

func (m *mockConversationRepo) UpdateLastMessage(ctx context.Context, id string) error {
	if conv, exists := m.conversations[id]; exists {
		conv.LastMessageAt = time.Now()
		conv.UpdatedAt = time.Now()
	}
	return nil
}

func (m *mockConversationRepo) IncrementMessageCount(ctx context.Context, id string) error {
	if conv, exists := m.conversations[id]; exists {
		conv.MessageCount++
	}
	return nil
}

// mockMessageRepo is a mock implementation of MessageRepository
type mockMessageRepo struct {
	messages map[string]*conversationDomain.Message
	byConv   map[string][]*conversationDomain.Message
}

func newMockMessageRepo() *mockMessageRepo {
	return &mockMessageRepo{
		messages: make(map[string]*conversationDomain.Message),
		byConv:   make(map[string][]*conversationDomain.Message),
	}
}

func (m *mockMessageRepo) Create(ctx context.Context, msg *conversationDomain.Message) (string, error) {
	id := "msg_" + msg.ConversationID + "_" + string(rune(len(m.messages)))
	msg.ID = id
	m.messages[id] = msg
	m.byConv[msg.ConversationID] = append(m.byConv[msg.ConversationID], msg)
	return id, nil
}

func (m *mockMessageRepo) GetByConversationID(ctx context.Context, conversationID string, limit, offset int) ([]conversationDomain.Message, error) {
	msgs := m.byConv[conversationID]
	result := make([]conversationDomain.Message, 0)
	for _, msg := range msgs {
		result = append(result, *msg)
	}
	return result, nil
}

func (m *mockMessageRepo) GetByID(ctx context.Context, id string) (*conversationDomain.Message, error) {
	msg, exists := m.messages[id]
	if !exists {
		return nil, nil
	}
	return msg, nil
}

func (m *mockMessageRepo) CountByConversation(ctx context.Context, conversationID string) (int64, error) {
	return int64(len(m.byConv[conversationID])), nil
}

func TestNewConversationService(t *testing.T) {
	convRepo := newMockConversationRepo()
	msgRepo := newMockMessageRepo()
	svc := NewService(ServiceConfig{
		ConvRepo: convRepo,
		MsgRepo:  msgRepo,
	})

	if svc == nil {
		t.Fatal("Expected service to be created")
	}
}

func TestGetOrCreateConversation_Create(t *testing.T) {
	convRepo := newMockConversationRepo()
	msgRepo := newMockMessageRepo()
	svc := NewService(ServiceConfig{
		ConvRepo: convRepo,
		MsgRepo:  msgRepo,
	})

	ctx := context.Background()
	conv, err := svc.GetOrCreateConversation(ctx, "user-123", "+1234567890", "John Doe")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if conv.PhoneNumber != "+1234567890" {
		t.Errorf("Expected phone number +1234567890, got %s", conv.PhoneNumber)
	}
	if conv.ContactName != "John Doe" {
		t.Errorf("Expected contact name John Doe, got %s", conv.ContactName)
	}
	if conv.UserID != "user-123" {
		t.Errorf("Expected user ID user-123, got %s", conv.UserID)
	}
}

func TestGetOrCreateConversation_Get(t *testing.T) {
	convRepo := newMockConversationRepo()
	msgRepo := newMockMessageRepo()
	svc := NewService(ServiceConfig{
		ConvRepo: convRepo,
		MsgRepo:  msgRepo,
	})

	ctx := context.Background()

	// Create a conversation first
	conv1, _ := svc.GetOrCreateConversation(ctx, "user-123", "+1234567890", "John Doe")

	// Get the same conversation
	conv2, err := svc.GetOrCreateConversation(ctx, "user-456", "+1234567890", "Different Name")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should return the existing conversation
	if conv2.ID != conv1.ID {
		t.Errorf("Expected same conversation ID, got different: %s vs %s", conv1.ID, conv2.ID)
	}
}

func TestListConversations(t *testing.T) {
	convRepo := newMockConversationRepo()
	msgRepo := newMockMessageRepo()
	svc := NewService(ServiceConfig{
		ConvRepo: convRepo,
		MsgRepo:  msgRepo,
	})

	ctx := context.Background()

	// Create some conversations
	svc.GetOrCreateConversation(ctx, "user-123", "+1111111111", "User 1")
	svc.GetOrCreateConversation(ctx, "user-123", "+2222222222", "User 2")
	svc.GetOrCreateConversation(ctx, "user-456", "+3333333333", "User 3")

	// List as user-123
	userCtx := conversationDomain.UserContext{
		UserID:  "user-123",
		IsAdmin: false,
	}
	convs, total, err := svc.ListConversations(ctx, userCtx, 10, 0)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(convs) != 2 {
		t.Errorf("Expected 2 conversations for user-123, got %d", len(convs))
	}
	if total != 2 {
		t.Errorf("Expected total 2, got %d", total)
	}
}

func TestListConversationsAsAdmin(t *testing.T) {
	convRepo := newMockConversationRepo()
	msgRepo := newMockMessageRepo()
	svc := NewService(ServiceConfig{
		ConvRepo: convRepo,
		MsgRepo:  msgRepo,
	})

	ctx := context.Background()

	// Create conversations for different users
	svc.GetOrCreateConversation(ctx, "user-123", "+1111111111", "User 1")
	svc.GetOrCreateConversation(ctx, "user-456", "+2222222222", "User 2")

	// List as admin
	adminCtx := conversationDomain.UserContext{
		UserID:  "admin",
		IsAdmin: true,
	}
	convs, total, err := svc.ListConversations(ctx, adminCtx, 10, 0)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(convs) != 2 {
		t.Errorf("Expected 2 conversations for admin, got %d", len(convs))
	}
	if total != 2 {
		t.Errorf("Expected total 2, got %d", total)
	}
}

func TestListConversationsWithLimits(t *testing.T) {
	convRepo := newMockConversationRepo()
	msgRepo := newMockMessageRepo()
	svc := NewService(ServiceConfig{
		ConvRepo: convRepo,
		MsgRepo:  msgRepo,
	})

	ctx := context.Background()
	userCtx := conversationDomain.UserContext{
		UserID:  "user-123",
		IsAdmin: false,
	}

	// Test with negative limit (should default to 20)
	_, _, err := svc.ListConversations(ctx, userCtx, -1, 0)
	if err != nil {
		t.Fatalf("Expected no error with negative limit, got %v", err)
	}

	// Test with limit > 100 (should cap at 100)
	_, _, err = svc.ListConversations(ctx, userCtx, 200, 0)
	if err != nil {
		t.Fatalf("Expected no error with large limit, got %v", err)
	}

	// Test with negative offset (should default to 0)
	_, _, err = svc.ListConversations(ctx, userCtx, 10, -5)
	if err != nil {
		t.Fatalf("Expected no error with negative offset, got %v", err)
	}
}

func TestGetConversation(t *testing.T) {
	convRepo := newMockConversationRepo()
	msgRepo := newMockMessageRepo()
	svc := NewService(ServiceConfig{
		ConvRepo: convRepo,
		MsgRepo:  msgRepo,
	})

	ctx := context.Background()

	// Create a conversation
	conv, _ := svc.GetOrCreateConversation(ctx, "user-123", "+1234567890", "John Doe")

	// Get it
	userCtx := conversationDomain.UserContext{
		UserID:  "user-123",
		IsAdmin: false,
	}
	retrieved, err := svc.GetConversation(ctx, userCtx, conv.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if retrieved.PhoneNumber != "+1234567890" {
		t.Errorf("Expected phone +1234567890, got %s", retrieved.PhoneNumber)
	}
}

func TestGetConversationNotFound(t *testing.T) {
	convRepo := newMockConversationRepo()
	msgRepo := newMockMessageRepo()
	svc := NewService(ServiceConfig{
		ConvRepo: convRepo,
		MsgRepo:  msgRepo,
	})

	ctx := context.Background()
	userCtx := conversationDomain.UserContext{
		UserID:  "user-123",
		IsAdmin: false,
	}

	_, err := svc.GetConversation(ctx, userCtx, "non-existent-id")
	if !errors.Is(err, ErrConversationNotFound) {
		t.Errorf("Expected ErrConversationNotFound, got %v", err)
	}
}

func TestGetConversationForbidden(t *testing.T) {
	convRepo := newMockConversationRepo()
	msgRepo := newMockMessageRepo()
	svc := NewService(ServiceConfig{
		ConvRepo: convRepo,
		MsgRepo:  msgRepo,
	})

	ctx := context.Background()

	// Create conversation as user-123
	conv, _ := svc.GetOrCreateConversation(ctx, "user-123", "+1234567890", "John Doe")

	// Try to access as different user
	otherUserCtx := conversationDomain.UserContext{
		UserID:  "user-456",
		IsAdmin: false,
	}

	_, err := svc.GetConversation(ctx, otherUserCtx, conv.ID)
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("Expected ErrForbidden, got %v", err)
	}
}

func TestSaveIncomingMessage(t *testing.T) {
	convRepo := newMockConversationRepo()
	msgRepo := newMockMessageRepo()
	svc := NewService(ServiceConfig{
		ConvRepo: convRepo,
		MsgRepo:  msgRepo,
	})

	ctx := context.Background()

	msg, err := svc.SaveIncomingMessage(ctx, "+1234567890", "John Doe", "wa-msg-123", "Hello!", "text")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if msg.Direction != conversationDomain.DirectionIncoming {
		t.Errorf("Expected incoming direction, got %s", msg.Direction)
	}
	if msg.Content != "Hello!" {
		t.Errorf("Expected content Hello!, got %s", msg.Content)
	}
	if msg.WhatsAppMsgID != "wa-msg-123" {
		t.Errorf("Expected WhatsApp msg ID wa-msg-123, got %s", msg.WhatsAppMsgID)
	}
}

func TestSaveOutgoingMessage(t *testing.T) {
	convRepo := newMockConversationRepo()
	msgRepo := newMockMessageRepo()
	svc := NewService(ServiceConfig{
		ConvRepo: convRepo,
		MsgRepo:  msgRepo,
	})

	ctx := context.Background()

	// Create a conversation first
	conv, _ := svc.GetOrCreateConversation(ctx, "user-123", "+1234567890", "John Doe")

	msg, err := svc.SaveOutgoingMessage(ctx, conv.ID, "Hello back!", "RAG generated answer")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if msg.Direction != conversationDomain.DirectionOutgoing {
		t.Errorf("Expected outgoing direction, got %s", msg.Direction)
	}
	if msg.Content != "Hello back!" {
		t.Errorf("Expected content Hello back!, got %s", msg.Content)
	}
	if msg.RAGAnswer != "RAG generated answer" {
		t.Errorf("Expected RAG answer, got %s", msg.RAGAnswer)
	}
}

func TestGetMessages(t *testing.T) {
	convRepo := newMockConversationRepo()
	msgRepo := newMockMessageRepo()
	svc := NewService(ServiceConfig{
		ConvRepo: convRepo,
		MsgRepo:  msgRepo,
	})

	ctx := context.Background()

	// Create conversation and messages
	conv, _ := svc.GetOrCreateConversation(ctx, "user-123", "+1234567890", "John Doe")
	svc.SaveIncomingMessage(ctx, "+1234567890", "John Doe", "wa-1", "Message 1", "text")
	svc.SaveOutgoingMessage(ctx, conv.ID, "Reply 1", "")

	userCtx := conversationDomain.UserContext{
		UserID:  "user-123",
		IsAdmin: false,
	}

	msgs, total, err := svc.GetMessages(ctx, userCtx, conv.ID, 50, 0)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(msgs) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(msgs))
	}
	if total != 2 {
		t.Errorf("Expected total 2, got %d", total)
	}
}

func TestGetMessagesConversationNotFound(t *testing.T) {
	convRepo := newMockConversationRepo()
	msgRepo := newMockMessageRepo()
	svc := NewService(ServiceConfig{
		ConvRepo: convRepo,
		MsgRepo:  msgRepo,
	})

	ctx := context.Background()
	userCtx := conversationDomain.UserContext{
		UserID:  "user-123",
		IsAdmin: false,
	}

	_, _, err := svc.GetMessages(ctx, userCtx, "non-existent-conv", 50, 0)
	if !errors.Is(err, ErrConversationNotFound) {
		t.Errorf("Expected ErrConversationNotFound, got %v", err)
	}
}

func TestGetMessagesForbidden(t *testing.T) {
	convRepo := newMockConversationRepo()
	msgRepo := newMockMessageRepo()
	svc := NewService(ServiceConfig{
		ConvRepo: convRepo,
		MsgRepo:  msgRepo,
	})

	ctx := context.Background()

	// Create conversation as user-123
	conv, _ := svc.GetOrCreateConversation(ctx, "user-123", "+1234567890", "John Doe")

	// Try to access messages as different user
	otherUserCtx := conversationDomain.UserContext{
		UserID:  "user-456",
		IsAdmin: false,
	}

	_, _, err := svc.GetMessages(ctx, otherUserCtx, conv.ID, 50, 0)
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("Expected ErrForbidden, got %v", err)
	}
}

func TestGetMessagesWithLimits(t *testing.T) {
	convRepo := newMockConversationRepo()
	msgRepo := newMockMessageRepo()
	svc := NewService(ServiceConfig{
		ConvRepo: convRepo,
		MsgRepo:  msgRepo,
	})

	ctx := context.Background()

	// Create conversation
	conv, _ := svc.GetOrCreateConversation(ctx, "user-123", "+1234567890", "John Doe")

	userCtx := conversationDomain.UserContext{
		UserID:  "user-123",
		IsAdmin: false,
	}

	// Test with negative limit (should default to 50)
	_, _, err := svc.GetMessages(ctx, userCtx, conv.ID, -1, 0)
	if err != nil {
		t.Fatalf("Expected no error with negative limit, got %v", err)
	}

	// Test with limit > 200 (should cap at 200)
	_, _, err = svc.GetMessages(ctx, userCtx, conv.ID, 500, 0)
	if err != nil {
		t.Fatalf("Expected no error with large limit, got %v", err)
	}

	// Test with negative offset (should default to 0)
	_, _, err = svc.GetMessages(ctx, userCtx, conv.ID, 50, -5)
	if err != nil {
		t.Fatalf("Expected no error with negative offset, got %v", err)
	}
}
