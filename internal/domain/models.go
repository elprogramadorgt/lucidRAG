package domain

import "time"

// Message represents a WhatsApp message
type Message struct {
	ID          string    `json:"id"`
	From        string    `json:"from"`
	To          string    `json:"to"`
	Content     string    `json:"content"`
	MessageType string    `json:"message_type"`
	Timestamp   time.Time `json:"timestamp"`
	Status      string    `json:"status"`
}

// ChatSession represents a conversation session
type ChatSession struct {
	ID              string    `json:"id"`
	UserPhoneNumber string    `json:"user_phone_number"`
	StartedAt       time.Time `json:"started_at"`
	LastMessageAt   time.Time `json:"last_message_at"`
	IsActive        bool      `json:"is_active"`
	Context         string    `json:"context"`
}

// Document represents a document in the knowledge base
type Document struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Source      string    `json:"source"`
	UploadedAt  time.Time `json:"uploaded_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	IsActive    bool      `json:"is_active"`
	Metadata    string    `json:"metadata"`
}

// DocumentChunk represents a chunk of a document for RAG
type DocumentChunk struct {
	ID          string    `json:"id"`
	DocumentID  string    `json:"document_id"`
	ChunkIndex  int       `json:"chunk_index"`
	Content     string    `json:"content"`
	Embedding   []float64 `json:"embedding"`
	CreatedAt   time.Time `json:"created_at"`
}

// RAGQuery represents a query to the RAG system
type RAGQuery struct {
	Query     string   `json:"query"`
	TopK      int      `json:"top_k"`
	Threshold float64  `json:"threshold"`
}

// RAGResponse represents a response from the RAG system
type RAGResponse struct {
	Answer           string              `json:"answer"`
	RelevantChunks   []DocumentChunk     `json:"relevant_chunks"`
	ConfidenceScore  float64             `json:"confidence_score"`
	ProcessingTimeMs int64               `json:"processing_time_ms"`
}
