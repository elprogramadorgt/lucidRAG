package document

import (
	"testing"
	"time"
)

func TestDocumentStruct(t *testing.T) {
	now := time.Now()
	doc := Document{
		ID:         "doc-123",
		UserID:     "user-456",
		Title:      "Test Document",
		Content:    "This is the content",
		Source:     "manual",
		UploadedAt: now,
		UpdatedAt:  now,
		IsActive:   true,
		Metadata:   `{"key": "value"}`,
	}

	if doc.ID != "doc-123" {
		t.Errorf("Expected ID 'doc-123', got '%s'", doc.ID)
	}
	if doc.UserID != "user-456" {
		t.Errorf("Expected UserID 'user-456', got '%s'", doc.UserID)
	}
	if doc.Title != "Test Document" {
		t.Errorf("Expected Title 'Test Document', got '%s'", doc.Title)
	}
	if doc.Content != "This is the content" {
		t.Errorf("Expected Content 'This is the content', got '%s'", doc.Content)
	}
	if doc.Source != "manual" {
		t.Errorf("Expected Source 'manual', got '%s'", doc.Source)
	}
	if !doc.IsActive {
		t.Error("Expected IsActive to be true")
	}
}

func TestChunkStruct(t *testing.T) {
	now := time.Now()
	embedding := []float64{0.1, 0.2, 0.3}
	chunk := Chunk{
		ID:         "chunk-123",
		DocumentID: "doc-456",
		ChunkIndex: 0,
		Content:    "Chunk content",
		Embedding:  embedding,
		CreatedAt:  now,
	}

	if chunk.ID != "chunk-123" {
		t.Errorf("Expected ID 'chunk-123', got '%s'", chunk.ID)
	}
	if chunk.DocumentID != "doc-456" {
		t.Errorf("Expected DocumentID 'doc-456', got '%s'", chunk.DocumentID)
	}
	if chunk.ChunkIndex != 0 {
		t.Errorf("Expected ChunkIndex 0, got %d", chunk.ChunkIndex)
	}
	if len(chunk.Embedding) != 3 {
		t.Errorf("Expected Embedding length 3, got %d", len(chunk.Embedding))
	}
	if chunk.Embedding[0] != 0.1 {
		t.Errorf("Expected first embedding 0.1, got %f", chunk.Embedding[0])
	}
}

func TestRAGQueryStruct(t *testing.T) {
	query := RAGQuery{
		Query:     "What is AI?",
		TopK:      5,
		Threshold: 0.7,
	}

	if query.Query != "What is AI?" {
		t.Errorf("Expected Query 'What is AI?', got '%s'", query.Query)
	}
	if query.TopK != 5 {
		t.Errorf("Expected TopK 5, got %d", query.TopK)
	}
	if query.Threshold != 0.7 {
		t.Errorf("Expected Threshold 0.7, got %f", query.Threshold)
	}
}

func TestRAGResponseStruct(t *testing.T) {
	chunk := Chunk{ID: "chunk-1", Content: "Relevant content"}
	response := RAGResponse{
		Answer:           "AI is artificial intelligence",
		RelevantChunks:   []Chunk{chunk},
		ConfidenceScore:  0.95,
		ProcessingTimeMs: 150,
	}

	if response.Answer != "AI is artificial intelligence" {
		t.Errorf("Expected Answer, got '%s'", response.Answer)
	}
	if len(response.RelevantChunks) != 1 {
		t.Errorf("Expected 1 RelevantChunk, got %d", len(response.RelevantChunks))
	}
	if response.ConfidenceScore != 0.95 {
		t.Errorf("Expected ConfidenceScore 0.95, got %f", response.ConfidenceScore)
	}
	if response.ProcessingTimeMs != 150 {
		t.Errorf("Expected ProcessingTimeMs 150, got %d", response.ProcessingTimeMs)
	}
}

func TestDocumentZeroValue(t *testing.T) {
	var doc Document
	if doc.ID != "" {
		t.Errorf("Expected empty ID, got '%s'", doc.ID)
	}
	if doc.IsActive {
		t.Error("Expected IsActive to be false by default")
	}
}

func TestChunkZeroValue(t *testing.T) {
	var chunk Chunk
	if chunk.ID != "" {
		t.Errorf("Expected empty ID, got '%s'", chunk.ID)
	}
	if chunk.Embedding != nil {
		t.Error("Expected nil Embedding by default")
	}
	if chunk.ChunkIndex != 0 {
		t.Errorf("Expected ChunkIndex 0, got %d", chunk.ChunkIndex)
	}
}
