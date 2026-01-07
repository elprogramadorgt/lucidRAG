package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	ragApp "github.com/elprogramadorgt/lucidRAG/internal/application/rag"
	ragDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/rag"
	"github.com/elprogramadorgt/lucidRAG/pkg/logger"
	"github.com/gin-gonic/gin"
)

type mockRAGService struct {
	response *ragDomain.Response
	err      error
}

func (m *mockRAGService) Query(_ context.Context, _ ragDomain.Query) (*ragDomain.Response, error) {
	return m.response, m.err
}

func (m *mockRAGService) IndexDocument(_ context.Context, _ string, _ string) error {
	return nil
}

func (m *mockRAGService) DeleteDocumentChunks(_ context.Context, _ string) error {
	return nil
}

func setupTest() (*gin.Engine, *logger.Logger) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	log := logger.New(logger.Options{Level: "error"})
	return r, log
}

func TestNewHandler(t *testing.T) {
	_, log := setupTest()
	svc := &mockRAGService{}

	h := NewHandler(svc, log)
	if h == nil {
		t.Fatal("Expected handler to be created")
	}
}

func TestQuerySuccess(t *testing.T) {
	r, log := setupTest()
	svc := &mockRAGService{
		response: &ragDomain.Response{
			Answer:          "Test answer",
			ConfidenceScore: 0.95,
		},
	}

	h := NewHandler(svc, log)
	r.POST("/query", h.Query)

	body := `{"query": "test question"}`
	req := httptest.NewRequest(http.MethodPost, "/query", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var resp ragDomain.Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if resp.Answer != "Test answer" {
		t.Errorf("Expected answer 'Test answer', got '%s'", resp.Answer)
	}
}

func TestQueryInvalidBody(t *testing.T) {
	r, log := setupTest()
	svc := &mockRAGService{}

	h := NewHandler(svc, log)
	r.POST("/query", h.Query)

	req := httptest.NewRequest(http.MethodPost, "/query", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestQueryMissingQuery(t *testing.T) {
	r, log := setupTest()
	svc := &mockRAGService{}

	h := NewHandler(svc, log)
	r.POST("/query", h.Query)

	body := `{"top_k": 5}`
	req := httptest.NewRequest(http.MethodPost, "/query", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestQueryInvalidQueryError(t *testing.T) {
	r, log := setupTest()
	svc := &mockRAGService{
		err: ragApp.ErrInvalidQuery,
	}

	h := NewHandler(svc, log)
	r.POST("/query", h.Query)

	body := `{"query": "test"}`
	req := httptest.NewRequest(http.MethodPost, "/query", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestQueryServiceError(t *testing.T) {
	r, log := setupTest()
	svc := &mockRAGService{
		err: errors.New("service error"),
	}

	h := NewHandler(svc, log)
	r.POST("/query", h.Query)

	body := `{"query": "test question"}`
	req := httptest.NewRequest(http.MethodPost, "/query", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestQueryWithOptionalParams(t *testing.T) {
	r, log := setupTest()
	svc := &mockRAGService{
		response: &ragDomain.Response{
			Answer: "Answer with params",
		},
	}

	h := NewHandler(svc, log)
	r.POST("/query", h.Query)

	body := `{"query": "test", "top_k": 10, "threshold": 0.8}`
	req := httptest.NewRequest(http.MethodPost, "/query", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}
