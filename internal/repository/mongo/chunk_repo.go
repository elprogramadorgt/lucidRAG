package mongo

import (
	"context"
	"time"

	"github.com/elprogramadorgt/lucidRAG/internal/domain/rag"
	"github.com/elprogramadorgt/lucidRAG/pkg/vectormath"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ChunkRepo implements rag.ChunkRepository using MongoDB.
type ChunkRepo struct {
	collection *mongo.Collection
}

// NewChunkRepo creates a new ChunkRepo with the given database client.
func NewChunkRepo(client *DbClient) *ChunkRepo {
	return &ChunkRepo{
		collection: client.DB.Collection("chunks"),
	}
}

func (r *ChunkRepo) CreateBatch(ctx context.Context, chunks []rag.Chunk) error {
	if len(chunks) == 0 {
		return nil
	}

	docs := make([]interface{}, len(chunks))
	for i, chunk := range chunks {
		if chunk.ID == "" {
			chunk.ID = primitive.NewObjectID().Hex()
		}
		chunk.CreatedAt = time.Now()
		docs[i] = chunk
	}

	_, err := r.collection.InsertMany(ctx, docs)
	return err
}

func (r *ChunkRepo) GetByDocumentID(ctx context.Context, documentID string) ([]rag.Chunk, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"document_id": documentID})
	if err != nil {
		return nil, err
	}
	defer func() { _ = cursor.Close(ctx) }()

	var chunks []rag.Chunk
	if err := cursor.All(ctx, &chunks); err != nil {
		return nil, err
	}

	if chunks == nil {
		chunks = []rag.Chunk{}
	}

	return chunks, nil
}

func (r *ChunkRepo) DeleteByDocumentID(ctx context.Context, documentID string) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"document_id": documentID})
	return err
}

func (r *ChunkRepo) Search(ctx context.Context, embedding []float64, topK int, threshold float64) ([]rag.Chunk, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer func() { _ = cursor.Close(ctx) }()

	var allChunks []rag.Chunk
	if err := cursor.All(ctx, &allChunks); err != nil {
		return nil, err
	}

	if len(allChunks) == 0 {
		return []rag.Chunk{}, nil
	}

	vectors := make([][]float64, len(allChunks))
	for i, chunk := range allChunks {
		vectors[i] = chunk.Embedding
	}

	topResults := vectormath.TopKBySimilarity(embedding, vectors, topK, threshold)

	results := make([]rag.Chunk, len(topResults))
	for i, scored := range topResults {
		results[i] = allChunks[scored.Index]
	}

	return results, nil
}
