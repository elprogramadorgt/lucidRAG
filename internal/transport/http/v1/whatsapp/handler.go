package whatsapp

import (
	"net/http"

	whatsappApp "github.com/elprogramadorgt/lucidRAG/internal/domain/whatsapp"
	"github.com/elprogramadorgt/lucidRAG/internal/transport/http/v1/whatsapp/dto"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	svc                whatsappApp.Service
	webhookVerifyToken string
}

func NewHandler(svc whatsappApp.Service, webhookVerifyToken string) *Handler {
	return &Handler{svc: svc, webhookVerifyToken: webhookVerifyToken}
}

func (h *Handler) HandleWebhookVerification(ctx *gin.Context) {
	var request dto.HookRequest
	if err := ctx.ShouldBindQuery(&request); err != nil {
		logrus.Error(err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Bad Request"})
		return
	}

	challenge, err := h.svc.VerifyWebhook(mapToHookInput(request), h.webhookVerifyToken)
	if err != nil {
		ctx.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, toHookVerificationDTO(challenge))
}
