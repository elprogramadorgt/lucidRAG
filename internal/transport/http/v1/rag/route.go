package rag

import "github.com/gin-gonic/gin"

// Register registers RAG routes on the router group.
func Register(rg *gin.RouterGroup, handler *Handler) {
	rg.POST("/query", handler.Query)
}
