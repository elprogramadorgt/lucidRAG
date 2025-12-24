package auth

import "github.com/gin-gonic/gin"

func Register(rg *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	auth := rg.Group("/auth")
	{
		auth.POST("/register", handler.Register)
		auth.POST("/login", handler.Login)
		auth.GET("/me", authMiddleware, handler.Me)
	}
}
