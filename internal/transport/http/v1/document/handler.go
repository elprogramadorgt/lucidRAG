package document

import (
	"errors"
	"net/http"
	"strconv"

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
		log: log.With("handler", "document"),
	}
}

func (h *Handler) List(ctx *gin.Context) {
	limit, err := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	if err != nil || limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	offset, err := strconv.Atoi(ctx.DefaultQuery("offset", "0"))
	if err != nil || offset < 0 {
		offset = 0
	}

	id := ctx.Query("id")
	if id != "" {
		h.getByID(ctx, id)
		return
	}

	docs, total, err := h.svc.ListDocuments(ctx.Request.Context(), limit, offset)
	if err != nil {
		h.log.Error("failed to list documents", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list documents"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"documents": docs,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

func (h *Handler) getByID(ctx *gin.Context, id string) {
	doc, err := h.svc.GetDocument(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, docApp.ErrDocumentNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
			return
		}
		h.log.Error("failed to get document", "error", err, "id", id)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get document"})
		return
	}

	ctx.JSON(http.StatusOK, doc)
}

type createDocumentRequest struct {
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Source   string `json:"source"`
	Metadata string `json:"metadata"`
}

func (h *Handler) Create(ctx *gin.Context) {
	var req createDocumentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	doc := &documentDomain.Document{
		Title:    req.Title,
		Content:  req.Content,
		Source:   req.Source,
		Metadata: req.Metadata,
	}

	id, err := h.svc.CreateDocument(ctx.Request.Context(), doc)
	if err != nil {
		h.log.Error("failed to create document", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create document"})
		return
	}

	h.log.Info("document created", "id", id, "title", req.Title)
	ctx.JSON(http.StatusCreated, gin.H{
		"id":      id,
		"message": "document created successfully",
	})
}

type updateDocumentRequest struct {
	ID       string `json:"id" binding:"required"`
	Title    string `json:"title" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Source   string `json:"source"`
	Metadata string `json:"metadata"`
	IsActive bool   `json:"is_active"`
}

func (h *Handler) Update(ctx *gin.Context) {
	var req updateDocumentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	doc := &documentDomain.Document{
		ID:       req.ID,
		Title:    req.Title,
		Content:  req.Content,
		Source:   req.Source,
		Metadata: req.Metadata,
		IsActive: req.IsActive,
	}

	err := h.svc.UpdateDocument(ctx.Request.Context(), doc)
	if err != nil {
		if errors.Is(err, docApp.ErrDocumentNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
			return
		}
		h.log.Error("failed to update document", "error", err, "id", req.ID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update document"})
		return
	}

	h.log.Info("document updated", "id", req.ID)
	ctx.JSON(http.StatusOK, gin.H{"message": "document updated successfully"})
}

func (h *Handler) Delete(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	err := h.svc.DeleteDocument(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, docApp.ErrDocumentNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
			return
		}
		h.log.Error("failed to delete document", "error", err, "id", id)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete document"})
		return
	}

	h.log.Info("document deleted", "id", id)
	ctx.JSON(http.StatusOK, gin.H{"message": "document deleted successfully"})
}
