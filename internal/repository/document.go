package repository

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/elprogramadorgt/lucidRAG/internal/domain"
)

// InMemoryDocumentRepository is an in-memory implementation of DocumentRepository
// This is a stub implementation for demonstration purposes
// In production, this should be replaced with a real database implementation
type InMemoryDocumentRepository struct {
	documents map[string]*domain.Document
	chunks    map[string][]*domain.DocumentChunk
	mu        sync.RWMutex
	idCounter uint64
}

// NewInMemoryDocumentRepository creates a new in-memory document repository
func NewInMemoryDocumentRepository() *InMemoryDocumentRepository {
	return &InMemoryDocumentRepository{
		documents: make(map[string]*domain.Document),
		chunks:    make(map[string][]*domain.DocumentChunk),
		idCounter: 0,
	}
}

// Save saves a document
func (r *InMemoryDocumentRepository) Save(ctx context.Context, doc *domain.Document) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if doc.ID == "" {
		id := atomic.AddUint64(&r.idCounter, 1)
		doc.ID = fmt.Sprintf("doc_%d", id)
	}

	r.documents[doc.ID] = doc
	return nil
}

// Update updates a document
func (r *InMemoryDocumentRepository) Update(ctx context.Context, doc *domain.Document) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.documents[doc.ID]; !exists {
		return fmt.Errorf("document not found")
	}

	r.documents[doc.ID] = doc
	return nil
}

// Delete deletes a document
func (r *InMemoryDocumentRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.documents[id]; !exists {
		return fmt.Errorf("document not found")
	}

	delete(r.documents, id)
	delete(r.chunks, id)
	return nil
}

// GetByID retrieves a document by ID
func (r *InMemoryDocumentRepository) GetByID(ctx context.Context, id string) (*domain.Document, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	doc, exists := r.documents[id]
	if !exists {
		return nil, fmt.Errorf("document not found")
	}

	return doc, nil
}

// List retrieves a list of documents
func (r *InMemoryDocumentRepository) List(ctx context.Context, limit, offset int) ([]*domain.Document, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	docs := make([]*domain.Document, 0)
	count := 0
	for _, doc := range r.documents {
		if count >= offset && len(docs) < limit {
			docs = append(docs, doc)
		}
		count++
	}

	return docs, nil
}

// SaveChunk saves a document chunk
func (r *InMemoryDocumentRepository) SaveChunk(ctx context.Context, chunk *domain.DocumentChunk) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.chunks[chunk.DocumentID] = append(r.chunks[chunk.DocumentID], chunk)
	return nil
}

// GetChunksByDocument retrieves chunks for a document
func (r *InMemoryDocumentRepository) GetChunksByDocument(ctx context.Context, docID string) ([]*domain.DocumentChunk, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	chunks, exists := r.chunks[docID]
	if !exists {
		return []*domain.DocumentChunk{}, nil
	}

	return chunks, nil
}

// SearchSimilarChunks searches for similar chunks based on embedding
func (r *InMemoryDocumentRepository) SearchSimilarChunks(ctx context.Context, embedding []float64, topK int, threshold float64) ([]*domain.DocumentChunk, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// TODO: Implement actual similarity search
	// This would require computing cosine similarity between embeddings
	// For now, return empty slice

	return []*domain.DocumentChunk{}, nil
}
