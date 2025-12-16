package repository

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/elprogramadorgt/lucidRAG/internal/domain"
)

// InMemorySessionRepository is an in-memory implementation of SessionRepository
type InMemorySessionRepository struct {
	sessions      map[string]*domain.ChatSession
	phoneIndex    map[string]string // phoneNumber -> sessionID
	mu            sync.RWMutex
	idCounter     uint64
}

// NewInMemorySessionRepository creates a new in-memory session repository
func NewInMemorySessionRepository() *InMemorySessionRepository {
	return &InMemorySessionRepository{
		sessions:   make(map[string]*domain.ChatSession),
		phoneIndex: make(map[string]string),
		idCounter:  0,
	}
}

// Save saves a session
func (r *InMemorySessionRepository) Save(ctx context.Context, session *domain.ChatSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if session.ID == "" {
		id := atomic.AddUint64(&r.idCounter, 1)
		session.ID = fmt.Sprintf("session_%d", id)
	}

	if session.StartedAt.IsZero() {
		session.StartedAt = time.Now()
	}

	if session.LastMessageAt.IsZero() {
		session.LastMessageAt = time.Now()
	}

	r.sessions[session.ID] = session
	r.phoneIndex[session.UserPhoneNumber] = session.ID

	return nil
}

// GetByID retrieves a session by ID
func (r *InMemorySessionRepository) GetByID(ctx context.Context, id string) (*domain.ChatSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	session, exists := r.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	return session, nil
}

// GetByPhoneNumber retrieves a session by phone number
func (r *InMemorySessionRepository) GetByPhoneNumber(ctx context.Context, phoneNumber string) (*domain.ChatSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sessionID, exists := r.phoneIndex[phoneNumber]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	session, exists := r.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	return session, nil
}

// UpdateLastMessage updates the last message timestamp for a session
func (r *InMemorySessionRepository) UpdateLastMessage(ctx context.Context, sessionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	session, exists := r.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	session.LastMessageAt = time.Now()
	return nil
}

// SetInactive marks a session as inactive
func (r *InMemorySessionRepository) SetInactive(ctx context.Context, sessionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	session, exists := r.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	session.IsActive = false
	return nil
}

// List retrieves all sessions with pagination
func (r *InMemorySessionRepository) List(ctx context.Context, limit, offset int) ([]*domain.ChatSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	sessions := make([]*domain.ChatSession, 0, limit)
	count := 0
	for _, session := range r.sessions {
		if count >= offset {
			sessions = append(sessions, session)
			if len(sessions) >= limit {
				break
			}
		}
		count++
	}

	return sessions, nil
}
