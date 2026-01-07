package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/elprogramadorgt/lucidRAG/internal/config"
	userDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/user"
	"github.com/elprogramadorgt/lucidRAG/pkg/logger"
	"github.com/gin-gonic/gin"
)

// OAuthHandler handles OAuth authentication flows.
type OAuthHandler struct {
	userSvc      userDomain.Service
	log          *logger.Logger
	oauthConfig  config.OAuthConfig
	cookieConfig CookieConfig
}

// NewOAuthHandler creates a new OAuthHandler.
func NewOAuthHandler(userSvc userDomain.Service, log *logger.Logger, oauthCfg config.OAuthConfig, cookieCfg CookieConfig) *OAuthHandler {
	return &OAuthHandler{
		userSvc:      userSvc,
		log:          log.With("handler", "oauth"),
		oauthConfig:  oauthCfg,
		cookieConfig: cookieCfg,
	}
}

// OAuthUserInfo represents user info from OAuth providers.
type OAuthUserInfo struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
	Provider  string
}

// ProvidersResponse returns which OAuth providers are enabled.
type ProvidersResponse struct {
	Google   bool `json:"google"`
	Facebook bool `json:"facebook"`
	Apple    bool `json:"apple"`
}

// GetProviders returns which OAuth providers are enabled.
func (h *OAuthHandler) GetProviders(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, ProvidersResponse{
		Google:   h.oauthConfig.Google.Enabled,
		Facebook: h.oauthConfig.Facebook.Enabled,
		Apple:    h.oauthConfig.Apple.Enabled,
	})
}

// generateState creates a random state parameter for OAuth.
func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// handleOAuthUser processes the authenticated OAuth user.
func (h *OAuthHandler) handleOAuthUser(ctx *gin.Context, userInfo *OAuthUserInfo) {
	if userInfo.Email == "" {
		h.log.Warn("oauth_user", "provider", userInfo.Provider, "error", "no email provided")
		h.redirectWithError(ctx, "Email is required for registration")
		return
	}

	// Try to find existing user by email
	user, err := h.userSvc.GetUserByEmail(ctx.Request.Context(), userInfo.Email)
	if err != nil {
		// User doesn't exist, create new one
		firstName := userInfo.FirstName
		if firstName == "" {
			firstName = strings.Split(userInfo.Email, "@")[0]
		}
		lastName := userInfo.LastName
		if lastName == "" {
			lastName = "User"
		}

		user, err = h.userSvc.RegisterOAuth(ctx.Request.Context(), userDomain.User{
			Email:     userInfo.Email,
			FirstName: firstName,
			LastName:  lastName,
		}, userInfo.Provider, userInfo.ID)
		if err != nil {
			h.log.Error("oauth_register", "provider", userInfo.Provider, "error", err)
			h.redirectWithError(ctx, "Failed to create account")
			return
		}
		h.log.Info("oauth_register", "provider", userInfo.Provider, "user_id", user.ID, "email", user.Email)
	} else {
		h.log.Info("oauth_login", "provider", userInfo.Provider, "user_id", user.ID, "email", user.Email)
	}

	// Generate JWT token
	token, err := h.userSvc.GenerateToken(user)
	if err != nil {
		h.log.Error("oauth_token", "error", err)
		h.redirectWithError(ctx, "Failed to generate session")
		return
	}

	// Set auth cookie
	h.cookieConfig.SetAuthCookie(ctx, token)

	// Redirect to frontend with success
	redirectURL := fmt.Sprintf("%s/oauth/callback?success=true", h.oauthConfig.RedirectBaseURL)
	ctx.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

func (h *OAuthHandler) redirectWithError(ctx *gin.Context, errorMsg string) {
	redirectURL := fmt.Sprintf("%s/oauth/callback?error=%s", h.oauthConfig.RedirectBaseURL, url.QueryEscape(errorMsg))
	ctx.Redirect(http.StatusTemporaryRedirect, redirectURL)
}
