package system

import "github.com/gin-gonic/gin"

// Register registers system administration routes on the router group.
func Register(rg *gin.RouterGroup, handler *Handler) {
	rg.GET("/info", handler.GetServerInfo)
	rg.GET("/logs", handler.ListLogs)
	rg.GET("/logs/stats", handler.GetStats)
	rg.DELETE("/logs", handler.CleanupLogs)
}
