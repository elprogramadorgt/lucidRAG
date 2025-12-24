package rag

import "github.com/gin-gonic/gin"

func Register(rg *gin.RouterGroup, handler *Handler) {
	rg.POST("/query", handler.Query)
}
