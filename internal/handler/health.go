package handler

import (
	"encoding/json"
	"net/http"

	"github.com/elprogramadorgt/lucidRAG/pkg/logger"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	logger *logger.Logger
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(log *logger.Logger) *HealthHandler {
	return &HealthHandler{
		logger: log,
	}
}

// HealthCheck returns the health status of the service
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "healthy",
		"service": "lucidRAG",
		"version": "0.1.0",
	})
}
