package rag

import "context"

// ChunkRepository defines operations for chunk persistence and vector search.
type ChunkRepository interface {
	CreateBatch(ctx context.Context, chunks []Chunk) error
	GetByDocumentID(ctx context.Context, documentID string) ([]Chunk, error)
	DeleteByDocumentID(ctx context.Context, documentID string) error
	Search(ctx context.Context, embedding []float64, topK int, threshold float64) ([]Chunk, error)
}
