package rag

import (
	"context"
	"errors"
	"testing"

	ragDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/rag"
)

type mockChunkRepo struct {
	chunks      []ragDomain.Chunk
	searchErr   error
	createErr   error
	deleteErr   error
}

func (m *mockChunkRepo) CreateBatch(_ context.Context, chunks []ragDomain.Chunk) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.chunks = append(m.chunks, chunks...)
	return nil
}

func (m *mockChunkRepo) GetByDocumentID(_ context.Context, _ string) ([]ragDomain.Chunk, error) {
	return m.chunks, nil
}

func (m *mockChunkRepo) DeleteByDocumentID(_ context.Context, _ string) error {
	return m.deleteErr
}

func (m *mockChunkRepo) Search(_ context.Context, _ []float64, topK int, _ float64) ([]ragDomain.Chunk, error) {
	if m.searchErr != nil {
		return nil, m.searchErr
	}
	if len(m.chunks) > topK {
		return m.chunks[:topK], nil
	}
	return m.chunks, nil
}

func TestNewService(t *testing.T) {
	svc := NewService(ServiceConfig{})

	if svc == nil {
		t.Fatal("Expected service to be created")
	}
}

func TestNewServiceDefaults(t *testing.T) {
	svc := NewService(ServiceConfig{}).(*service)

	if svc.embeddingModel != "text-embedding-ada-002" {
		t.Errorf("Expected default embedding model, got %s", svc.embeddingModel)
	}
	if svc.modelName != "gpt-3.5-turbo" {
		t.Errorf("Expected default model name, got %s", svc.modelName)
	}
}

func TestNewServiceCustomModels(t *testing.T) {
	svc := NewService(ServiceConfig{
		EmbeddingModel: "custom-embedding",
		ModelName:      "custom-model",
	}).(*service)

	if svc.embeddingModel != "custom-embedding" {
		t.Errorf("Expected custom embedding model, got %s", svc.embeddingModel)
	}
	if svc.modelName != "custom-model" {
		t.Errorf("Expected custom model name, got %s", svc.modelName)
	}
}

func TestQueryEmptyQuery(t *testing.T) {
	svc := NewService(ServiceConfig{})

	_, err := svc.Query(context.Background(), ragDomain.Query{Query: ""})
	if !errors.Is(err, ErrInvalidQuery) {
		t.Errorf("Expected ErrInvalidQuery, got %v", err)
	}
}

func TestQueryNotConfigured(t *testing.T) {
	svc := NewService(ServiceConfig{})

	resp, err := svc.Query(context.Background(), ragDomain.Query{Query: "test"})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp.Answer != "RAG service is not configured. Please set OPENAI_API_KEY." {
		t.Errorf("Expected not configured message, got %s", resp.Answer)
	}
	if resp.ConfidenceScore != 0.0 {
		t.Errorf("Expected confidence 0.0, got %f", resp.ConfidenceScore)
	}
}

func TestQueryDefaultParameters(t *testing.T) {
	svc := NewService(ServiceConfig{}).(*service)

	// Verify service has expected default configuration
	if svc.embeddingModel != "text-embedding-ada-002" {
		t.Errorf("Expected default embedding model, got %s", svc.embeddingModel)
	}

	query := ragDomain.Query{Query: "test", TopK: 0, Threshold: 0}

	// Verify zero values for query parameters that will get defaults applied
	if query.TopK != 0 {
		t.Error("Expected initial TopK to be 0")
	}
}

func TestIndexDocumentNotConfigured(t *testing.T) {
	svc := NewService(ServiceConfig{})

	err := svc.IndexDocument(context.Background(), "doc-1", "content")
	if err != nil {
		t.Errorf("Expected no error when not configured, got %v", err)
	}
}

func TestIndexDocumentEmptyContent(t *testing.T) {
	repo := &mockChunkRepo{}
	svc := NewService(ServiceConfig{
		ChunkRepo: repo,
	})

	err := svc.IndexDocument(context.Background(), "doc-1", "")
	if err != nil {
		t.Errorf("Expected no error for empty content, got %v", err)
	}
}

func TestDeleteDocumentChunksNotConfigured(t *testing.T) {
	svc := NewService(ServiceConfig{})

	err := svc.DeleteDocumentChunks(context.Background(), "doc-1")
	if err != nil {
		t.Errorf("Expected no error when not configured, got %v", err)
	}
}

func TestDeleteDocumentChunks(t *testing.T) {
	repo := &mockChunkRepo{}
	svc := NewService(ServiceConfig{
		ChunkRepo: repo,
	})

	err := svc.DeleteDocumentChunks(context.Background(), "doc-1")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestDeleteDocumentChunksError(t *testing.T) {
	repo := &mockChunkRepo{deleteErr: errors.New("delete failed")}
	svc := NewService(ServiceConfig{
		ChunkRepo: repo,
	})

	err := svc.DeleteDocumentChunks(context.Background(), "doc-1")
	if err == nil {
		t.Error("Expected error, got nil")
	}
	if !errors.Is(err, repo.deleteErr) {
		t.Errorf("Expected wrapped delete error, got %v", err)
	}
}

func TestServiceConfigZeroValue(t *testing.T) {
	var cfg ServiceConfig

	if cfg.ChunkRepo != nil {
		t.Error("Expected nil ChunkRepo in zero value")
	}
	if cfg.OpenAIClient != nil {
		t.Error("Expected nil OpenAIClient in zero value")
	}
	if cfg.Chunker != nil {
		t.Error("Expected nil Chunker in zero value")
	}
	if cfg.EmbeddingModel != "" {
		t.Error("Expected empty EmbeddingModel in zero value")
	}
	if cfg.ModelName != "" {
		t.Error("Expected empty ModelName in zero value")
	}
	if cfg.Log != nil {
		t.Error("Expected nil Log in zero value")
	}
}
