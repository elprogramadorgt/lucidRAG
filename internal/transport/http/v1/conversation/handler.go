package conversation

import (
	"errors"
	"net/http"
	"strconv"

	convApp "github.com/elprogramadorgt/lucidRAG/internal/application/conversation"
	conversationDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/conversation"
	"github.com/elprogramadorgt/lucidRAG/pkg/logger"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc conversationDomain.Service
	log *logger.Logger
}

func NewHandler(svc conversationDomain.Service, log *logger.Logger) *Handler {
	return &Handler{
		svc: svc,
		log: log.With("handler", "conversation"),
	}
}

func (h *Handler) ListConversations(ctx *gin.Context) {
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	convs, total, err := h.svc.ListConversations(ctx.Request.Context(), limit, offset)
	if err != nil {
		h.log.Error("failed to list conversations", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list conversations"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"conversations": convs,
		"total":         total,
		"limit":         limit,
		"offset":        offset,
	})
}

func (h *Handler) GetConversation(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "conversation id is required"})
		return
	}

	conv, err := h.svc.GetConversation(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, convApp.ErrConversationNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "conversation not found"})
			return
		}
		h.log.Error("failed to get conversation", "error", err, "id", id)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get conversation"})
		return
	}

	ctx.JSON(http.StatusOK, conv)
}

func (h *Handler) GetMessages(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "conversation id is required"})
		return
	}

	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	msgs, total, err := h.svc.GetMessages(ctx.Request.Context(), id, limit, offset)
	if err != nil {
		h.log.Error("failed to get messages", "error", err, "conversation_id", id)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get messages"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"messages": msgs,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}
