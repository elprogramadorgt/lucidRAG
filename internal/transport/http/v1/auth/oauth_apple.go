package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
)

// AppleLogin initiates Apple OAuth flow.
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

// AppleCallback handles the Apple OAuth callback.
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
		_ = json.Unmarshal([]byte(userJSON), &appleUser)
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

func (h *OAuthHandler) exchangeAppleCode(_ context.Context, code, firstName, lastName string) (*OAuthUserInfo, error) {
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

// parseJWTClaims extracts claims from a JWT without verification (for demo purposes).
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
