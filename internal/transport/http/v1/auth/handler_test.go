package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	userApp "github.com/elprogramadorgt/lucidRAG/internal/application/user"
	userDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/user"
	"github.com/elprogramadorgt/lucidRAG/pkg/logger"
	"github.com/gin-gonic/gin"
)

// mockUserServiceHandler is a mock implementation for handler testing
type mockUserServiceHandler struct {
	registerFunc  func(ctx context.Context, newUser userDomain.User) (*userDomain.User, error)
	loginFunc     func(ctx context.Context, email, password string) (string, *userDomain.User, error)
	getUserFunc   func(ctx context.Context, id string) (*userDomain.User, error)
}

func (m *mockUserServiceHandler) Register(ctx context.Context, newUser userDomain.User) (*userDomain.User, error) {
	if m.registerFunc != nil {
		return m.registerFunc(ctx, newUser)
	}
	return &userDomain.User{
		ID:        "user-123",
		Email:     newUser.Email,
		FirstName: newUser.FirstName,
		LastName:  newUser.LastName,
		Role:      userDomain.RoleUser,
	}, nil
}

func (m *mockUserServiceHandler) RegisterOAuth(ctx context.Context, newUser userDomain.User, provider, providerID string) (*userDomain.User, error) {
	return nil, nil
}

func (m *mockUserServiceHandler) Login(ctx context.Context, email, password string) (string, *userDomain.User, error) {
	if m.loginFunc != nil {
		return m.loginFunc(ctx, email, password)
	}
	return "mock-token", &userDomain.User{
		ID:    "user-123",
		Email: email,
		Role:  userDomain.RoleUser,
	}, nil
}

func (m *mockUserServiceHandler) GetUser(ctx context.Context, id string) (*userDomain.User, error) {
	if m.getUserFunc != nil {
		return m.getUserFunc(ctx, id)
	}
	return &userDomain.User{
		ID:    id,
		Email: "test@example.com",
		Role:  userDomain.RoleUser,
	}, nil
}

func (m *mockUserServiceHandler) GetUserByEmail(ctx context.Context, email string) (*userDomain.User, error) {
	return nil, nil
}

func (m *mockUserServiceHandler) ValidateToken(token string) (*userDomain.Claims, error) {
	return nil, nil
}

func (m *mockUserServiceHandler) GenerateToken(user *userDomain.User) (string, error) {
	return "mock-token", nil
}

func setupHandlerTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func createTestHandler(mockSvc *mockUserServiceHandler) *Handler {
	log := logger.New(logger.Options{Level: "error"})
	return NewHandler(
		mockSvc,
		log,
		CookieConfig{
			Domain:      "localhost",
			Secure:      false,
			ExpiryHours: 24,
		},
	)
}

func TestRegisterSuccess(t *testing.T) {
	mockSvc := &mockUserServiceHandler{}
	handler := createTestHandler(mockSvc)

	router := setupHandlerTestRouter()
	router.POST("/register", handler.Register)

	body := `{"email":"test@example.com","password":"password123","first_name":"Test","last_name":"User"}`
	req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.Code)
	}

	var result authResponse
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if result.User == nil {
		t.Error("Expected user in response")
	}
}

func TestRegisterInvalidRequest(t *testing.T) {
	mockSvc := &mockUserServiceHandler{}
	handler := createTestHandler(mockSvc)

	router := setupHandlerTestRouter()
	router.POST("/register", handler.Register)

	// Missing required fields
	body := `{"email":"invalid-email"}`
	req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.Code)
	}
}

func TestRegisterEmailExists(t *testing.T) {
	mockSvc := &mockUserServiceHandler{
		registerFunc: func(ctx context.Context, newUser userDomain.User) (*userDomain.User, error) {
			return nil, userApp.ErrEmailExists
		},
	}
	handler := createTestHandler(mockSvc)

	router := setupHandlerTestRouter()
	router.POST("/register", handler.Register)

	body := `{"email":"existing@example.com","password":"password123","first_name":"Test","last_name":"User"}`
	req, _ := http.NewRequest("POST", "/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusConflict {
		t.Errorf("Expected status 409, got %d", resp.Code)
	}
}

