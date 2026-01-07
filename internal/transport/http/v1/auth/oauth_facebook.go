package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

// FacebookLogin initiates Facebook OAuth flow.
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

// FacebookCallback handles the Facebook OAuth callback.
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
