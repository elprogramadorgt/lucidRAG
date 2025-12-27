package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	docApp "github.com/elprogramadorgt/lucidRAG/internal/application/document"
	documentDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/document"
	"github.com/elprogramadorgt/lucidRAG/pkg/logger"
	"github.com/gin-gonic/gin"
)

type mockDocumentService struct {
	queryRAGFunc func(ctx context.Context, query documentDomain.RAGQuery) (*documentDomain.RAGResponse, error)
}

func (m *mockDocumentService) CreateDocument(ctx context.Context, userCtx documentDomain.UserContext, doc *documentDomain.Document) (string, error) {
	return "", nil
}

func (m *mockDocumentService) GetDocument(ctx context.Context, userCtx documentDomain.UserContext, id string) (*documentDomain.Document, error) {
	return nil, nil
}

func (m *mockDocumentService) ListDocuments(ctx context.Context, userCtx documentDomain.UserContext, limit, offset int) ([]documentDomain.Document, int64, error) {
	return nil, 0, nil
}

func (m *mockDocumentService) UpdateDocument(ctx context.Context, userCtx documentDomain.UserContext, doc *documentDomain.Document) error {
	return nil
}

func (m *mockDocumentService) DeleteDocument(ctx context.Context, userCtx documentDomain.UserContext, id string) error {
	return nil
}

func (m *mockDocumentService) QueryRAG(ctx context.Context, query documentDomain.RAGQuery) (*documentDomain.RAGResponse, error) {
	if m.queryRAGFunc != nil {
		return m.queryRAGFunc(ctx, query)
	}
	return nil, nil
}

func setupTestRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/query", handler.Query)
	return r
}

func createTestLogger() *logger.Logger {
	return logger.New(logger.Options{Level: "error"})
}

func TestNewHandler(t *testing.T) {
	log := createTestLogger()
	svc := &mockDocumentService{}

	handler := NewHandler(svc, log)

	if handler == nil {
		t.Fatal("NewHandler returned nil")
	}
	if handler.svc == nil {
		t.Error("Handler service is nil")
	}
	if handler.log == nil {
		t.Error("Handler logger is nil")
	}
}

func TestQuery_Success(t *testing.T) {
	log := createTestLogger()
	expectedResponse := &documentDomain.RAGResponse{
		Answer: "Test answer",
		RelevantChunks: []documentDomain.Chunk{
			{ID: "chunk1", Content: "chunk content"},
		},
		ConfidenceScore:  0.95,
		ProcessingTimeMs: 100,
	}

	svc := &mockDocumentService{
		queryRAGFunc: func(ctx context.Context, query documentDomain.RAGQuery) (*documentDomain.RAGResponse, error) {
			if query.Query != "test query" {
				t.Errorf("Expected query 'test query', got '%s'", query.Query)
			}
			if query.TopK != 5 {
				t.Errorf("Expected TopK 5, got %d", query.TopK)
			}
			if query.Threshold != 0.5 {
				t.Errorf("Expected Threshold 0.5, got %f", query.Threshold)
			}
			return expectedResponse, nil
		},
	}

	handler := NewHandler(svc, log)
	router := setupTestRouter(handler)

	reqBody := queryRequest{
		Query:     "test query",
		TopK:      5,
		Threshold: 0.5,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/query", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	var response documentDomain.RAGResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Answer != expectedResponse.Answer {
		t.Errorf("Expected answer %q, got %q", expectedResponse.Answer, response.Answer)
	}
	if response.ConfidenceScore != expectedResponse.ConfidenceScore {
		t.Errorf("Expected confidence %f, got %f", expectedResponse.ConfidenceScore, response.ConfidenceScore)
	}
}

func TestQuery_InvalidJSON(t *testing.T) {
	log := createTestLogger()
	svc := &mockDocumentService{}

	handler := NewHandler(svc, log)
	router := setupTestRouter(handler)

	req := httptest.NewRequest(http.MethodPost, "/query", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	if response["error"] != "invalid request body" {
		t.Errorf("Expected error 'invalid request body', got %q", response["error"])
	}
}

func TestQuery_MissingRequiredField(t *testing.T) {
	log := createTestLogger()
	svc := &mockDocumentService{}

	handler := NewHandler(svc, log)
	router := setupTestRouter(handler)

	reqBody := map[string]interface{}{
		"top_k": 5,
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/query", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestQuery_InvalidQueryError(t *testing.T) {
	log := createTestLogger()
	svc := &mockDocumentService{
		queryRAGFunc: func(ctx context.Context, query documentDomain.RAGQuery) (*documentDomain.RAGResponse, error) {
			return nil, docApp.ErrInvalidQuery
		},
	}

	handler := NewHandler(svc, log)
	router := setupTestRouter(handler)

	reqBody := queryRequest{
		Query: "test query",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/query", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	if response["error"] != "invalid query" {
		t.Errorf("Expected error 'invalid query', got %q", response["error"])
	}
}

func TestQuery_InternalError(t *testing.T) {
	log := createTestLogger()
	svc := &mockDocumentService{
		queryRAGFunc: func(ctx context.Context, query documentDomain.RAGQuery) (*documentDomain.RAGResponse, error) {
			return nil, errors.New("database connection failed")
		},
	}

	handler := NewHandler(svc, log)
	router := setupTestRouter(handler)

	reqBody := queryRequest{
		Query: "test query",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/query", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", http.StatusInternalServerError, w.Code)
	}

	var response map[string]string
	json.Unmarshal(w.Body.Bytes(), &response)
	if response["error"] != "failed to process query" {
		t.Errorf("Expected error 'failed to process query', got %q", response["error"])
	}
}

func TestQuery_DefaultValues(t *testing.T) {
	log := createTestLogger()

	var capturedQuery documentDomain.RAGQuery
	svc := &mockDocumentService{
		queryRAGFunc: func(ctx context.Context, query documentDomain.RAGQuery) (*documentDomain.RAGResponse, error) {
			capturedQuery = query
			return &documentDomain.RAGResponse{}, nil
		},
	}

	handler := NewHandler(svc, log)
	router := setupTestRouter(handler)

	reqBody := queryRequest{
		Query: "test query",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/query", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	if capturedQuery.TopK != 0 {
		t.Errorf("Expected default TopK 0, got %d", capturedQuery.TopK)
	}
	if capturedQuery.Threshold != 0 {
		t.Errorf("Expected default Threshold 0, got %f", capturedQuery.Threshold)
	}
}

func TestQuery_EmptyBody(t *testing.T) {
	log := createTestLogger()
	svc := &mockDocumentService{}

	handler := NewHandler(svc, log)
	router := setupTestRouter(handler)

	req := httptest.NewRequest(http.MethodPost, "/query", bytes.NewReader([]byte{}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}
