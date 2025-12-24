package document

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	documentDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/document"
	"github.com/elprogramadorgt/lucidRAG/pkg/chunker"
	"github.com/elprogramadorgt/lucidRAG/pkg/openai"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrDocumentNotFound = errors.New("document not found")
	ErrInvalidQuery     = errors.New("invalid query")
)

type service struct {
	repo           documentDomain.Repository
	chunkRepo      documentDomain.ChunkRepository
	openaiClient   *openai.Client
	chunker        *chunker.Chunker
	embeddingModel string
	modelName      string
}

type ServiceConfig struct {
	Repo           documentDomain.Repository
	ChunkRepo      documentDomain.ChunkRepository
	OpenAIClient   *openai.Client
	Chunker        *chunker.Chunker
	EmbeddingModel string
	ModelName      string
}

func NewService(cfg ServiceConfig) documentDomain.Service {
	embeddingModel := cfg.EmbeddingModel
	if embeddingModel == "" {
		embeddingModel = "text-embedding-ada-002"
	}

	modelName := cfg.ModelName
	if modelName == "" {
		modelName = "gpt-3.5-turbo"
	}

	return &service{
		repo:           cfg.Repo,
		chunkRepo:      cfg.ChunkRepo,
		openaiClient:   cfg.OpenAIClient,
		chunker:        cfg.Chunker,
		embeddingModel: embeddingModel,
		modelName:      modelName,
	}
}

func (s *service) CreateDocument(ctx context.Context, doc *documentDomain.Document) (string, error) {
	id, err := s.repo.Create(ctx, doc)
	if err != nil {
		return "", err
	}

	if s.openaiClient != nil && s.chunker != nil && s.chunkRepo != nil && doc.Content != "" {
		if err := s.createChunksForDocument(ctx, id, doc.Content); err != nil {
			fmt.Printf("warning: failed to create chunks for document %s: %v\n", id, err)
		}
	}

	return id, nil
}

func (s *service) createChunksForDocument(ctx context.Context, documentID, content string) error {
	textChunks := s.chunker.Chunk(content)
	if len(textChunks) == 0 {
		return nil
	}

	chunks := make([]documentDomain.Chunk, 0, len(textChunks))
	for i, text := range textChunks {
		embedding, err := s.openaiClient.CreateEmbedding(ctx, text, s.embeddingModel)
		if err != nil {
			fmt.Printf("warning: failed to create embedding for chunk %d: %v\n", i, err)
			continue
		}

		chunks = append(chunks, documentDomain.Chunk{
			ID:         primitive.NewObjectID().Hex(),
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

	return s.chunkRepo.CreateBatch(ctx, chunks)
}

func (s *service) GetDocument(ctx context.Context, id string) (*documentDomain.Document, error) {
	doc, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if doc == nil {
		return nil, ErrDocumentNotFound
	}
	return doc, nil
}

func (s *service) ListDocuments(ctx context.Context, limit, offset int) ([]documentDomain.Document, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	docs, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	return docs, total, nil
}

func (s *service) UpdateDocument(ctx context.Context, doc *documentDomain.Document) error {
	existing, err := s.repo.GetByID(ctx, doc.ID)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrDocumentNotFound
	}

	doc.UploadedAt = existing.UploadedAt

	if err := s.repo.Update(ctx, doc); err != nil {
		return err
	}

	if s.chunkRepo != nil && doc.Content != existing.Content {
		if err := s.chunkRepo.DeleteByDocumentID(ctx, doc.ID); err != nil {
			fmt.Printf("warning: failed to delete old chunks for document %s: %v\n", doc.ID, err)
		}

		if s.openaiClient != nil && s.chunker != nil && doc.Content != "" {
			if err := s.createChunksForDocument(ctx, doc.ID, doc.Content); err != nil {
				fmt.Printf("warning: failed to create new chunks for document %s: %v\n", doc.ID, err)
			}
		}
	}

	return nil
}

func (s *service) DeleteDocument(ctx context.Context, id string) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrDocumentNotFound
	}

	if s.chunkRepo != nil {
		if err := s.chunkRepo.DeleteByDocumentID(ctx, id); err != nil {
			fmt.Printf("warning: failed to delete chunks for document %s: %v\n", id, err)
		}
	}

	return s.repo.Delete(ctx, id)
}

func (s *service) QueryRAG(ctx context.Context, query documentDomain.RAGQuery) (*documentDomain.RAGResponse, error) {
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
		return &documentDomain.RAGResponse{
			Answer:           "RAG service is not configured. Please set OPENAI_API_KEY.",
			RelevantChunks:   []documentDomain.Chunk{},
			ConfidenceScore:  0.0,
			ProcessingTimeMs: time.Since(start).Milliseconds(),
		}, nil
	}

	queryEmbedding, err := s.openaiClient.CreateEmbedding(ctx, query.Query, s.embeddingModel)
	if err != nil {
		return nil, fmt.Errorf("failed to generate query embedding: %w", err)
	}

	relevantChunks, err := s.chunkRepo.Search(ctx, queryEmbedding, query.TopK, query.Threshold)
	if err != nil {
		return nil, fmt.Errorf("failed to search chunks: %w", err)
	}

	if len(relevantChunks) == 0 {
		return &documentDomain.RAGResponse{
			Answer:           "I couldn't find any relevant information in the knowledge base to answer your question.",
			RelevantChunks:   []documentDomain.Chunk{},
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
		return nil, fmt.Errorf("failed to generate answer: %w", err)
	}

	confidenceScore := 0.85
	if len(relevantChunks) < query.TopK/2 {
		confidenceScore = 0.6
	}

	return &documentDomain.RAGResponse{
		Answer:           answer,
		RelevantChunks:   relevantChunks,
		ConfidenceScore:  confidenceScore,
		ProcessingTimeMs: time.Since(start).Milliseconds(),
	}, nil
}
