package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elprogramadorgt/lucidRAG/internal/config"
	userDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/user"
	"github.com/elprogramadorgt/lucidRAG/pkg/logger"
	"github.com/gin-gonic/gin"
)

// mockUserServiceOAuth is a mock implementation for OAuth testing
type mockUserServiceOAuth struct {
	getUserByEmailFunc func(ctx context.Context, email string) (*userDomain.User, error)
	registerOAuthFunc  func(ctx context.Context, newUser userDomain.User, provider, providerID string) (*userDomain.User, error)
	generateTokenFunc  func(user *userDomain.User) (string, error)
}

func (m *mockUserServiceOAuth) Register(ctx context.Context, newUser userDomain.User) (*userDomain.User, error) {
	return nil, nil
}

func (m *mockUserServiceOAuth) RegisterOAuth(ctx context.Context, newUser userDomain.User, provider, providerID string) (*userDomain.User, error) {
	if m.registerOAuthFunc != nil {
		return m.registerOAuthFunc(ctx, newUser, provider, providerID)
	}
	return &userDomain.User{
		ID:              "user-123",
		Email:           newUser.Email,
		FirstName:       newUser.FirstName,
		LastName:        newUser.LastName,
		Role:            userDomain.RoleUser,
		OAuthProvider:   provider,
		OAuthProviderID: providerID,
	}, nil
}

func (m *mockUserServiceOAuth) Login(ctx context.Context, email, password string) (string, *userDomain.User, error) {
	return "", nil, nil
}

func (m *mockUserServiceOAuth) GetUser(ctx context.Context, id string) (*userDomain.User, error) {
	return nil, nil
}

func (m *mockUserServiceOAuth) GetUserByEmail(ctx context.Context, email string) (*userDomain.User, error) {
	if m.getUserByEmailFunc != nil {
		return m.getUserByEmailFunc(ctx, email)
	}
	return nil, errors.New("user not found")
}

func (m *mockUserServiceOAuth) ValidateToken(token string) (*userDomain.Claims, error) {
	return nil, nil
}

func (m *mockUserServiceOAuth) GenerateToken(user *userDomain.User) (string, error) {
	if m.generateTokenFunc != nil {
		return m.generateTokenFunc(user)
	}
	return "mock-jwt-token", nil
}

func setupOAuthTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func createTestOAuthHandler(mockSvc *mockUserServiceOAuth) *OAuthHandler {
	log := logger.New(logger.Options{Level: "error"})
	return NewOAuthHandler(
		mockSvc,
		log,
		config.OAuthConfig{
			RedirectBaseURL: "http://localhost:4200",
			Google: config.OAuthProviderConfig{
				ClientID:     "google-client-id",
				ClientSecret: "google-client-secret",
				Enabled:      true,
			},
			Facebook: config.OAuthProviderConfig{
				ClientID:     "facebook-client-id",
				ClientSecret: "facebook-client-secret",
				Enabled:      true,
			},
			Apple: config.AppleOAuthConfig{
				ClientID:   "apple-client-id",
				TeamID:     "apple-team-id",
				KeyID:      "apple-key-id",
				PrivateKey: "apple-private-key",
				Enabled:    true,
			},
		},
		CookieConfig{
			Domain:      "localhost",
			Secure:      false,
			ExpiryHours: 24,
		},
	)
}

func TestGetProviders(t *testing.T) {
	mockSvc := &mockUserServiceOAuth{}
	handler := createTestOAuthHandler(mockSvc)

	router := setupOAuthTestRouter()
	router.GET("/providers", handler.GetProviders)

	req, _ := http.NewRequest("GET", "/providers", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.Code)
	}

	var result ProvidersResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if !result.Google {
		t.Error("Expected Google to be enabled")
	}
	if !result.Facebook {
		t.Error("Expected Facebook to be enabled")
	}
	if !result.Apple {
		t.Error("Expected Apple to be enabled")
	}
}

