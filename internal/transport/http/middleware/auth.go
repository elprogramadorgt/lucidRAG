package middleware

import (
	"net/http"
	"strings"

	userDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/user"
	"github.com/gin-gonic/gin"
)

const cookieName = "lucidrag_token"

// AuthMiddleware validates JWT tokens from cookies or Authorization header.
func AuthMiddleware(userSvc userDomain.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		var token string

		// First, try to get token from cookie (primary method for browser clients)
		if cookieToken, err := c.Cookie(cookieName); err == nil && cookieToken != "" {
			token = cookieToken
		}

		// Fall back to Authorization header (for API clients)
		if token == "" {
			authHeader := c.GetHeader("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
					token = parts[1]
				}
			}
		}

		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}

		claims, err := userSvc.ValidateToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)
		c.Next()
	}
}

// RequireRole restricts access to users with one of the specified roles.
func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("user_role")
		if userRole == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		for _, role := range roles {
			if userRole == role {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
	}
}
