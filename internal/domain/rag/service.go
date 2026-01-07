package rag

import "context"

// Service defines the RAG (Retrieval-Augmented Generation) operations.
type Service interface {
	// Query performs semantic search and generates an answer using LLM.
	Query(ctx context.Context, query Query) (*Response, error)

	// IndexDocument creates chunks and embeddings for a document's content.
	IndexDocument(ctx context.Context, documentID, content string) error

	// DeleteDocumentChunks removes all chunks associated with a document.
	DeleteDocumentChunks(ctx context.Context, documentID string) error
}
