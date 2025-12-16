package repository

import (
	"context"
	"testing"
	"time"

	"github.com/elprogramadorgt/lucidRAG/internal/domain"
)

func TestInMemoryMessageRepository(t *testing.T) {
	repo := NewInMemoryMessageRepository()
	ctx := context.Background()

	t.Run("Save and GetByID", func(t *testing.T) {
		msg := &domain.Message{
			From:        "+1234567890",
			To:          "PHONE_ID",
			Content:     "Test message",
			MessageType: "text",
			Status:      "sent",
		}

		err := repo.Save(ctx, msg)
		if err != nil {
			t.Fatalf("Failed to save message: %v", err)
		}

		if msg.ID == "" {
			t.Error("Expected message ID to be generated")
		}

		retrieved, err := repo.GetByID(ctx, msg.ID)
		if err != nil {
			t.Fatalf("Failed to get message: %v", err)
		}

		if retrieved.Content != msg.Content {
			t.Errorf("Expected content %s, got %s", msg.Content, retrieved.Content)
		}
	})

	t.Run("GetBySession", func(t *testing.T) {
		sessionID := "+test123"
		
		msg1 := &domain.Message{
			From:        sessionID,
			To:          "PHONE_ID",
			Content:     "Message 1",
			MessageType: "text",
			Timestamp:   time.Now(),
		}
		
		msg2 := &domain.Message{
			From:        sessionID,
			To:          "PHONE_ID",
			Content:     "Message 2",
			MessageType: "text",
			Timestamp:   time.Now(),
		}

		repo.Save(ctx, msg1)
		repo.Save(ctx, msg2)

		messages, err := repo.GetBySession(ctx, sessionID, 10, 0)
		if err != nil {
			t.Fatalf("Failed to get messages by session: %v", err)
		}

		if len(messages) < 2 {
			t.Errorf("Expected at least 2 messages, got %d", len(messages))
		}
	})

	t.Run("GetByID not found", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "nonexistent")
		if err == nil {
			t.Error("Expected error for nonexistent message")
		}
	})
}
