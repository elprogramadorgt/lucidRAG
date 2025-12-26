package document

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	docApp "github.com/elprogramadorgt/lucidRAG/internal/application/document"
	docDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/document"
	"github.com/elprogramadorgt/lucidRAG/pkg/logger"
	"github.com/gin-gonic/gin"
)

type mockDocumentService struct {
	listDocumentsFunc  func(ctx context.Context, userCtx docDomain.UserContext, limit, offset int) ([]docDomain.Document, int64, error)
	getDocumentFunc    func(ctx context.Context, userCtx docDomain.UserContext, id string) (*docDomain.Document, error)
	createDocumentFunc func(ctx context.Context, userCtx docDomain.UserContext, doc *docDomain.Document) (string, error)
	updateDocumentFunc func(ctx context.Context, userCtx docDomain.UserContext, doc *docDomain.Document) error
	deleteDocumentFunc func(ctx context.Context, userCtx docDomain.UserContext, id string) error
}

func (m *mockDocumentService) ListDocuments(ctx context.Context, userCtx docDomain.UserContext, limit, offset int) ([]docDomain.Document, int64, error) {
	if m.listDocumentsFunc != nil {
		return m.listDocumentsFunc(ctx, userCtx, limit, offset)
	}
	return []docDomain.Document{}, 0, nil
}

func (m *mockDocumentService) GetDocument(ctx context.Context, userCtx docDomain.UserContext, id string) (*docDomain.Document, error) {
	if m.getDocumentFunc != nil {
		return m.getDocumentFunc(ctx, userCtx, id)
	}
	return nil, errors.New("not found")
}

func (m *mockDocumentService) CreateDocument(ctx context.Context, userCtx docDomain.UserContext, doc *docDomain.Document) (string, error) {
	if m.createDocumentFunc != nil {
		return m.createDocumentFunc(ctx, userCtx, doc)
	}
	return "doc-123", nil
}

func (m *mockDocumentService) UpdateDocument(ctx context.Context, userCtx docDomain.UserContext, doc *docDomain.Document) error {
	if m.updateDocumentFunc != nil {
		return m.updateDocumentFunc(ctx, userCtx, doc)
	}
	return nil
}

func (m *mockDocumentService) DeleteDocument(ctx context.Context, userCtx docDomain.UserContext, id string) error {
	if m.deleteDocumentFunc != nil {
		return m.deleteDocumentFunc(ctx, userCtx, id)
	}
	return nil
}

func (m *mockDocumentService) QueryRAG(ctx context.Context, query docDomain.RAGQuery) (*docDomain.RAGResponse, error) {
	return nil, nil
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func createTestHandler(mockSvc *mockDocumentService) *Handler {
	log := logger.New(logger.Options{Level: "error"})
	return NewHandler(mockSvc, log)
}

func TestListDocuments(t *testing.T) {
	mockSvc := &mockDocumentService{
		listDocumentsFunc: func(ctx context.Context, userCtx docDomain.UserContext, limit, offset int) ([]docDomain.Document, int64, error) {
			return []docDomain.Document{
				{ID: "doc-1", Title: "Document 1"},
				{ID: "doc-2", Title: "Document 2"},
			}, 2, nil
		},
	}
	handler := createTestHandler(mockSvc)

	router := setupTestRouter()
	router.GET("/documents", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Set("user_role", "user")
		handler.List(c)
	})

	req, _ := http.NewRequest("GET", "/documents", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	docs, ok := result["documents"].([]interface{})
	if !ok {
		t.Fatal("Expected documents array in response")
	}
	if len(docs) != 2 {
		t.Errorf("Expected 2 documents, got %d", len(docs))
	}
}

func TestListDocumentsWithID(t *testing.T) {
	mockSvc := &mockDocumentService{
		getDocumentFunc: func(ctx context.Context, userCtx docDomain.UserContext, id string) (*docDomain.Document, error) {
			return &docDomain.Document{
				ID:    id,
				Title: "Test Document",
			}, nil
		},
	}
	handler := createTestHandler(mockSvc)

	router := setupTestRouter()
	router.GET("/documents", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Set("user_role", "user")
		handler.List(c)
	})

	req, _ := http.NewRequest("GET", "/documents?id=doc-123", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.Code)
	}
}

func TestListDocumentsError(t *testing.T) {
	mockSvc := &mockDocumentService{
		listDocumentsFunc: func(ctx context.Context, userCtx docDomain.UserContext, limit, offset int) ([]docDomain.Document, int64, error) {
			return nil, 0, errors.New("database error")
		},
	}
	handler := createTestHandler(mockSvc)

	router := setupTestRouter()
	router.GET("/documents", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Set("user_role", "user")
		handler.List(c)
	})

	req, _ := http.NewRequest("GET", "/documents", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", resp.Code)
	}
}

