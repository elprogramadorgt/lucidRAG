package rag

import (
	"context"
	"fmt"
	"strings"
	"time"

	ragDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/rag"
	"github.com/elprogramadorgt/lucidRAG/pkg/chunker"
	"github.com/elprogramadorgt/lucidRAG/pkg/logger"
	"github.com/elprogramadorgt/lucidRAG/pkg/openai"
)

// Sentinel errors for RAG operations.
var (
	// ErrInvalidQuery is returned when a RAG query is empty or malformed.
	ErrInvalidQuery = fmt.Errorf("invalid query")
	// ErrNotConfigured is returned when RAG dependencies are not set up.
	ErrNotConfigured = fmt.Errorf("RAG service is not configured")
)

type service struct {
	chunkRepo      ragDomain.ChunkRepository
	openaiClient   *openai.Client
	chunker        *chunker.Chunker
	embeddingModel string
	modelName      string
	log            *logger.Logger
}

// ServiceConfig contains dependencies for creating a RAG service.
type ServiceConfig struct {
	ChunkRepo      ragDomain.ChunkRepository
	OpenAIClient   *openai.Client
	Chunker        *chunker.Chunker
	EmbeddingModel string
	ModelName      string
	Log            *logger.Logger
}

// NewService creates a new RAG service with the given configuration.
func NewService(cfg ServiceConfig) ragDomain.Service {
	embeddingModel := cfg.EmbeddingModel
	if embeddingModel == "" {
		embeddingModel = "text-embedding-ada-002"
	}

	modelName := cfg.ModelName
	if modelName == "" {
		modelName = "gpt-3.5-turbo"
	}

	return &service{
		chunkRepo:      cfg.ChunkRepo,
		openaiClient:   cfg.OpenAIClient,
		chunker:        cfg.Chunker,
		embeddingModel: embeddingModel,
		modelName:      modelName,
		log:            cfg.Log,
	}
}

func (s *service) logWarn(msg string, args ...any) {
	if s.log != nil {
		s.log.Warn(msg, args...)
	}
}

func (s *service) Query(ctx context.Context, query ragDomain.Query) (*ragDomain.Response, error) {
	start := time.Now()

	if query.Query == "" {
		return nil, ErrInvalidQuery
	}

	if query.TopK <= 0 {
		query.TopK = 5
	}
	if query.Threshold <= 0 {
		query.Threshold = 0.7
	}

	if s.openaiClient == nil || s.chunkRepo == nil {
		return &ragDomain.Response{
			Answer:           "RAG service is not configured. Please set OPENAI_API_KEY.",
			RelevantChunks:   []ragDomain.Chunk{},
			ConfidenceScore:  0.0,
			ProcessingTimeMs: time.Since(start).Milliseconds(),
		}, nil
	}

	queryEmbedding, err := s.openaiClient.CreateEmbedding(ctx, query.Query, s.embeddingModel)
	if err != nil {
		return nil, fmt.Errorf("generate query embedding: %w", err)
	}

	relevantChunks, err := s.chunkRepo.Search(ctx, queryEmbedding, query.TopK, query.Threshold)
	if err != nil {
		return nil, fmt.Errorf("search chunks: %w", err)
	}

	if len(relevantChunks) == 0 {
		return &ragDomain.Response{
			Answer:           "I couldn't find any relevant information in the knowledge base to answer your question.",
			RelevantChunks:   []ragDomain.Chunk{},
			ConfidenceScore:  0.0,
			ProcessingTimeMs: time.Since(start).Milliseconds(),
		}, nil
	}

	var contextBuilder strings.Builder
	for i, chunk := range relevantChunks {
		contextBuilder.WriteString(fmt.Sprintf("[Source %d]\n%s\n\n", i+1, chunk.Content))
	}

	systemPrompt := `You are a helpful assistant for a store. Answer questions based ONLY on the provided context.
If the context doesn't contain enough information to answer the question, say so honestly.
Be concise and helpful in your responses.`

	userPrompt := fmt.Sprintf("Context:\n%s\nQuestion: %s", contextBuilder.String(), query.Query)

	messages := []openai.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	answer, err := s.openaiClient.CreateChatCompletion(ctx, messages, s.modelName, nil)
	if err != nil {
		return nil, fmt.Errorf("generate answer: %w", err)
	}

	confidenceScore := 0.85
	if len(relevantChunks) < query.TopK/2 {
		confidenceScore = 0.6
	}

	return &ragDomain.Response{
		Answer:           answer,
		RelevantChunks:   relevantChunks,
		ConfidenceScore:  confidenceScore,
		ProcessingTimeMs: time.Since(start).Milliseconds(),
	}, nil
}

func (s *service) IndexDocument(ctx context.Context, documentID, content string) error {
	if s.openaiClient == nil || s.chunker == nil || s.chunkRepo == nil {
		return nil // RAG not configured, skip indexing
	}

	if content == "" {
		return nil
	}

	textChunks := s.chunker.Chunk(content)
	if len(textChunks) == 0 {
		return nil
	}

	chunks := make([]ragDomain.Chunk, 0, len(textChunks))
	for i, text := range textChunks {
		embedding, err := s.openaiClient.CreateEmbedding(ctx, text, s.embeddingModel)
		if err != nil {
			s.logWarn("failed to create embedding for chunk", "chunk_index", i, "error", err)
			continue
		}

		chunks = append(chunks, ragDomain.Chunk{
			DocumentID: documentID,
			ChunkIndex: i,
			Content:    text,
			Embedding:  embedding,
			CreatedAt:  time.Now(),
		})
	}

	if len(chunks) == 0 {
		return nil
	}

	if err := s.chunkRepo.CreateBatch(ctx, chunks); err != nil {
		return fmt.Errorf("create chunk batch: %w", err)
	}

	return nil
}

func (s *service) DeleteDocumentChunks(ctx context.Context, documentID string) error {
	if s.chunkRepo == nil {
		return nil
	}

	if err := s.chunkRepo.DeleteByDocumentID(ctx, documentID); err != nil {
		return fmt.Errorf("delete document chunks: %w", err)
	}

	return nil
}
