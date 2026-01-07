package whatsapp

import (
	"net/http"

	conversationDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/conversation"
	ragDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/rag"
	whatsappDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/whatsapp"
	"github.com/elprogramadorgt/lucidRAG/internal/transport/http/v1/whatsapp/dto"
	"github.com/elprogramadorgt/lucidRAG/pkg/logger"
	"github.com/gin-gonic/gin"
)

// Handler handles WhatsApp webhook requests.
type Handler struct {
	svc                whatsappDomain.Service
	convSvc            conversationDomain.Service
	ragSvc             ragDomain.Service
	webhookVerifyToken string
	log                *logger.Logger
}

// HandlerConfig contains dependencies for creating a WhatsApp handler.
type HandlerConfig struct {
	WhatsAppSvc        whatsappDomain.Service
	ConversationSvc    conversationDomain.Service
	RAGSvc             ragDomain.Service
	WebhookVerifyToken string
	Log                *logger.Logger
}

// NewHandler creates a new WhatsApp handler.
func NewHandler(cfg HandlerConfig) *Handler {
	return &Handler{
		svc:                cfg.WhatsAppSvc,
		convSvc:            cfg.ConversationSvc,
		ragSvc:             cfg.RAGSvc,
		webhookVerifyToken: cfg.WebhookVerifyToken,
		log:                cfg.Log.With("handler", "whatsapp"),
	}
}

func (h *Handler) HandleWebhookVerification(ctx *gin.Context) {
	var request dto.HookRequest
	if err := ctx.ShouldBindQuery(&request); err != nil {
		h.log.Error("failed to bind query", "error", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Bad Request"})
		return
	}

	challenge, err := h.svc.VerifyWebhook(mapToHookInput(request), h.webhookVerifyToken)
	if err != nil {
		h.log.Warn("webhook verification failed", "error", err)
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, toHookVerificationDTO(challenge))
}

func (h *Handler) HandleIncomingMessage(ctx *gin.Context) {
	var payload dto.WebhookPayload
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		h.log.Error("failed to parse webhook payload", "error", err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Bad Request"})
		return
	}

	if payload.Object != "whatsapp_business_account" {
		h.log.Warn("unexpected webhook object type", "object", payload.Object)
		ctx.JSON(http.StatusOK, gin.H{"status": "ignored"})
		return
	}

	for _, entry := range payload.Entry {
		for _, change := range entry.Changes {
			for _, msg := range change.Value.Messages {
				h.processMessage(ctx, msg, change.Value.Contacts)
			}
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "received"})
}

func (h *Handler) processMessage(ctx *gin.Context, msg dto.Message, contacts []dto.Contact) {
	var senderName string
	for _, c := range contacts {
		if c.WaID == msg.From {
			senderName = c.Profile.Name
			break
		}
	}

	h.log.Info("received message",
		"request_id", ctx.GetString("request_id"),
		"from", msg.From,
		"sender_name", senderName,
		"type", msg.Type,
		"message_id", msg.ID,
	)

	if msg.Type != "text" || msg.Text == nil {
		return
	}

	content := msg.Text.Body

	if h.convSvc == nil {
		h.log.Debug("conversation service not configured, skipping message persistence")
		return
	}

	savedMsg, err := h.convSvc.SaveIncomingMessage(
		ctx.Request.Context(),
		msg.From,
		senderName,
		msg.ID,
		content,
		msg.Type,
	)
	if err != nil {
		h.log.Error("failed to save incoming message", "error", err)
		return
	}

	h.log.Info("message saved", "message_id", savedMsg.ID, "conversation_id", savedMsg.ConversationID)

	if h.ragSvc == nil {
		h.log.Debug("RAG service not configured, skipping RAG query")
		return
	}

	ragQuery := ragDomain.Query{
		Query:     content,
		TopK:      5,
		Threshold: 0.7,
	}

	ragResponse, err := h.ragSvc.Query(ctx.Request.Context(), ragQuery)
	if err != nil {
		h.log.Error("failed to query RAG", "error", err)
		return
	}

	_, err = h.convSvc.SaveOutgoingMessage(
		ctx.Request.Context(),
		savedMsg.ConversationID,
		ragResponse.Answer,
		ragResponse.Answer,
	)
	if err != nil {
		h.log.Error("failed to save outgoing message", "error", err)
		return
	}

	h.log.Info("RAG response saved",
		"conversation_id", savedMsg.ConversationID,
		"confidence", ragResponse.ConfidenceScore,
		"processing_time_ms", ragResponse.ProcessingTimeMs,
	)
}
