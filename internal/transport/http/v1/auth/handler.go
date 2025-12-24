package auth

import (
	"errors"
	"net/http"

	userApp "github.com/elprogramadorgt/lucidRAG/internal/application/user"
	userDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/user"
	"github.com/elprogramadorgt/lucidRAG/pkg/logger"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc userDomain.Service
	log *logger.Logger
}

func NewHandler(svc userDomain.Service, log *logger.Logger) *Handler {
	return &Handler{
		svc: svc,
		log: log.With("handler", "auth"),
	}
}

type registerRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type authResponse struct {
	Token string          `json:"token,omitempty"`
	User  *userDomain.User `json:"user"`
}

func (h *Handler) Register(ctx *gin.Context) {
	var req registerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	user, err := h.svc.Register(ctx.Request.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		if errors.Is(err, userApp.ErrEmailExists) {
			ctx.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}
		h.log.Error("failed to register user", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
		return
	}

	token, err := h.svc.Login(ctx.Request.Context(), req.Email, req.Password)
	if err != nil {
		h.log.Error("failed to generate token after registration", "error", err)
		ctx.JSON(http.StatusCreated, authResponse{User: user})
		return
	}

	h.log.Info("user registered", "user_id", user.ID, "email", user.Email)
	ctx.JSON(http.StatusCreated, authResponse{Token: token, User: user})
}

func (h *Handler) Login(ctx *gin.Context) {
	var req loginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	token, err := h.svc.Login(ctx.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, userApp.ErrInvalidCredentials) {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}
		h.log.Error("failed to login", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		return
	}

	h.log.Info("user logged in", "email", req.Email)
	ctx.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *Handler) Me(ctx *gin.Context) {
	userID := ctx.GetString("user_id")
	if userID == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	user, err := h.svc.GetUser(ctx.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, userApp.ErrUserNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		h.log.Error("failed to get user", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user"})
		return
	}

	ctx.JSON(http.StatusOK, user)
}
