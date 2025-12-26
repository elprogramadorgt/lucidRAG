package conversation

import (
	"testing"
	"time"
)

func TestMessageDirectionConstants(t *testing.T) {
	if DirectionIncoming != "incoming" {
		t.Errorf("Expected DirectionIncoming to be 'incoming', got '%s'", DirectionIncoming)
	}
	if DirectionOutgoing != "outgoing" {
		t.Errorf("Expected DirectionOutgoing to be 'outgoing', got '%s'", DirectionOutgoing)
	}
}

func TestConversationStruct(t *testing.T) {
	now := time.Now()
	conv := Conversation{
		ID:            "conv-123",
		UserID:        "user-456",
		PhoneNumber:   "+1234567890",
		ContactName:   "John Doe",
		LastMessageAt: now,
		MessageCount:  10,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if conv.ID != "conv-123" {
		t.Errorf("Expected ID 'conv-123', got '%s'", conv.ID)
	}
	if conv.UserID != "user-456" {
		t.Errorf("Expected UserID 'user-456', got '%s'", conv.UserID)
	}
	if conv.PhoneNumber != "+1234567890" {
		t.Errorf("Expected PhoneNumber '+1234567890', got '%s'", conv.PhoneNumber)
	}
	if conv.ContactName != "John Doe" {
		t.Errorf("Expected ContactName 'John Doe', got '%s'", conv.ContactName)
	}
	if conv.MessageCount != 10 {
		t.Errorf("Expected MessageCount 10, got %d", conv.MessageCount)
	}
}

func TestMessageStruct(t *testing.T) {
	now := time.Now()
	msg := Message{
		ID:             "msg-123",
		ConversationID: "conv-456",
		WhatsAppMsgID:  "wa-789",
		Direction:      DirectionIncoming,
		Content:        "Hello, world!",
		MessageType:    "text",
		RAGQueryID:     "rag-123",
		RAGAnswer:      "This is the answer",
		Timestamp:      now,
		CreatedAt:      now,
	}

	if msg.ID != "msg-123" {
		t.Errorf("Expected ID 'msg-123', got '%s'", msg.ID)
	}
	if msg.ConversationID != "conv-456" {
		t.Errorf("Expected ConversationID 'conv-456', got '%s'", msg.ConversationID)
	}
	if msg.WhatsAppMsgID != "wa-789" {
		t.Errorf("Expected WhatsAppMsgID 'wa-789', got '%s'", msg.WhatsAppMsgID)
	}
	if msg.Direction != DirectionIncoming {
		t.Errorf("Expected Direction 'incoming', got '%s'", msg.Direction)
	}
	if msg.Content != "Hello, world!" {
		t.Errorf("Expected Content 'Hello, world!', got '%s'", msg.Content)
	}
	if msg.MessageType != "text" {
		t.Errorf("Expected MessageType 'text', got '%s'", msg.MessageType)
	}
	if msg.RAGQueryID != "rag-123" {
		t.Errorf("Expected RAGQueryID 'rag-123', got '%s'", msg.RAGQueryID)
	}
	if msg.RAGAnswer != "This is the answer" {
		t.Errorf("Expected RAGAnswer 'This is the answer', got '%s'", msg.RAGAnswer)
	}
}

func TestConversationZeroValue(t *testing.T) {
	var conv Conversation
	if conv.ID != "" {
		t.Errorf("Expected empty ID, got '%s'", conv.ID)
	}
	if conv.MessageCount != 0 {
		t.Errorf("Expected MessageCount 0, got %d", conv.MessageCount)
	}
}

func TestMessageZeroValue(t *testing.T) {
	var msg Message
	if msg.ID != "" {
		t.Errorf("Expected empty ID, got '%s'", msg.ID)
	}
	if msg.Direction != "" {
		t.Errorf("Expected empty Direction, got '%s'", msg.Direction)
	}
	if msg.RAGQueryID != "" {
		t.Errorf("Expected empty RAGQueryID, got '%s'", msg.RAGQueryID)
	}
}

func TestMessageDirectionComparison(t *testing.T) {
	incomingMsg := Message{Direction: DirectionIncoming}
	outgoingMsg := Message{Direction: DirectionOutgoing}

	if incomingMsg.Direction == outgoingMsg.Direction {
		t.Error("Incoming and outgoing messages should have different directions")
	}

	if incomingMsg.Direction != DirectionIncoming {
		t.Error("Incoming message should have DirectionIncoming")
	}

	if outgoingMsg.Direction != DirectionOutgoing {
		t.Error("Outgoing message should have DirectionOutgoing")
	}
}
