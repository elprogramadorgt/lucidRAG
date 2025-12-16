package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/elprogramadorgt/lucidRAG/internal/domain"
	"github.com/elprogramadorgt/lucidRAG/pkg/logger"
)

// ConversationHandler handles conversation-related requests
type ConversationHandler struct {
	sessionRepo domain.SessionRepository
	messageRepo domain.MessageRepository
	logger      *logger.Logger
}

// NewConversationHandler creates a new conversation handler
func NewConversationHandler(sessionRepo domain.SessionRepository, messageRepo domain.MessageRepository, log *logger.Logger) *ConversationHandler {
	return &ConversationHandler{
		sessionRepo: sessionRepo,
		messageRepo: messageRepo,
		logger:      log,
	}
}

// ListSessions returns all chat sessions
func (h *ConversationHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse pagination parameters
	limit := 50
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get sessions from repository
	sessions, err := h.sessionRepo.List(r.Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to list sessions: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"sessions": sessions,
		"limit":    limit,
		"offset":   offset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetSession returns a specific session by ID
func (h *ConversationHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("id")
	if sessionID == "" {
		http.Error(w, "Missing session ID", http.StatusBadRequest)
		return
	}

	session, err := h.sessionRepo.GetByID(r.Context(), sessionID)
	if err != nil {
		h.logger.Error("Failed to get session: %v", err)
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(session)
}

// GetMessages returns messages for a specific session
func (h *ConversationHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		http.Error(w, "Missing session_id parameter", http.StatusBadRequest)
		return
	}

	// Parse pagination parameters
	limit := 100
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get messages from repository
	messages, err := h.messageRepo.GetBySession(r.Context(), sessionID, limit, offset)
	if err != nil {
		h.logger.Error("Failed to get messages: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"messages":   messages,
		"session_id": sessionID,
		"limit":      limit,
		"offset":     offset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
