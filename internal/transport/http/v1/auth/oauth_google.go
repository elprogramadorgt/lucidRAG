package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// GoogleLogin initiates Google OAuth flow.
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

// GoogleCallback handles the Google OAuth callback.
func (h *OAuthHandler) GoogleCallback(ctx *gin.Context) {
	if !h.oauthConfig.Google.Enabled {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Google OAuth is not enabled"})
		return
	}

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

	redirectURL := fmt.Sprintf("%s/api/v1/auth/oauth/google/callback", h.oauthConfig.RedirectBaseURL)
	tokenResp, err := h.exchangeGoogleCode(code, redirectURL)
	if err != nil {
		h.log.Error("google_token_exchange", "error", err)
		h.redirectWithError(ctx, "Failed to authenticate with Google")
		return
	}

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
		ID         string `json:"id"`
		Email      string `json:"email"`
		GivenName  string `json:"given_name"`
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
