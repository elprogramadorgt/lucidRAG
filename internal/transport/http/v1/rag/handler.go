package rag

import (
	"errors"
	"net/http"

	docApp "github.com/elprogramadorgt/lucidRAG/internal/application/document"
	documentDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/document"
	"github.com/elprogramadorgt/lucidRAG/pkg/logger"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc documentDomain.Service
	log *logger.Logger
}

func NewHandler(svc documentDomain.Service, log *logger.Logger) *Handler {
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

func (h *Handler) Query(ctx *gin.Context) {
	var req queryRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	query := documentDomain.RAGQuery{
		Query:     req.Query,
		TopK:      req.TopK,
		Threshold: req.Threshold,
	}

	response, err := h.svc.QueryRAG(ctx.Request.Context(), query)
	if err != nil {
		if errors.Is(err, docApp.ErrInvalidQuery) {
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
