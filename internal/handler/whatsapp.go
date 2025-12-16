package handler

import (
	"encoding/json"
	"net/http"

	"github.com/elprogramadorgt/lucidRAG/internal/domain"
	"github.com/elprogramadorgt/lucidRAG/pkg/logger"
)

// WhatsAppHandler handles WhatsApp webhook requests
type WhatsAppHandler struct {
	whatsappSvc domain.WhatsAppService
	logger      *logger.Logger
}

// NewWhatsAppHandler creates a new WhatsApp handler
func NewWhatsAppHandler(whatsappSvc domain.WhatsAppService, log *logger.Logger) *WhatsAppHandler {
	return &WhatsAppHandler{
		whatsappSvc: whatsappSvc,
		logger:      log,
	}
}

// VerifyWebhook handles webhook verification
func (h *WhatsAppHandler) VerifyWebhook(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("hub.mode")
	token := r.URL.Query().Get("hub.verify_token")
	challenge := r.URL.Query().Get("hub.challenge")

	result, err := h.whatsappSvc.VerifyWebhook(token, mode, challenge)
	if err != nil {
		h.logger.Error("Webhook verification failed: %v", err)
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(result))
}

// HandleWebhook processes incoming webhook events
func (h *WhatsAppHandler) HandleWebhook(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var payload map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		h.logger.Error("Failed to decode webhook payload: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	payloadBytes, _ := json.Marshal(payload)
	if err := h.whatsappSvc.ProcessWebhook(r.Context(), payloadBytes); err != nil {
		h.logger.Error("Failed to process webhook: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
