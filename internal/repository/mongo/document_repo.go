package mongo

import (
	"context"
	"time"

	"github.com/elprogramadorgt/lucidRAG/internal/domain/document"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DocumentRepo implements document.Repository using MongoDB.
type DocumentRepo struct {
	collection *mongo.Collection
}

// NewDocumentRepo creates a new DocumentRepo with the given database client.
func NewDocumentRepo(client *DbClient) *DocumentRepo {
	return &DocumentRepo{
		collection: client.DB.Collection("documents"),
	}
}

func (r *DocumentRepo) Create(ctx context.Context, doc *document.Document) (string, error) {
	doc.UploadedAt = time.Now()
	doc.UpdatedAt = time.Now()
	doc.IsActive = true

	if doc.ID == "" {
		doc.ID = primitive.NewObjectID().Hex()
	}

	_, err := r.collection.InsertOne(ctx, doc)
	if err != nil {
		return "", err
	}

	return doc.ID, nil
}

func (r *DocumentRepo) GetByID(ctx context.Context, id string) (*document.Document, error) {
	var doc document.Document
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &doc, nil
}

func (r *DocumentRepo) List(ctx context.Context, limit, offset int) ([]document.Document, error) {
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "uploaded_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{"is_active": true}, opts)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cursor.Close(ctx) }()

	var docs []document.Document
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	if docs == nil {
		docs = []document.Document{}
	}

	return docs, nil
}

func (r *DocumentRepo) Update(ctx context.Context, doc *document.Document) error {
	doc.UpdatedAt = time.Now()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": doc.ID},
		bson.M{"$set": doc},
	)
	return err
}

func (r *DocumentRepo) Delete(ctx context.Context, id string) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"is_active": false, "updated_at": time.Now()}},
	)
	return err
}

func (r *DocumentRepo) Count(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"is_active": true})
}

func (r *DocumentRepo) ListByUser(ctx context.Context, userID string, limit, offset int) ([]document.Document, error) {
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "uploaded_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{"is_active": true, "user_id": userID}, opts)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cursor.Close(ctx) }()

	var docs []document.Document
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, err
	}

	if docs == nil {
		docs = []document.Document{}
	}

	return docs, nil
}

func (r *DocumentRepo) CountByUser(ctx context.Context, userID string) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"is_active": true, "user_id": userID})
}
