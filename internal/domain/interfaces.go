package domain

import "context"

// WhatsAppService defines the interface for WhatsApp operations
type WhatsAppService interface {
	SendMessage(ctx context.Context, to, content string) error
	SendTemplateMessage(ctx context.Context, to, templateName string, params map[string]string) error
	VerifyWebhook(verifyToken, mode, challenge string) (string, error)
	ProcessWebhook(ctx context.Context, payload []byte) error
}

// RAGService defines the interface for RAG operations
type RAGService interface {
	Query(ctx context.Context, query RAGQuery) (*RAGResponse, error)
	AddDocument(ctx context.Context, doc *Document) error
	UpdateDocument(ctx context.Context, doc *Document) error
	DeleteDocument(ctx context.Context, docID string) error
	GetDocument(ctx context.Context, docID string) (*Document, error)
	ListDocuments(ctx context.Context, limit, offset int) ([]*Document, error)
}

// MessageRepository defines the interface for message persistence
type MessageRepository interface {
	Save(ctx context.Context, message *Message) error
	GetByID(ctx context.Context, id string) (*Message, error)
	GetBySession(ctx context.Context, sessionID string, limit, offset int) ([]*Message, error)
}

// SessionRepository defines the interface for session persistence
type SessionRepository interface {
	Save(ctx context.Context, session *ChatSession) error
	GetByID(ctx context.Context, id string) (*ChatSession, error)
	GetByPhoneNumber(ctx context.Context, phoneNumber string) (*ChatSession, error)
	UpdateLastMessage(ctx context.Context, sessionID string) error
	SetInactive(ctx context.Context, sessionID string) error
	List(ctx context.Context, limit, offset int) ([]*ChatSession, error)
}

// DocumentRepository defines the interface for document persistence
type DocumentRepository interface {
	Save(ctx context.Context, doc *Document) error
	Update(ctx context.Context, doc *Document) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*Document, error)
	List(ctx context.Context, limit, offset int) ([]*Document, error)
	SaveChunk(ctx context.Context, chunk *DocumentChunk) error
	GetChunksByDocument(ctx context.Context, docID string) ([]*DocumentChunk, error)
	SearchSimilarChunks(ctx context.Context, embedding []float64, topK int, threshold float64) ([]*DocumentChunk, error)
}
