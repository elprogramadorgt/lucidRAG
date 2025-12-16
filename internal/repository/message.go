package repository

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/elprogramadorgt/lucidRAG/internal/domain"
)

// InMemoryMessageRepository is an in-memory implementation of MessageRepository
type InMemoryMessageRepository struct {
	messages       map[string]*domain.Message
	sessionIndex   map[string][]string // sessionID -> []messageID
	mu             sync.RWMutex
	idCounter      uint64
}

// NewInMemoryMessageRepository creates a new in-memory message repository
func NewInMemoryMessageRepository() *InMemoryMessageRepository {
	return &InMemoryMessageRepository{
		messages:     make(map[string]*domain.Message),
		sessionIndex: make(map[string][]string),
		idCounter:    0,
	}
}

// Save saves a message
func (r *InMemoryMessageRepository) Save(ctx context.Context, message *domain.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if message.ID == "" {
		id := atomic.AddUint64(&r.idCounter, 1)
		message.ID = fmt.Sprintf("msg_%d", id)
	}

	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}

	r.messages[message.ID] = message
	
	// For now, use the From field as a proxy for session identification
	// In a real implementation, this would be tied to actual session IDs
	sessionKey := message.From
	r.sessionIndex[sessionKey] = append(r.sessionIndex[sessionKey], message.ID)

	return nil
}

// GetByID retrieves a message by ID
func (r *InMemoryMessageRepository) GetByID(ctx context.Context, id string) (*domain.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	msg, exists := r.messages[id]
	if !exists {
		return nil, fmt.Errorf("message not found")
	}

	return msg, nil
}

// GetBySession retrieves messages for a session
func (r *InMemoryMessageRepository) GetBySession(ctx context.Context, sessionID string, limit, offset int) ([]*domain.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	messageIDs, exists := r.sessionIndex[sessionID]
	if !exists {
		return []*domain.Message{}, nil
	}

	// Apply pagination
	start := offset
	if start >= len(messageIDs) {
		return []*domain.Message{}, nil
	}

	end := start + limit
	if end > len(messageIDs) {
		end = len(messageIDs)
	}

	messages := make([]*domain.Message, 0, end-start)
	for i := start; i < end; i++ {
		if msg, exists := r.messages[messageIDs[i]]; exists {
			messages = append(messages, msg)
		}
	}

	return messages, nil
}

// SaveWithSession saves a message and associates it with a session ID
func (r *InMemoryMessageRepository) SaveWithSession(ctx context.Context, message *domain.Message, sessionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if message.ID == "" {
		id := atomic.AddUint64(&r.idCounter, 1)
		message.ID = fmt.Sprintf("msg_%d", id)
	}

	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}

	r.messages[message.ID] = message
	r.sessionIndex[sessionID] = append(r.sessionIndex[sessionID], message.ID)

	return nil
}
