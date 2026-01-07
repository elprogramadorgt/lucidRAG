package document

import "github.com/gin-gonic/gin"

// Register registers document routes on the router group.
func Register(rg *gin.RouterGroup, handler *Handler) {
	rg.GET("", handler.List)
	rg.POST("", handler.Create)
	rg.PUT("", handler.Update)
	rg.DELETE("", handler.Delete)
}
