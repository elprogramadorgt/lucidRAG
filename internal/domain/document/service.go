package document

import "context"

type UserContext struct {
	UserID  string
	IsAdmin bool
}

type Service interface {
	CreateDocument(ctx context.Context, userCtx UserContext, doc *Document) (string, error)
	GetDocument(ctx context.Context, userCtx UserContext, id string) (*Document, error)
	ListDocuments(ctx context.Context, userCtx UserContext, limit, offset int) ([]Document, int64, error)
	UpdateDocument(ctx context.Context, userCtx UserContext, doc *Document) error
	DeleteDocument(ctx context.Context, userCtx UserContext, id string) error
	QueryRAG(ctx context.Context, query RAGQuery) (*RAGResponse, error)
}
