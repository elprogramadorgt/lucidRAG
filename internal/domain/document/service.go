package document

import (
	"context"

	"github.com/elprogramadorgt/lucidRAG/internal/domain/common"
)

// UserContext is an alias for common.UserContext for backwards compatibility.
type UserContext = common.UserContext

// Service defines document CRUD operations.
type Service interface {
	CreateDocument(ctx context.Context, userCtx UserContext, doc *Document) (string, error)
	GetDocument(ctx context.Context, userCtx UserContext, id string) (*Document, error)
	ListDocuments(ctx context.Context, userCtx UserContext, limit, offset int) ([]Document, int64, error)
	UpdateDocument(ctx context.Context, userCtx UserContext, doc *Document) error
	DeleteDocument(ctx context.Context, userCtx UserContext, id string) error
}
