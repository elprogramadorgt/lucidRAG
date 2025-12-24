package document

import "github.com/gin-gonic/gin"

func Register(rg *gin.RouterGroup, handler *Handler) {
	rg.GET("", handler.List)
	rg.POST("", handler.Create)
	rg.PUT("", handler.Update)
	rg.DELETE("", handler.Delete)
}
