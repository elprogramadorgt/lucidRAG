package whatsapp

import (
	"github.com/gin-gonic/gin"
)

func Register(rg *gin.RouterGroup, handler *Handler) {
	whatsapp := rg.Group("/whatsapp")
	{
		whatsapp.GET("/webhook", handler.HandleWebhookVerification)
		whatsapp.POST("/webhook", handler.HandleIncomingMessage)
	}
}
