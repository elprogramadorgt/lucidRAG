package document

import (
	"context"

	"github.com/elprogramadorgt/lucidRAG/internal/domain/rag"
)

// Repository defines CRUD operations for documents.
type Repository interface {
	Create(ctx context.Context, doc *Document) (string, error)
	GetByID(ctx context.Context, id string) (*Document, error)
	List(ctx context.Context, limit, offset int) ([]Document, error)
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]Document, error)
	Update(ctx context.Context, doc *Document) error
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int64, error)
	CountByUser(ctx context.Context, userID string) (int64, error)
}

// ChunkRepository is an alias to rag.ChunkRepository for backwards compatibility.
type ChunkRepository = rag.ChunkRepository
