package whatsapp

import (
	"github.com/gin-gonic/gin"
)

// Register registers WhatsApp webhook routes on the router group.
func Register(rg *gin.RouterGroup, handler *Handler) {
	whatsapp := rg.Group("/whatsapp")
	{
		whatsapp.GET("/webhook", handler.HandleWebhookVerification)
		whatsapp.POST("/webhook", handler.HandleIncomingMessage)
	}
}
