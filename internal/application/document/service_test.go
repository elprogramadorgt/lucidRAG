package document

import (
	"context"
	"errors"
	"testing"
	"time"

	documentDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/document"
)

// mockDocumentRepo is a mock implementation of document.Repository
type mockDocumentRepo struct {
	documents   map[string]*documentDomain.Document
	createError error
	getError    error
}

func newMockDocumentRepo() *mockDocumentRepo {
	return &mockDocumentRepo{
		documents: make(map[string]*documentDomain.Document),
	}
}

func (m *mockDocumentRepo) Create(ctx context.Context, doc *documentDomain.Document) (string, error) {
	if m.createError != nil {
		return "", m.createError
	}
	id := "doc_" + doc.Title
	doc.ID = id
	m.documents[id] = doc
	return id, nil
}

func (m *mockDocumentRepo) GetByID(ctx context.Context, id string) (*documentDomain.Document, error) {
	if m.getError != nil {
		return nil, m.getError
	}
	doc, exists := m.documents[id]
	if !exists {
		return nil, nil
	}
	return doc, nil
}

func (m *mockDocumentRepo) List(ctx context.Context, limit, offset int) ([]documentDomain.Document, error) {
	docs := make([]documentDomain.Document, 0, len(m.documents))
	for _, doc := range m.documents {
		docs = append(docs, *doc)
	}
	return docs, nil
}

func (m *mockDocumentRepo) ListByUser(ctx context.Context, userID string, limit, offset int) ([]documentDomain.Document, error) {
	docs := make([]documentDomain.Document, 0)
	for _, doc := range m.documents {
		if doc.UserID == userID {
			docs = append(docs, *doc)
		}
	}
	return docs, nil
}

func (m *mockDocumentRepo) Count(ctx context.Context) (int64, error) {
	return int64(len(m.documents)), nil
}

func (m *mockDocumentRepo) CountByUser(ctx context.Context, userID string) (int64, error) {
	count := int64(0)
	for _, doc := range m.documents {
		if doc.UserID == userID {
			count++
		}
	}
	return count, nil
}

func (m *mockDocumentRepo) Update(ctx context.Context, doc *documentDomain.Document) error {
	m.documents[doc.ID] = doc
	return nil
}

func (m *mockDocumentRepo) Delete(ctx context.Context, id string) error {
	delete(m.documents, id)
	return nil
}

func TestNewService(t *testing.T) {
	repo := newMockDocumentRepo()
	svc := NewService(ServiceConfig{
		Repo: repo,
	})

	if svc == nil {
		t.Fatal("Expected service to be created")
	}
}

func TestNewServiceDefaults(t *testing.T) {
	repo := newMockDocumentRepo()
	svc := NewService(ServiceConfig{
		Repo: repo,
	})

	if svc == nil {
		t.Fatal("Expected service to be created with defaults")
	}
}

func TestCreateDocument(t *testing.T) {
	repo := newMockDocumentRepo()
	svc := NewService(ServiceConfig{
		Repo: repo,
	})

	ctx := context.Background()
	userCtx := documentDomain.UserContext{
		UserID:  "user-123",
		IsAdmin: false,
	}

	doc := &documentDomain.Document{
		Title:      "test.txt",
		Source:     "text/plain",
		Content:    "Test content",
		UploadedAt: time.Now(),
	}

	id, err := svc.CreateDocument(ctx, userCtx, doc)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if id == "" {
		t.Error("Expected non-empty ID")
	}

	if doc.UserID != "user-123" {
		t.Errorf("Expected UserID user-123, got %s", doc.UserID)
	}
}

func TestGetDocument(t *testing.T) {
	repo := newMockDocumentRepo()
	svc := NewService(ServiceConfig{
		Repo: repo,
	})

	ctx := context.Background()
	userCtx := documentDomain.UserContext{
		UserID:  "user-123",
		IsAdmin: false,
	}

	// Create a document first
	doc := &documentDomain.Document{
		Title:   "test.txt",
		Source:  "text/plain",
		Content: "Test content",
	}
	id, _ := svc.CreateDocument(ctx, userCtx, doc)

	// Get the document
	retrieved, err := svc.GetDocument(ctx, userCtx, id)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if retrieved.Title != "test.txt" {
		t.Errorf("Expected title test.txt, got %s", retrieved.Title)
	}
}

