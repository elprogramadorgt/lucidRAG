package conversation

import "github.com/gin-gonic/gin"

// Register registers conversation routes on the router group.
func Register(rg *gin.RouterGroup, handler *Handler) {
	rg.GET("", handler.ListConversations)
	rg.GET("/:id", handler.GetConversation)
	rg.GET("/:id/messages", handler.GetMessages)
}
