package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type mockDBPinger struct {
	err error
}

func (m *mockDBPinger) Ping(_ context.Context) error {
	return m.err
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestLiveness(t *testing.T) {
	r := setupTestRouter()
	h := NewHandler(&mockDBPinger{})
	Register(r, h)

	req, _ := http.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != `{"status":"ok"}` {
		t.Errorf("Expected body {\"status\":\"ok\"}, got %s", w.Body.String())
	}
}

func TestReadinessHealthy(t *testing.T) {
	r := setupTestRouter()
	h := NewHandler(&mockDBPinger{err: nil})
	Register(r, h)

	req, _ := http.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != `{"status":"ok"}` {
		t.Errorf("Expected body {\"status\":\"ok\"}, got %s", w.Body.String())
	}
}

func TestReadinessUnhealthy(t *testing.T) {
	r := setupTestRouter()
	h := NewHandler(&mockDBPinger{err: errors.New("connection refused")})
	Register(r, h)

	req, _ := http.NewRequest(http.MethodGet, "/readyz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status %d, got %d", http.StatusServiceUnavailable, w.Code)
	}

	if w.Body.String() != `{"status":"error"}` {
		t.Errorf("Expected body {\"status\":\"error\"}, got %s", w.Body.String())
	}
}

func TestNewHandler(t *testing.T) {
	db := &mockDBPinger{}
	h := NewHandler(db)

	if h == nil {
		t.Fatal("Expected handler to be created")
	}
	if h.db != db {
		t.Error("Expected db to be set")
	}
}