func TestGetDocumentNotFound(t *testing.T) {
	repo := newMockDocumentRepo()
	svc := NewService(ServiceConfig{
		Repo: repo,
	})

	ctx := context.Background()
	userCtx := documentDomain.UserContext{
		UserID:  "user-123",
		IsAdmin: false,
	}

	_, err := svc.GetDocument(ctx, userCtx, "non-existent-id")
	if !errors.Is(err, ErrDocumentNotFound) {
		t.Errorf("Expected ErrDocumentNotFound, got %v", err)
	}
}

func TestGetDocumentForbidden(t *testing.T) {
	repo := newMockDocumentRepo()
	svc := NewService(ServiceConfig{
		Repo: repo,
	})

	ctx := context.Background()

	// Create a document as user-123
	ownerCtx := documentDomain.UserContext{
		UserID:  "user-123",
		IsAdmin: false,
	}
	doc := &documentDomain.Document{
		Title:  "test.txt",
		Source: "text/plain",
	}
	id, _ := svc.CreateDocument(ctx, ownerCtx, doc)

	// Try to access as different user
	otherUserCtx := documentDomain.UserContext{
		UserID:  "user-456",
		IsAdmin: false,
	}

	_, err := svc.GetDocument(ctx, otherUserCtx, id)
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("Expected ErrForbidden, got %v", err)
	}
}

func TestGetDocumentAdminCanAccessAll(t *testing.T) {
	repo := newMockDocumentRepo()
	svc := NewService(ServiceConfig{
		Repo: repo,
	})

	ctx := context.Background()

	// Create a document as user-123
	ownerCtx := documentDomain.UserContext{
		UserID:  "user-123",
		IsAdmin: false,
	}
	doc := &documentDomain.Document{
		Title:  "test.txt",
		Source: "text/plain",
	}
	id, _ := svc.CreateDocument(ctx, ownerCtx, doc)

	// Admin should be able to access
	adminCtx := documentDomain.UserContext{
		UserID:  "admin-user",
		IsAdmin: true,
	}

	retrieved, err := svc.GetDocument(ctx, adminCtx, id)
	if err != nil {
		t.Fatalf("Expected admin to access document, got error: %v", err)
	}
	if retrieved.Title != "test.txt" {
		t.Errorf("Expected title test.txt, got %s", retrieved.Title)
	}
}

func TestListDocuments(t *testing.T) {
	repo := newMockDocumentRepo()
	svc := NewService(ServiceConfig{
		Repo: repo,
	})

	ctx := context.Background()
	userCtx := documentDomain.UserContext{
		UserID:  "user-123",
		IsAdmin: false,
	}

	// Create some documents
	for i := 0; i < 5; i++ {
		doc := &documentDomain.Document{
			Title:  "test" + string(rune('0'+i)) + ".txt",
			Source: "text/plain",
		}
		svc.CreateDocument(ctx, userCtx, doc)
	}

	docs, total, err := svc.ListDocuments(ctx, userCtx, 10, 0)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(docs) != 5 {
		t.Errorf("Expected 5 documents, got %d", len(docs))
	}
	if total != 5 {
		t.Errorf("Expected total 5, got %d", total)
	}
}

func TestListDocumentsWithLimits(t *testing.T) {
	repo := newMockDocumentRepo()
	svc := NewService(ServiceConfig{
		Repo: repo,
	})

	ctx := context.Background()
	userCtx := documentDomain.UserContext{
		UserID:  "user-123",
		IsAdmin: false,
	}

	// Test with negative limit (should default to 10)
	_, _, err := svc.ListDocuments(ctx, userCtx, -1, 0)
	if err != nil {
		t.Fatalf("Expected no error with negative limit, got %v", err)
	}

	// Test with limit > 100 (should cap at 100)
	_, _, err = svc.ListDocuments(ctx, userCtx, 200, 0)
	if err != nil {
		t.Fatalf("Expected no error with large limit, got %v", err)
	}

	// Test with negative offset (should default to 0)
	_, _, err = svc.ListDocuments(ctx, userCtx, 10, -5)
	if err != nil {
		t.Fatalf("Expected no error with negative offset, got %v", err)
	}
}

