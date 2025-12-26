package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/elprogramadorgt/lucidRAG/internal/config"
	userDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/user"
	"github.com/elprogramadorgt/lucidRAG/pkg/logger"
	"github.com/gin-gonic/gin"
)

type OAuthHandler struct {
	userSvc      userDomain.Service
	log          *logger.Logger
	oauthConfig  config.OAuthConfig
	cookieConfig CookieConfig
}

func NewOAuthHandler(userSvc userDomain.Service, log *logger.Logger, oauthCfg config.OAuthConfig, cookieCfg CookieConfig) *OAuthHandler {
	return &OAuthHandler{
		userSvc:      userSvc,
		log:          log.With("handler", "oauth"),
		oauthConfig:  oauthCfg,
		cookieConfig: cookieCfg,
	}
}

// OAuthUserInfo represents user info from OAuth providers
type OAuthUserInfo struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
	Provider  string
}

// ProvidersResponse returns which OAuth providers are enabled
type ProvidersResponse struct {
	Google   bool `json:"google"`
	Facebook bool `json:"facebook"`
	Apple    bool `json:"apple"`
}

// GetProviders returns which OAuth providers are enabled
func (h *OAuthHandler) GetProviders(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, ProvidersResponse{
		Google:   h.oauthConfig.Google.Enabled,
		Facebook: h.oauthConfig.Facebook.Enabled,
		Apple:    h.oauthConfig.Apple.Enabled,
	})
}

// generateState creates a random state parameter for OAuth
func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Google OAuth

func (h *OAuthHandler) GoogleLogin(ctx *gin.Context) {
	if !h.oauthConfig.Google.Enabled {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Google OAuth is not enabled"})
		return
	}

	state, err := generateState()
	if err != nil {
		h.log.Error("failed to generate state", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initiate OAuth"})
		return
	}

	// Store state in cookie for verification
	ctx.SetCookie("oauth_state", state, 600, "/", h.cookieConfig.Domain, h.cookieConfig.Secure, true)

	redirectURL := fmt.Sprintf("%s/api/v1/auth/oauth/google/callback", h.oauthConfig.RedirectBaseURL)
	authURL := fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=email%%20profile&state=%s",
		url.QueryEscape(h.oauthConfig.Google.ClientID),
		url.QueryEscape(redirectURL),
		url.QueryEscape(state),
	)

	ctx.Redirect(http.StatusTemporaryRedirect, authURL)
}

func (h *OAuthHandler) GoogleCallback(ctx *gin.Context) {
	if !h.oauthConfig.Google.Enabled {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Google OAuth is not enabled"})
		return
	}

	// Verify state
	state := ctx.Query("state")
	storedState, err := ctx.Cookie("oauth_state")
	if err != nil || state != storedState {
		h.log.Warn("oauth_callback", "provider", "google", "error", "invalid state")
		h.redirectWithError(ctx, "Invalid OAuth state")
		return
	}

	code := ctx.Query("code")
	if code == "" {
		h.redirectWithError(ctx, "No authorization code received")
		return
	}

	// Exchange code for token
	redirectURL := fmt.Sprintf("%s/api/v1/auth/oauth/google/callback", h.oauthConfig.RedirectBaseURL)
	tokenResp, err := h.exchangeGoogleCode(code, redirectURL)
	if err != nil {
		h.log.Error("google_token_exchange", "error", err)
		h.redirectWithError(ctx, "Failed to authenticate with Google")
		return
	}

	// Get user info
	userInfo, err := h.getGoogleUserInfo(tokenResp.AccessToken)
	if err != nil {
		h.log.Error("google_userinfo", "error", err)
		h.redirectWithError(ctx, "Failed to get user info from Google")
		return
	}

	h.handleOAuthUser(ctx, userInfo)
}

type googleTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

func (h *OAuthHandler) exchangeGoogleCode(code, redirectURL string) (*googleTokenResponse, error) {
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", h.oauthConfig.Google.ClientID)
	data.Set("client_secret", h.oauthConfig.Google.ClientSecret)
	data.Set("redirect_uri", redirectURL)
	data.Set("grant_type", "authorization_code")

	resp, err := http.Post("https://oauth2.googleapis.com/token", "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed: %s", string(body))
	}

	var tokenResp googleTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	return &tokenResp, nil
}

