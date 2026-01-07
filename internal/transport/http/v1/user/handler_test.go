package user

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	userDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/user"
	"github.com/gin-gonic/gin"
)

type mockUserService struct {
	user *userDomain.User
	err  error
}

func (m *mockUserService) Register(_ context.Context, _ userDomain.User) (*userDomain.User, error) {
	return m.user, m.err
}

func (m *mockUserService) RegisterOAuth(_ context.Context, _ userDomain.User, _, _ string) (*userDomain.User, error) {
	return m.user, m.err
}

func (m *mockUserService) Login(_ context.Context, _, _ string) (string, *userDomain.User, error) {
	return "", m.user, m.err
}

func (m *mockUserService) GetUser(_ context.Context, _ string) (*userDomain.User, error) {
	return m.user, m.err
}

func (m *mockUserService) GetUserByEmail(_ context.Context, _ string) (*userDomain.User, error) {
	return m.user, m.err
}

func (m *mockUserService) ValidateToken(_ string) (*userDomain.Claims, error) {
	return nil, m.err
}

func (m *mockUserService) GenerateToken(_ *userDomain.User) (string, error) {
	return "", m.err
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestNewHandler(t *testing.T) {
	svc := &mockUserService{}
	h := NewHandler(svc)

	if h == nil {
		t.Fatal("Expected handler to be created")
	}
}

func TestHandleRegisterSuccess(t *testing.T) {
	r := setupRouter()
	svc := &mockUserService{
		user: &userDomain.User{
			ID:    "user-123",
			Email: "test@example.com",
		},
	}

	h := NewHandler(svc)
	r.POST("/register", h.HandleRegister)

	body := `{"email": "test@example.com", "password": "secret123", "first_name": "John", "last_name": "Doe"}`
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}
}

func TestHandleRegisterInvalidJSON(t *testing.T) {
	r := setupRouter()
	svc := &mockUserService{}

	h := NewHandler(svc)
	r.POST("/register", h.HandleRegister)

	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleRegisterMissingFields(t *testing.T) {
	r := setupRouter()
	svc := &mockUserService{}

	h := NewHandler(svc)
	r.POST("/register", h.HandleRegister)

	body := `{"email": "test@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleRegisterInvalidEmail(t *testing.T) {
	r := setupRouter()
	svc := &mockUserService{}

	h := NewHandler(svc)
	r.POST("/register", h.HandleRegister)

	body := `{"email": "invalid-email", "password": "secret123", "first_name": "John", "last_name": "Doe"}`
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestHandleRegisterServiceError(t *testing.T) {
	r := setupRouter()
	svc := &mockUserService{
		err: errors.New("registration failed"),
	}

	h := NewHandler(svc)
	r.POST("/register", h.HandleRegister)

	body := `{"email": "test@example.com", "password": "secret123", "first_name": "John", "last_name": "Doe"}`
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestHandlerZeroValue(t *testing.T) {
	var h Handler

	if h.svc != nil {
		t.Error("Expected nil service in zero value handler")
	}
}
