package rag

import (
	"errors"
	"net/http"

	ragApp "github.com/elprogramadorgt/lucidRAG/internal/application/rag"
	ragDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/rag"
	"github.com/elprogramadorgt/lucidRAG/pkg/logger"
	"github.com/gin-gonic/gin"
)

// Handler handles RAG query endpoints.
type Handler struct {
	svc ragDomain.Service
	log *logger.Logger
}

// NewHandler creates a new RAG handler.
func NewHandler(svc ragDomain.Service, log *logger.Logger) *Handler {
	return &Handler{
		svc: svc,
		log: log.With("handler", "rag"),
	}
}

type queryRequest struct {
	Query     string  `json:"query" binding:"required"`
	TopK      int     `json:"top_k"`
	Threshold float64 `json:"threshold"`
}

// Query handles RAG query requests.
func (h *Handler) Query(ctx *gin.Context) {
	var req queryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	query := ragDomain.Query{
		Query:     req.Query,
		TopK:      req.TopK,
		Threshold: req.Threshold,
	}

	response, err := h.svc.Query(ctx.Request.Context(), query)
	if err != nil {
		if errors.Is(err, ragApp.ErrInvalidQuery) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid query"})
			return
		}
		h.log.Error("failed to process RAG query", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process query"})
		return
	}

	h.log.Info("RAG query processed",
		"request_id", ctx.GetString("request_id"),
		"query_length", len(req.Query),
		"processing_time_ms", response.ProcessingTimeMs,
	)

	ctx.JSON(http.StatusOK, response)
}
