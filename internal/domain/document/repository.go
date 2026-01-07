package document

import "context"

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

type ChunkRepository interface {
	CreateBatch(ctx context.Context, chunks []Chunk) error
	GetByDocumentID(ctx context.Context, documentID string) ([]Chunk, error)
	DeleteByDocumentID(ctx context.Context, documentID string) error
	Search(ctx context.Context, embedding []float64, topK int, threshold float64) ([]Chunk, error)
}
