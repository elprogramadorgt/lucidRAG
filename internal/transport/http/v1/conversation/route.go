package conversation

import "github.com/gin-gonic/gin"

func Register(rg *gin.RouterGroup, handler *Handler) {
	rg.GET("", handler.ListConversations)
	rg.GET("/:id", handler.GetConversation)
	rg.GET("/:id/messages", handler.GetMessages)
}