func TestCreateDocument(t *testing.T) {
	mockSvc := &mockDocumentService{
		createDocumentFunc: func(ctx context.Context, userCtx docDomain.UserContext, doc *docDomain.Document) (string, error) {
			return "doc-new-123", nil
		},
	}
	handler := createTestHandler(mockSvc)

	router := setupTestRouter()
	router.POST("/documents", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Set("user_role", "user")
		handler.Create(c)
	})

	body := `{"title": "New Document", "content": "Document content"}`
	req, _ := http.NewRequest("POST", "/documents", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.Code)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body.Bytes(), &result); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if result["id"] != "doc-new-123" {
		t.Errorf("Expected id 'doc-new-123', got '%v'", result["id"])
	}
}

func TestCreateDocumentInvalidBody(t *testing.T) {
	mockSvc := &mockDocumentService{}
	handler := createTestHandler(mockSvc)

	router := setupTestRouter()
	router.POST("/documents", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Set("user_role", "user")
		handler.Create(c)
	})

	body := `{"title": ""}` // missing required content
	req, _ := http.NewRequest("POST", "/documents", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.Code)
	}
}

func TestUpdateDocument(t *testing.T) {
	mockSvc := &mockDocumentService{
		updateDocumentFunc: func(ctx context.Context, userCtx docDomain.UserContext, doc *docDomain.Document) error {
			return nil
		},
	}
	handler := createTestHandler(mockSvc)

	router := setupTestRouter()
	router.PUT("/documents", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Set("user_role", "user")
		handler.Update(c)
	})

	body := `{"id": "doc-123", "title": "Updated Document", "content": "Updated content"}`
	req, _ := http.NewRequest("PUT", "/documents", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.Code)
	}
}

func TestUpdateDocumentNotFound(t *testing.T) {
	mockSvc := &mockDocumentService{
		updateDocumentFunc: func(ctx context.Context, userCtx docDomain.UserContext, doc *docDomain.Document) error {
			return docApp.ErrDocumentNotFound
		},
	}
	handler := createTestHandler(mockSvc)

	router := setupTestRouter()
	router.PUT("/documents", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Set("user_role", "user")
		handler.Update(c)
	})

	body := `{"id": "doc-123", "title": "Updated Document", "content": "Updated content"}`
	req, _ := http.NewRequest("PUT", "/documents", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.Code)
	}
}

func TestUpdateDocumentForbidden(t *testing.T) {
	mockSvc := &mockDocumentService{
		updateDocumentFunc: func(ctx context.Context, userCtx docDomain.UserContext, doc *docDomain.Document) error {
			return docApp.ErrForbidden
		},
	}
	handler := createTestHandler(mockSvc)

	router := setupTestRouter()
	router.PUT("/documents", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Set("user_role", "user")
		handler.Update(c)
	})

	body := `{"id": "doc-123", "title": "Updated Document", "content": "Updated content"}`
	req, _ := http.NewRequest("PUT", "/documents", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", resp.Code)
	}
}

func TestDeleteDocument(t *testing.T) {
	mockSvc := &mockDocumentService{
		deleteDocumentFunc: func(ctx context.Context, userCtx docDomain.UserContext, id string) error {
			return nil
		},
	}
	handler := createTestHandler(mockSvc)

	router := setupTestRouter()
	router.DELETE("/documents", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Set("user_role", "user")
		handler.Delete(c)
	})

	req, _ := http.NewRequest("DELETE", "/documents?id=doc-123", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.Code)
	}
}

func TestDeleteDocumentMissingID(t *testing.T) {
	mockSvc := &mockDocumentService{}
	handler := createTestHandler(mockSvc)

	router := setupTestRouter()
	router.DELETE("/documents", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Set("user_role", "user")
		handler.Delete(c)
	})

	req, _ := http.NewRequest("DELETE", "/documents", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.Code)
	}
}

func TestDeleteDocumentNotFound(t *testing.T) {
	mockSvc := &mockDocumentService{
		deleteDocumentFunc: func(ctx context.Context, userCtx docDomain.UserContext, id string) error {
			return docApp.ErrDocumentNotFound
		},
	}
	handler := createTestHandler(mockSvc)

	router := setupTestRouter()
	router.DELETE("/documents", func(c *gin.Context) {
		c.Set("user_id", "user-123")
		c.Set("user_role", "user")
		handler.Delete(c)
	})

	req, _ := http.NewRequest("DELETE", "/documents?id=doc-123", nil)
	resp := httptest.NewRecorder()

	router.ServeHTTP(resp, req)

	if resp.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", resp.Code)
	}
}

func TestGetUserContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Set("user_id", "user-123")
	ctx.Set("user_role", "admin")

	userCtx := getUserContext(ctx)

	if userCtx.UserID != "user-123" {
		t.Errorf("Expected UserID 'user-123', got '%s'", userCtx.UserID)
	}
	if !userCtx.IsAdmin {
		t.Error("Expected IsAdmin to be true for admin role")
	}
}

func TestGetUserContextNonAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)
	ctx.Set("user_id", "user-123")
	ctx.Set("user_role", "user")

	userCtx := getUserContext(ctx)

	if userCtx.IsAdmin {
		t.Error("Expected IsAdmin to be false for user role")
	}
}
