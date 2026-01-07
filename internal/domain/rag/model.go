package rag

import "time"

// Chunk represents a document chunk with its embedding for vector search.
type Chunk struct {
	ID          string    `json:"id" bson:"_id,omitempty"`
	DocumentID  string    `json:"document_id" bson:"document_id"`
	ChunkIndex  int       `json:"chunk_index" bson:"chunk_index"`
	Content     string    `json:"content" bson:"content"`
	Embedding   []float64 `json:"embedding" bson:"embedding"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
}

// Query represents a RAG query request.
type Query struct {
	Query     string  `json:"query"`
	TopK      int     `json:"top_k"`
	Threshold float64 `json:"threshold"`
}

// Response represents the result of a RAG query.
type Response struct {
	Answer           string  `json:"answer"`
	RelevantChunks   []Chunk `json:"relevant_chunks"`
	ConfidenceScore  float64 `json:"confidence_score"`
	ProcessingTimeMs int64   `json:"processing_time_ms"`
}
