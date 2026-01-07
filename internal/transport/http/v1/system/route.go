package system

import "github.com/gin-gonic/gin"

func Register(rg *gin.RouterGroup, handler *Handler) {
	rg.GET("/info", handler.GetServerInfo)
	rg.GET("/logs", handler.ListLogs)
	rg.GET("/logs/stats", handler.GetStats)
	rg.DELETE("/logs", handler.CleanupLogs)
}