func TestLoginSuccess(t *testing.T) {
	mockSvc := &mockUserServiceHandler{}
	handler := createTestHandler(mockSvc)

	router := setupHandlerTestRouter()
	router.POST("/login", handler.Login)

	body := `{"email":"test@example.com","password":"password123"}`
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.Code)
	}

	// Check that cookie was set
	cookies := resp.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == cookieName {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected auth cookie to be set")
	}
}

func TestLoginInvalidRequest(t *testing.T) {
	mockSvc := &mockUserServiceHandler{}
	handler := createTestHandler(mockSvc)

	router := setupHandlerTestRouter()
	router.POST("/login", handler.Login)

	body := `{"email":"invalid-email"}`
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.Code)
	}
}

func TestLoginInvalidCredentials(t *testing.T) {
	mockSvc := &mockUserServiceHandler{
		loginFunc: func(ctx context.Context, email, password string) (string, *userDomain.User, error) {
			return "", nil, userApp.ErrInvalidCredentials
		},
	}
	handler := createTestHandler(mockSvc)

	router := setupHandlerTestRouter()
	router.POST("/login", handler.Login)

	body := `{"email":"test@example.com","password":"wrongpassword"}`
	req, _ := http.NewRequest("POST", "/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.Code)
	}
}

func TestLogout(t *testing.T) {
	mockSvc := &mockUserServiceHandler{}
	handler := createTestHandler(mockSvc)

	router := setupHandlerTestRouter()
	router.POST("/logout", handler.Logout)

	req, _ := http.NewRequest("POST", "/logout", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.Code)
	}

	// Check that cookie was cleared (MaxAge = -1)
	cookies := resp.Result().Cookies()
	for _, c := range cookies {
		if c.Name == cookieName {
			if c.MaxAge >= 0 {
				t.Error("Expected cookie to be cleared (MaxAge < 0)")
			}
			break
		}
	}
}

func TestMeSuccess(t *testing.T) {
	mockSvc := &mockUserServiceHandler{}
	handler := createTestHandler(mockSvc)

	router := setupHandlerTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Next()
	})
	router.GET("/me", handler.Me)

	req, _ := http.NewRequest("GET", "/me", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.Code)
	}
}

func TestMeUnauthorized(t *testing.T) {
	mockSvc := &mockUserServiceHandler{}
	handler := createTestHandler(mockSvc)

	router := setupHandlerTestRouter()
	router.GET("/me", handler.Me)

	req, _ := http.NewRequest("GET", "/me", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", resp.Code)
	}
}

func TestMeUserNotFound(t *testing.T) {
	mockSvc := &mockUserServiceHandler{
		getUserFunc: func(ctx context.Context, id string) (*userDomain.User, error) {
			return nil, userApp.ErrUserNotFound
		},
	}
	handler := createTestHandler(mockSvc)

	router := setupHandlerTestRouter()
	router.Use(func(c *gin.Context) {
		c.Set("user_id", "non-existent-user")
		c.Next()
	})
	router.GET("/me", handler.Me)

	req, _ := http.NewRequest("GET", "/me", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.Code)
	}
}

func TestCookieConfig(t *testing.T) {
	cfg := CookieConfig{
		Domain:      "example.com",
		Secure:      true,
		ExpiryHours: 48,
	}

	if cfg.Domain != "example.com" {
		t.Errorf("Expected domain example.com, got %s", cfg.Domain)
	}
	if !cfg.Secure {
		t.Error("Expected secure to be true")
	}
	if cfg.ExpiryHours != 48 {
		t.Errorf("Expected expiry hours 48, got %d", cfg.ExpiryHours)
	}
}

func TestNewHandler(t *testing.T) {
	mockSvc := &mockUserServiceHandler{}
	log := logger.New(logger.Options{Level: "error"})
	handler := NewHandler(mockSvc, log, CookieConfig{})

	if handler == nil {
		t.Fatal("Expected handler to be created")
	}
}