func TestGetProvidersDisabled(t *testing.T) {
	log := logger.New(logger.Options{Level: "error"})
	handler := NewOAuthHandler(
		&mockUserServiceOAuth{},
		log,
		config.OAuthConfig{
			RedirectBaseURL: "http://localhost:4200",
			Google:          config.OAuthProviderConfig{Enabled: false},
			Facebook:        config.OAuthProviderConfig{Enabled: false},
			Apple:           config.AppleOAuthConfig{Enabled: false},
		},
		CookieConfig{},
	)

	router := setupOAuthTestRouter()
	router.GET("/providers", handler.GetProviders)

	req, _ := http.NewRequest("GET", "/providers", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	var result ProvidersResponse
	json.Unmarshal(resp.Body.Bytes(), &result)

	if result.Google {
		t.Error("Expected Google to be disabled")
	}
	if result.Facebook {
		t.Error("Expected Facebook to be disabled")
	}
	if result.Apple {
		t.Error("Expected Apple to be disabled")
	}
}

func TestGoogleLoginDisabled(t *testing.T) {
	log := logger.New(logger.Options{Level: "error"})
	handler := NewOAuthHandler(
		&mockUserServiceOAuth{},
		log,
		config.OAuthConfig{
			Google: config.OAuthProviderConfig{Enabled: false},
		},
		CookieConfig{},
	)

	router := setupOAuthTestRouter()
	router.GET("/google", handler.GoogleLogin)

	req, _ := http.NewRequest("GET", "/google", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.Code)
	}
}

func TestGoogleLoginRedirect(t *testing.T) {
	mockSvc := &mockUserServiceOAuth{}
	handler := createTestOAuthHandler(mockSvc)

	router := setupOAuthTestRouter()
	router.GET("/google", handler.GoogleLogin)

	req, _ := http.NewRequest("GET", "/google", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected status 307, got %d", resp.Code)
	}

	location := resp.Header().Get("Location")
	if location == "" {
		t.Error("Expected Location header to be set")
	}

	// Check that it redirects to Google's OAuth endpoint
	if !contains(location, "accounts.google.com") {
		t.Errorf("Expected redirect to Google OAuth, got %s", location)
	}
}

func TestFacebookLoginDisabled(t *testing.T) {
	log := logger.New(logger.Options{Level: "error"})
	handler := NewOAuthHandler(
		&mockUserServiceOAuth{},
		log,
		config.OAuthConfig{
			Facebook: config.OAuthProviderConfig{Enabled: false},
		},
		CookieConfig{},
	)

	router := setupOAuthTestRouter()
	router.GET("/facebook", handler.FacebookLogin)

	req, _ := http.NewRequest("GET", "/facebook", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.Code)
	}
}

func TestFacebookLoginRedirect(t *testing.T) {
	mockSvc := &mockUserServiceOAuth{}
	handler := createTestOAuthHandler(mockSvc)

	router := setupOAuthTestRouter()
	router.GET("/facebook", handler.FacebookLogin)

	req, _ := http.NewRequest("GET", "/facebook", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected status 307, got %d", resp.Code)
	}

	location := resp.Header().Get("Location")
	if !contains(location, "facebook.com") {
		t.Errorf("Expected redirect to Facebook OAuth, got %s", location)
	}
}

func TestAppleLoginDisabled(t *testing.T) {
	log := logger.New(logger.Options{Level: "error"})
	handler := NewOAuthHandler(
		&mockUserServiceOAuth{},
		log,
		config.OAuthConfig{
			Apple: config.AppleOAuthConfig{Enabled: false},
		},
		CookieConfig{},
	)

	router := setupOAuthTestRouter()
	router.GET("/apple", handler.AppleLogin)

	req, _ := http.NewRequest("GET", "/apple", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.Code)
	}
}

func TestAppleLoginRedirect(t *testing.T) {
	mockSvc := &mockUserServiceOAuth{}
	handler := createTestOAuthHandler(mockSvc)

	router := setupOAuthTestRouter()
	router.GET("/apple", handler.AppleLogin)

	req, _ := http.NewRequest("GET", "/apple", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected status 307, got %d", resp.Code)
	}

	location := resp.Header().Get("Location")
	if !contains(location, "appleid.apple.com") {
		t.Errorf("Expected redirect to Apple OAuth, got %s", location)
	}
}

