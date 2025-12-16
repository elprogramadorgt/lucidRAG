package repository

import (
	"context"
	"testing"

	"github.com/elprogramadorgt/lucidRAG/internal/domain"
)

func TestInMemorySessionRepository(t *testing.T) {
	repo := NewInMemorySessionRepository()
	ctx := context.Background()

	t.Run("Save and GetByID", func(t *testing.T) {
		session := &domain.ChatSession{
			UserPhoneNumber: "+1234567890",
			IsActive:        true,
			Context:         "test context",
		}

		err := repo.Save(ctx, session)
		if err != nil {
			t.Fatalf("Failed to save session: %v", err)
		}

		if session.ID == "" {
			t.Error("Expected session ID to be generated")
		}

		retrieved, err := repo.GetByID(ctx, session.ID)
		if err != nil {
			t.Fatalf("Failed to get session: %v", err)
		}

		if retrieved.UserPhoneNumber != session.UserPhoneNumber {
			t.Errorf("Expected phone %s, got %s", session.UserPhoneNumber, retrieved.UserPhoneNumber)
		}
	})

	t.Run("GetByPhoneNumber", func(t *testing.T) {
		phone := "+9876543210"
		session := &domain.ChatSession{
			UserPhoneNumber: phone,
			IsActive:        true,
		}

		repo.Save(ctx, session)

		retrieved, err := repo.GetByPhoneNumber(ctx, phone)
		if err != nil {
			t.Fatalf("Failed to get session by phone: %v", err)
		}

		if retrieved.UserPhoneNumber != phone {
			t.Errorf("Expected phone %s, got %s", phone, retrieved.UserPhoneNumber)
		}
	})

	t.Run("UpdateLastMessage", func(t *testing.T) {
		session := &domain.ChatSession{
			UserPhoneNumber: "+5555555555",
			IsActive:        true,
		}

		repo.Save(ctx, session)
		originalTime := session.LastMessageAt

		err := repo.UpdateLastMessage(ctx, session.ID)
		if err != nil {
			t.Fatalf("Failed to update last message: %v", err)
		}

		retrieved, _ := repo.GetByID(ctx, session.ID)
		if !retrieved.LastMessageAt.After(originalTime) {
			t.Error("Expected LastMessageAt to be updated")
		}
	})

	t.Run("SetInactive", func(t *testing.T) {
		session := &domain.ChatSession{
			UserPhoneNumber: "+7777777777",
			IsActive:        true,
		}

		repo.Save(ctx, session)

		err := repo.SetInactive(ctx, session.ID)
		if err != nil {
			t.Fatalf("Failed to set inactive: %v", err)
		}

		retrieved, _ := repo.GetByID(ctx, session.ID)
		if retrieved.IsActive {
			t.Error("Expected session to be inactive")
		}
	})

	t.Run("List sessions", func(t *testing.T) {
		sessions, err := repo.List(ctx, 10, 0)
		if err != nil {
			t.Fatalf("Failed to list sessions: %v", err)
		}

		if len(sessions) == 0 {
			t.Error("Expected at least one session")
		}
	})

	t.Run("GetByID not found", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "nonexistent")
		if err == nil {
			t.Error("Expected error for nonexistent session")
		}
	})
}
