package user

import (
	"github.com/gin-gonic/gin"
)

// Register registers user routes on the router group.
func Register(rg *gin.RouterGroup, handler *Handler) {
	user := rg.Group("/user")
	{

		user.POST("/register", handler.HandleRegister)
		user.POST("/login", handler.HandleLogin)
	}
}
