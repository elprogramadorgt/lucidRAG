package auth

import "github.com/gin-gonic/gin"

// Register registers authentication routes on the router group.
func Register(rg *gin.RouterGroup, handler *Handler, authMiddleware gin.HandlerFunc) {
	auth := rg.Group("/auth")
	{
		auth.POST("/register", handler.Register)
		auth.POST("/login", handler.Login)
		auth.POST("/logout", handler.Logout)
		auth.GET("/me", authMiddleware, handler.Me)
	}
}

// RegisterOAuth registers OAuth authentication routes on the router group.
func RegisterOAuth(rg *gin.RouterGroup, handler *OAuthHandler) {
	oauth := rg.Group("/auth/oauth")
	{
		// Get enabled providers
		oauth.GET("/providers", handler.GetProviders)

		// Google OAuth
		oauth.GET("/google", handler.GoogleLogin)
		oauth.GET("/google/callback", handler.GoogleCallback)

		// Facebook OAuth
		oauth.GET("/facebook", handler.FacebookLogin)
		oauth.GET("/facebook/callback", handler.FacebookCallback)

		// Apple OAuth
		oauth.GET("/apple", handler.AppleLogin)
		oauth.POST("/apple/callback", handler.AppleCallback) // Apple uses form_post
	}
}
