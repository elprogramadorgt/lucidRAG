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
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type authResponse struct {
	Token string           `json:"token"`
	User  *userDomain.User `json:"user"`
}

func (h *Handler) Register(ctx *gin.Context) {
	var req registerRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.Warn("registration_attempt", "status", "invalid_request", "ip", ctx.ClientIP(), "error", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	user, err := h.svc.Register(ctx.Request.Context(), userDomain.User{
		Email:        req.Email,
		PasswordHash: req.Password,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
	})
	if err != nil {
		if errors.Is(err, userApp.ErrEmailExists) {
			h.log.Warn("registration_attempt", "status", "failed", "email", req.Email, "ip", ctx.ClientIP(), "reason", "email_exists")
			ctx.JSON(http.StatusConflict, gin.H{"error": "email already registered"})
			return
		}
		h.log.Error("registration_attempt", "status", "error", "email", req.Email, "ip", ctx.ClientIP(), "error", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
		return
	}

	token, _, err := h.svc.Login(ctx.Request.Context(), req.Email, req.Password)
	if err != nil {
		h.log.Error("registration_attempt", "status", "partial", "user_id", user.ID, "email", user.Email, "ip", ctx.ClientIP(), "error", "token_generation_failed")
		ctx.JSON(http.StatusCreated, authResponse{User: user})
		return
	}

	h.log.Info("registration_attempt", "status", "success", "user_id", user.ID, "email", user.Email, "ip", ctx.ClientIP())
	ctx.JSON(http.StatusCreated, authResponse{Token: token, User: user})
}

func (h *Handler) Login(ctx *gin.Context) {
	var req loginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		h.log.Warn("login_attempt", "status", "invalid_request", "ip", ctx.ClientIP(), "error", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	token, user, err := h.svc.Login(ctx.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, userApp.ErrInvalidCredentials) {
			h.log.Warn("login_attempt", "status", "failed", "email", req.Email, "ip", ctx.ClientIP(), "reason", "invalid_credentials")
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}
		h.log.Error("login_attempt", "status", "error", "email", req.Email, "ip", ctx.ClientIP(), "error", err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		return
	}

	h.log.Info("login_attempt", "status", "success", "email", req.Email, "ip", ctx.ClientIP())
	ctx.JSON(http.StatusOK, authResponse{Token: token, User: user})
}

func (h *Handler) Logout(ctx *gin.Context) {
	h.log.Info("logout", "ip", ctx.ClientIP())
	ctx.JSON(http.StatusOK, gin.H{"message": "logged out"})
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