func TestGoogleCallbackInvalidState(t *testing.T) {
	mockSvc := &mockUserServiceOAuth{}
	handler := createTestOAuthHandler(mockSvc)

	router := setupOAuthTestRouter()
	router.GET("/callback", handler.GoogleCallback)

	req, _ := http.NewRequest("GET", "/callback?state=invalid&code=auth-code", nil)
	req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "different-state"})
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	// Should redirect with error
	if resp.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected status 307, got %d", resp.Code)
	}

	location := resp.Header().Get("Location")
	if !contains(location, "error=") {
		t.Errorf("Expected error in redirect URL, got %s", location)
	}
}

func TestGoogleCallbackNoCode(t *testing.T) {
	mockSvc := &mockUserServiceOAuth{}
	handler := createTestOAuthHandler(mockSvc)

	router := setupOAuthTestRouter()
	router.GET("/callback", handler.GoogleCallback)

	req, _ := http.NewRequest("GET", "/callback?state=test-state", nil)
	req.AddCookie(&http.Cookie{Name: "oauth_state", Value: "test-state"})
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusTemporaryRedirect {
		t.Errorf("Expected status 307, got %d", resp.Code)
	}

	location := resp.Header().Get("Location")
	if !contains(location, "error=") {
		t.Errorf("Expected error in redirect URL, got %s", location)
	}
}

func TestParseJWTClaims(t *testing.T) {
	// Create a test JWT payload
	payload := map[string]interface{}{
		"sub":   "user-123",
		"email": "test@example.com",
	}
	payloadBytes, _ := json.Marshal(payload)
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadBytes)

	// Create a mock JWT (header.payload.signature)
	mockJWT := "eyJhbGciOiJSUzI1NiJ9." + encodedPayload + ".mock-signature"

	claims, err := parseJWTClaims(mockJWT)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if claims["sub"] != "user-123" {
		t.Errorf("Expected sub user-123, got %v", claims["sub"])
	}
	if claims["email"] != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %v", claims["email"])
	}
}

func TestParseJWTClaimsInvalidFormat(t *testing.T) {
	_, err := parseJWTClaims("invalid-jwt")
	if err == nil {
		t.Error("Expected error for invalid JWT format")
	}
}

func TestParseJWTClaimsInvalidBase64(t *testing.T) {
	_, err := parseJWTClaims("header.!!!invalid-base64!!!.signature")
	if err == nil {
		t.Error("Expected error for invalid base64")
	}
}

func TestGenerateState(t *testing.T) {
	state1, err := generateState()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if state1 == "" {
		t.Error("Expected non-empty state")
	}

	state2, err := generateState()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// States should be unique
	if state1 == state2 {
		t.Error("Expected unique states")
	}

	// State should be base64 encoded
	_, err = base64.URLEncoding.DecodeString(state1)
	if err != nil {
		t.Errorf("Expected valid base64, got error: %v", err)
	}
}

func TestOAuthUserInfoStruct(t *testing.T) {
	userInfo := OAuthUserInfo{
		ID:        "123",
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Provider:  "google",
	}

	if userInfo.ID != "123" {
		t.Errorf("Expected ID 123, got %s", userInfo.ID)
	}
	if userInfo.Email != "test@example.com" {
		t.Errorf("Expected email test@example.com, got %s", userInfo.Email)
	}
	if userInfo.Provider != "google" {
		t.Errorf("Expected provider google, got %s", userInfo.Provider)
	}
}

func TestProvidersResponseStruct(t *testing.T) {
	resp := ProvidersResponse{
		Google:   true,
		Facebook: false,
		Apple:    true,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var parsed ProvidersResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if !parsed.Google {
		t.Error("Expected Google to be true")
	}
	if parsed.Facebook {
		t.Error("Expected Facebook to be false")
	}
	if !parsed.Apple {
		t.Error("Expected Apple to be true")
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
