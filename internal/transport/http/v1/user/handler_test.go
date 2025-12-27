package user

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	userDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/user"
	"github.com/elprogramadorgt/lucidRAG/internal/transport/http/v1/user/dto"
	"github.com/gin-gonic/gin"
)

type mockUserService struct {
	registerFunc      func(ctx context.Context, newUser userDomain.User) (*userDomain.User, error)
	registerOAuthFunc func(ctx context.Context, newUser userDomain.User, provider, providerID string) (*userDomain.User, error)
	loginFunc         func(ctx context.Context, email, password string) (string, *userDomain.User, error)
	getUserFunc       func(ctx context.Context, id string) (*userDomain.User, error)
	getUserByEmailFunc func(ctx context.Context, email string) (*userDomain.User, error)
	validateTokenFunc func(token string) (*userDomain.Claims, error)
	generateTokenFunc func(user *userDomain.User) (string, error)
}

func (m *mockUserService) Register(ctx context.Context, newUser userDomain.User) (*userDomain.User, error) {
	if m.registerFunc != nil {
		return m.registerFunc(ctx, newUser)
	}
	return &newUser, nil
}

func (m *mockUserService) RegisterOAuth(ctx context.Context, newUser userDomain.User, provider, providerID string) (*userDomain.User, error) {
	if m.registerOAuthFunc != nil {
		return m.registerOAuthFunc(ctx, newUser, provider, providerID)
	}
	return &newUser, nil
}

func (m *mockUserService) Login(ctx context.Context, email, password string) (string, *userDomain.User, error) {
	if m.loginFunc != nil {
		return m.loginFunc(ctx, email, password)
	}
	return "token", nil, nil
}

func (m *mockUserService) GetUser(ctx context.Context, id string) (*userDomain.User, error) {
	if m.getUserFunc != nil {
		return m.getUserFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockUserService) GetUserByEmail(ctx context.Context, email string) (*userDomain.User, error) {
	if m.getUserByEmailFunc != nil {
		return m.getUserByEmailFunc(ctx, email)
	}
	return nil, nil
}

func (m *mockUserService) ValidateToken(token string) (*userDomain.Claims, error) {
	if m.validateTokenFunc != nil {
		return m.validateTokenFunc(token)
	}
	return nil, nil
}

func (m *mockUserService) GenerateToken(user *userDomain.User) (string, error) {
	if m.generateTokenFunc != nil {
		return m.generateTokenFunc(user)
	}
	return "token", nil
}

func setupTestRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/register", handler.HandleRegister)
	r.POST("/login", handler.HandleLogin)
	return r
}

func TestNewHandler(t *testing.T) {
	svc := &mockUserService{}
	handler := NewHandler(svc)

	if handler == nil {
		t.Fatal("NewHandler returned nil")
	}
	if handler.svc == nil {
		t.Error("Handler service is nil")
	}
}

func TestHandleRegister_Success(t *testing.T) {
	var capturedUser userDomain.User
	svc := &mockUserService{
		registerFunc: func(ctx context.Context, newUser userDomain.User) (*userDomain.User, error) {
			capturedUser = newUser
			return &newUser, nil
		},
	}

	handler := NewHandler(svc)
	router := setupTestRouter(handler)

	reqBody := dto.RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	if response["message"] != "user registered successfully" {
		t.Errorf("Expected message 'user registered successfully', got %q", response["message"])
	}

	if capturedUser.Email != reqBody.Email {
		t.Errorf("Expected email %q, got %q", reqBody.Email, capturedUser.Email)
	}
	if capturedUser.PasswordHash != reqBody.Password {
		t.Errorf("Expected password %q, got %q", reqBody.Password, capturedUser.PasswordHash)
	}
	if capturedUser.FirstName != reqBody.FirstName {
		t.Errorf("Expected first name %q, got %q", reqBody.FirstName, capturedUser.FirstName)
	}
	if capturedUser.LastName != reqBody.LastName {
		t.Errorf("Expected last name %q, got %q", reqBody.LastName, capturedUser.LastName)
	}
}

func TestHandleRegister_InvalidJSON(t *testing.T) {
	svc := &mockUserService{}
	handler := NewHandler(svc)
	router := setupTestRouter(handler)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	if response["error"] != "invalid request" {
		t.Errorf("Expected error 'invalid request', got %q", response["error"])
	}
}

func TestHandleRegister_MissingEmail(t *testing.T) {
	svc := &mockUserService{}
	handler := NewHandler(svc)
	router := setupTestRouter(handler)

	reqBody := map[string]string{
		"password":   "password123",
		"first_name": "John",
		"last_name":  "Doe",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleRegister_MissingPassword(t *testing.T) {
	svc := &mockUserService{}
	handler := NewHandler(svc)
	router := setupTestRouter(handler)

	reqBody := map[string]string{
		"email":      "test@example.com",
		"first_name": "John",
		"last_name":  "Doe",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleRegister_InvalidEmail(t *testing.T) {
	svc := &mockUserService{}
	handler := NewHandler(svc)
	router := setupTestRouter(handler)

	reqBody := map[string]string{
		"email":      "not-an-email",
		"password":   "password123",
		"first_name": "John",
		"last_name":  "Doe",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleRegister_ServiceError(t *testing.T) {
	svc := &mockUserService{
		registerFunc: func(ctx context.Context, newUser userDomain.User) (*userDomain.User, error) {
			return nil, errors.New("database error")
		},
	}

	handler := NewHandler(svc)
	router := setupTestRouter(handler)

	reqBody := dto.RegisterRequest{
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "John",
		LastName:  "Doe",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	if response["error"] != "failed to register user" {
		t.Errorf("Expected error 'failed to register user', got %q", response["error"])
	}
}

func TestHandleRegister_EmptyBody(t *testing.T) {
	svc := &mockUserService{}
	handler := NewHandler(svc)
	router := setupTestRouter(handler)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader([]byte{}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandleLogin_NotImplemented(t *testing.T) {
	svc := &mockUserService{}
	handler := NewHandler(svc)
	router := setupTestRouter(handler)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// HandleLogin is currently a stub, so it returns 200 with no body
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}
