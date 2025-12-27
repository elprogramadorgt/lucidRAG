package document

import (
	"time"

	"github.com/elprogramadorgt/lucidRAG/internal/domain/rag"
)

// Document represents a knowledge base document.
type Document struct {
	ID         string    `json:"id" bson:"_id,omitempty"`
	UserID     string    `json:"user_id" bson:"user_id"`
	Title      string    `json:"title" bson:"title"`
	Content    string    `json:"content" bson:"content"`
	Source     string    `json:"source" bson:"source"`
	UploadedAt time.Time `json:"uploaded_at" bson:"uploaded_at"`
	UpdatedAt  time.Time `json:"updated_at" bson:"updated_at"`
	IsActive   bool      `json:"is_active" bson:"is_active"`
	Metadata   string    `json:"metadata" bson:"metadata"`
}

// Chunk is an alias to rag.Chunk for backwards compatibility.
type Chunk = rag.Chunk

// RAGQuery is an alias to rag.Query for backwards compatibility.
type RAGQuery = rag.Query

// RAGResponse is an alias to rag.Response for backwards compatibility.
type RAGResponse = rag.Response
