package user

import (
	userApp "github.com/elprogramadorgt/lucidRAG/internal/domain/user"
	"github.com/elprogramadorgt/lucidRAG/internal/transport/http/v1/user/dto"
	"github.com/gin-gonic/gin"
)

type Handler struct{ svc userApp.Service }

func NewHandler(svc userApp.Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) HandleRegister(ctx *gin.Context) {

	var req dto.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{"error": "invalid request"})
		return
	}

	_, err := h.svc.Register(ctx, userApp.User{
		Email:        req.Email,
		PasswordHash: req.Password,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
	})
	if err != nil {
		ctx.JSON(500, gin.H{"error": "failed to register user"})
		return
	}

	ctx.JSON(201, gin.H{"message": "user registered successfully"})

	// Implementation of user registration handler
}

func (h *Handler) HandleLogin(ctx *gin.Context) {
	// Implementation of user login handler
}
