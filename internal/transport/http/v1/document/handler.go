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

// Handler handles document-related HTTP requests.
type Handler struct {
	svc documentDomain.Service
	log *logger.Logger
}

// NewHandler creates a new document handler.
func NewHandler(svc documentDomain.Service, log *logger.Logger) *Handler {
	return &Handler{
		svc: svc,
		log: log.With("handler", "document"),
	}
}

func getUserContext(ctx *gin.Context) documentDomain.UserContext {
	userID := ctx.GetString("user_id")
	role := ctx.GetString("user_role")
	return documentDomain.UserContext{
		UserID:  userID,
		IsAdmin: role == "admin",
	}
}

func (h *Handler) List(ctx *gin.Context) {
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))
	userCtx := getUserContext(ctx)

	id := ctx.Query("id")
	if id != "" {
		h.getByID(ctx, userCtx, id)
		return
	}

	docs, total, err := h.svc.ListDocuments(ctx.Request.Context(), userCtx, limit, offset)
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

func (h *Handler) getByID(ctx *gin.Context, userCtx documentDomain.UserContext, id string) {
	doc, err := h.svc.GetDocument(ctx.Request.Context(), userCtx, id)
	if err != nil {
		if errors.Is(err, docApp.ErrDocumentNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
			return
		}
		if errors.Is(err, docApp.ErrForbidden) {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
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

	userCtx := getUserContext(ctx)
	doc := &documentDomain.Document{
		Title:    req.Title,
		Content:  req.Content,
		Source:   req.Source,
		Metadata: req.Metadata,
	}

	id, err := h.svc.CreateDocument(ctx.Request.Context(), userCtx, doc)
	if err != nil {
		h.log.Error("failed to create document", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create document"})
		return
	}

	if userCtx.IsAdmin {
		h.log.Info("admin_activity", "action", "document_create", "admin_id", userCtx.UserID, "document_id", id, "title", req.Title)
	} else {
		h.log.Info("document_create", "user_id", userCtx.UserID, "document_id", id, "title", req.Title)
	}
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

	userCtx := getUserContext(ctx)
	doc := &documentDomain.Document{
		ID:       req.ID,
		Title:    req.Title,
		Content:  req.Content,
		Source:   req.Source,
		Metadata: req.Metadata,
		IsActive: req.IsActive,
	}

	err := h.svc.UpdateDocument(ctx.Request.Context(), userCtx, doc)
	if err != nil {
		if errors.Is(err, docApp.ErrDocumentNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
			return
		}
		if errors.Is(err, docApp.ErrForbidden) {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		h.log.Error("failed to update document", "error", err, "id", req.ID)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update document"})
		return
	}

	if userCtx.IsAdmin {
		h.log.Info("admin_activity", "action", "document_update", "admin_id", userCtx.UserID, "document_id", req.ID)
	} else {
		h.log.Info("document_update", "user_id", userCtx.UserID, "document_id", req.ID)
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "document updated successfully"})
}

func (h *Handler) Delete(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	}

	userCtx := getUserContext(ctx)
	err := h.svc.DeleteDocument(ctx.Request.Context(), userCtx, id)
	if err != nil {
		if errors.Is(err, docApp.ErrDocumentNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "document not found"})
			return
		}
		if errors.Is(err, docApp.ErrForbidden) {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
		h.log.Error("failed to delete document", "error", err, "id", id)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete document"})
		return
	}

	if userCtx.IsAdmin {
		h.log.Info("admin_activity", "action", "document_delete", "admin_id", userCtx.UserID, "document_id", id)
	} else {
		h.log.Info("document_delete", "user_id", userCtx.UserID, "document_id", id)
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "document deleted successfully"})
}