func TestUpdateDocument(t *testing.T) {
	repo := newMockDocumentRepo()
	svc := NewService(ServiceConfig{
		Repo: repo,
	})

	ctx := context.Background()
	userCtx := documentDomain.UserContext{
		UserID:  "user-123",
		IsAdmin: false,
	}

	// Create a document
	doc := &documentDomain.Document{
		Title:      "test.txt",
		Source:     "text/plain",
		Content:    "Original content",
		UploadedAt: time.Now(),
	}
	id, _ := svc.CreateDocument(ctx, userCtx, doc)

	// Update the document
	updatedDoc := &documentDomain.Document{
		ID:      id,
		Title:   "updated.txt",
		Source:  "text/plain",
		Content: "Updated content",
	}

	err := svc.UpdateDocument(ctx, userCtx, updatedDoc)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify the update
	retrieved, _ := svc.GetDocument(ctx, userCtx, id)
	if retrieved.Title != "updated.txt" {
		t.Errorf("Expected title updated.txt, got %s", retrieved.Title)
	}
}

func TestUpdateDocumentNotFound(t *testing.T) {
	repo := newMockDocumentRepo()
	svc := NewService(ServiceConfig{
		Repo: repo,
	})

	ctx := context.Background()
	userCtx := documentDomain.UserContext{
		UserID:  "user-123",
		IsAdmin: false,
	}

	doc := &documentDomain.Document{
		ID:    "non-existent-id",
		Title: "test.txt",
	}

	err := svc.UpdateDocument(ctx, userCtx, doc)
	if !errors.Is(err, ErrDocumentNotFound) {
		t.Errorf("Expected ErrDocumentNotFound, got %v", err)
	}
}

func TestUpdateDocumentForbidden(t *testing.T) {
	repo := newMockDocumentRepo()
	svc := NewService(ServiceConfig{
		Repo: repo,
	})

	ctx := context.Background()

	// Create as user-123
	ownerCtx := documentDomain.UserContext{
		UserID:  "user-123",
		IsAdmin: false,
	}
	doc := &documentDomain.Document{
		Title: "test.txt",
	}
	id, _ := svc.CreateDocument(ctx, ownerCtx, doc)

	// Try to update as different user
	otherUserCtx := documentDomain.UserContext{
		UserID:  "user-456",
		IsAdmin: false,
	}
	updatedDoc := &documentDomain.Document{
		ID:    id,
		Title: "hacked.txt",
	}

	err := svc.UpdateDocument(ctx, otherUserCtx, updatedDoc)
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("Expected ErrForbidden, got %v", err)
	}
}

func TestDeleteDocument(t *testing.T) {
	repo := newMockDocumentRepo()
	svc := NewService(ServiceConfig{
		Repo: repo,
	})

	ctx := context.Background()
	userCtx := documentDomain.UserContext{
		UserID:  "user-123",
		IsAdmin: false,
	}

	// Create a document
	doc := &documentDomain.Document{
		Title: "test.txt",
	}
	id, _ := svc.CreateDocument(ctx, userCtx, doc)

	// Delete it
	err := svc.DeleteDocument(ctx, userCtx, id)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify it's deleted
	_, err = svc.GetDocument(ctx, userCtx, id)
	if !errors.Is(err, ErrDocumentNotFound) {
		t.Errorf("Expected ErrDocumentNotFound after deletion, got %v", err)
	}
}

func TestDeleteDocumentNotFound(t *testing.T) {
	repo := newMockDocumentRepo()
	svc := NewService(ServiceConfig{
		Repo: repo,
	})

	ctx := context.Background()
	userCtx := documentDomain.UserContext{
		UserID:  "user-123",
		IsAdmin: false,
	}

	err := svc.DeleteDocument(ctx, userCtx, "non-existent-id")
	if !errors.Is(err, ErrDocumentNotFound) {
		t.Errorf("Expected ErrDocumentNotFound, got %v", err)
	}
}

func TestDeleteDocumentForbidden(t *testing.T) {
	repo := newMockDocumentRepo()
	svc := NewService(ServiceConfig{
		Repo: repo,
	})

	ctx := context.Background()

	// Create as user-123
	ownerCtx := documentDomain.UserContext{
		UserID:  "user-123",
		IsAdmin: false,
	}
	doc := &documentDomain.Document{
		Title: "test.txt",
	}
	id, _ := svc.CreateDocument(ctx, ownerCtx, doc)

	// Try to delete as different user
	otherUserCtx := documentDomain.UserContext{
		UserID:  "user-456",
		IsAdmin: false,
	}

	err := svc.DeleteDocument(ctx, otherUserCtx, id)
	if !errors.Is(err, ErrForbidden) {
		t.Errorf("Expected ErrForbidden, got %v", err)
	}
}