func (h *OAuthHandler) getGoogleUserInfo(accessToken string) (*OAuthUserInfo, error) {
	req, err := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		GivenName string `json:"given_name"`
		FamilyName string `json:"family_name"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return &OAuthUserInfo{
		ID:        data.ID,
		Email:     data.Email,
		FirstName: data.GivenName,
		LastName:  data.FamilyName,
		Provider:  "google",
	}, nil
}

// Facebook OAuth

func (h *OAuthHandler) FacebookLogin(ctx *gin.Context) {
	if !h.oauthConfig.Facebook.Enabled {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Facebook OAuth is not enabled"})
		return
	}

	state, err := generateState()
	if err != nil {
		h.log.Error("failed to generate state", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initiate OAuth"})
		return
	}

	ctx.SetCookie("oauth_state", state, 600, "/", h.cookieConfig.Domain, h.cookieConfig.Secure, true)

	redirectURL := fmt.Sprintf("%s/api/v1/auth/oauth/facebook/callback", h.oauthConfig.RedirectBaseURL)
	authURL := fmt.Sprintf(
		"https://www.facebook.com/v18.0/dialog/oauth?client_id=%s&redirect_uri=%s&scope=email&state=%s",
		url.QueryEscape(h.oauthConfig.Facebook.ClientID),
		url.QueryEscape(redirectURL),
		url.QueryEscape(state),
	)

	ctx.Redirect(http.StatusTemporaryRedirect, authURL)
}

func (h *OAuthHandler) FacebookCallback(ctx *gin.Context) {
	if !h.oauthConfig.Facebook.Enabled {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Facebook OAuth is not enabled"})
		return
	}

	state := ctx.Query("state")
	storedState, err := ctx.Cookie("oauth_state")
	if err != nil || state != storedState {
		h.log.Warn("oauth_callback", "provider", "facebook", "error", "invalid state")
		h.redirectWithError(ctx, "Invalid OAuth state")
		return
	}

	code := ctx.Query("code")
	if code == "" {
		h.redirectWithError(ctx, "No authorization code received")
		return
	}

	redirectURL := fmt.Sprintf("%s/api/v1/auth/oauth/facebook/callback", h.oauthConfig.RedirectBaseURL)
	accessToken, err := h.exchangeFacebookCode(code, redirectURL)
	if err != nil {
		h.log.Error("facebook_token_exchange", "error", err)
		h.redirectWithError(ctx, "Failed to authenticate with Facebook")
		return
	}

	userInfo, err := h.getFacebookUserInfo(accessToken)
	if err != nil {
		h.log.Error("facebook_userinfo", "error", err)
		h.redirectWithError(ctx, "Failed to get user info from Facebook")
		return
	}

	h.handleOAuthUser(ctx, userInfo)
}

func (h *OAuthHandler) exchangeFacebookCode(code, redirectURL string) (string, error) {
	reqURL := fmt.Sprintf(
		"https://graph.facebook.com/v18.0/oauth/access_token?client_id=%s&client_secret=%s&redirect_uri=%s&code=%s",
		url.QueryEscape(h.oauthConfig.Facebook.ClientID),
		url.QueryEscape(h.oauthConfig.Facebook.ClientSecret),
		url.QueryEscape(redirectURL),
		url.QueryEscape(code),
	)

	resp, err := http.Get(reqURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var data struct {
		AccessToken string `json:"access_token"`
		Error       struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", err
	}

	if data.Error.Message != "" {
		return "", fmt.Errorf("facebook error: %s", data.Error.Message)
	}

	return data.AccessToken, nil
}

func (h *OAuthHandler) getFacebookUserInfo(accessToken string) (*OAuthUserInfo, error) {
	reqURL := fmt.Sprintf(
		"https://graph.facebook.com/me?fields=id,email,first_name,last_name&access_token=%s",
		url.QueryEscape(accessToken),
	)

	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, err
	}

	return &OAuthUserInfo{
		ID:        data.ID,
		Email:     data.Email,
		FirstName: data.FirstName,
		LastName:  data.LastName,
		Provider:  "facebook",
	}, nil
}

// Apple OAuth

func (h *OAuthHandler) AppleLogin(ctx *gin.Context) {
	if !h.oauthConfig.Apple.Enabled {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Apple OAuth is not enabled"})
		return
	}

	state, err := generateState()
	if err != nil {
		h.log.Error("failed to generate state", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initiate OAuth"})
		return
	}

	ctx.SetCookie("oauth_state", state, 600, "/", h.cookieConfig.Domain, h.cookieConfig.Secure, true)

	redirectURL := fmt.Sprintf("%s/api/v1/auth/oauth/apple/callback", h.oauthConfig.RedirectBaseURL)
	authURL := fmt.Sprintf(
		"https://appleid.apple.com/auth/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=email%%20name&response_mode=form_post&state=%s",
		url.QueryEscape(h.oauthConfig.Apple.ClientID),
		url.QueryEscape(redirectURL),
		url.QueryEscape(state),
	)

	ctx.Redirect(http.StatusTemporaryRedirect, authURL)
}

func (h *OAuthHandler) AppleCallback(ctx *gin.Context) {
	if !h.oauthConfig.Apple.Enabled {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Apple OAuth is not enabled"})
		return
	}

	// Apple uses form_post response mode
	state := ctx.PostForm("state")
	if state == "" {
		state = ctx.Query("state")
	}

	storedState, err := ctx.Cookie("oauth_state")
	if err != nil || state != storedState {
		h.log.Warn("oauth_callback", "provider", "apple", "error", "invalid state")
		h.redirectWithError(ctx, "Invalid OAuth state")
		return
	}

	code := ctx.PostForm("code")
	if code == "" {
		code = ctx.Query("code")
	}
	if code == "" {
		h.redirectWithError(ctx, "No authorization code received")
		return
	}

	// Apple provides user info in the first callback only
	userJSON := ctx.PostForm("user")
	var appleUser struct {
		Name struct {
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
		} `json:"name"`
		Email string `json:"email"`
	}
	if userJSON != "" {
		json.Unmarshal([]byte(userJSON), &appleUser)
	}

	// Exchange code for tokens and get ID token claims
	userInfo, err := h.exchangeAppleCode(ctx.Request.Context(), code, appleUser.Name.FirstName, appleUser.Name.LastName)
	if err != nil {
		h.log.Error("apple_token_exchange", "error", err)
		h.redirectWithError(ctx, "Failed to authenticate with Apple")
		return
	}

	// Use email from user info if available (first login only)
	if appleUser.Email != "" {
		userInfo.Email = appleUser.Email
	}
	if appleUser.Name.FirstName != "" {
		userInfo.FirstName = appleUser.Name.FirstName
	}
	if appleUser.Name.LastName != "" {
		userInfo.LastName = appleUser.Name.LastName
	}

	h.handleOAuthUser(ctx, userInfo)
}

func (h *OAuthHandler) exchangeAppleCode(ctx context.Context, code, firstName, lastName string) (*OAuthUserInfo, error) {
	// For Apple, we need to generate a client secret JWT
	// This is a simplified version - production should use proper JWT signing

	redirectURL := fmt.Sprintf("%s/api/v1/auth/oauth/apple/callback", h.oauthConfig.RedirectBaseURL)

	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", h.oauthConfig.Apple.ClientID)
	data.Set("client_secret", h.generateAppleClientSecret())
	data.Set("redirect_uri", redirectURL)
	data.Set("grant_type", "authorization_code")

	resp, err := http.Post("https://appleid.apple.com/auth/token", "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var tokenResp struct {
		IDToken string `json:"id_token"`
		Error   string `json:"error"`
	}
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, err
	}

	if tokenResp.Error != "" {
		return nil, fmt.Errorf("apple error: %s", tokenResp.Error)
	}

	// Parse ID token to get user info (simplified - should verify signature in production)
	claims, err := parseJWTClaims(tokenResp.IDToken)
	if err != nil {
		return nil, err
	}

	email, _ := claims["email"].(string)
	sub, _ := claims["sub"].(string)

	return &OAuthUserInfo{
		ID:        sub,
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Provider:  "apple",
	}, nil
}

func (h *OAuthHandler) generateAppleClientSecret() string {
	// In production, this should generate a proper JWT signed with Apple's private key
	// For now, return the configured secret or empty string
	return h.oauthConfig.Apple.PrivateKey
}

// parseJWTClaims extracts claims from a JWT without verification (for demo purposes)
func parseJWTClaims(tokenString string) (map[string]interface{}, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}

	return claims, nil
}

// Common OAuth user handling

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
	h.setAuthCookie(ctx, token)

	// Redirect to frontend with success
	redirectURL := fmt.Sprintf("%s/oauth/callback?success=true", h.oauthConfig.RedirectBaseURL)
	ctx.Redirect(http.StatusTemporaryRedirect, redirectURL)
}

func (h *OAuthHandler) setAuthCookie(ctx *gin.Context, token string) {
	maxAge := h.cookieConfig.ExpiryHours * 3600
	ctx.SetSameSite(http.SameSiteLaxMode)
	ctx.SetCookie(
		cookieName,
		token,
		maxAge,
		"/",
		h.cookieConfig.Domain,
		h.cookieConfig.Secure,
		true,
	)
}

func (h *OAuthHandler) redirectWithError(ctx *gin.Context, errorMsg string) {
	redirectURL := fmt.Sprintf("%s/oauth/callback?error=%s", h.oauthConfig.RedirectBaseURL, url.QueryEscape(errorMsg))
	ctx.Redirect(http.StatusTemporaryRedirect, redirectURL)
}
