package mongo

import (
	"context"
	"time"

	"github.com/elprogramadorgt/lucidRAG/internal/domain/conversation"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ConversationRepo struct {
	collection *mongo.Collection
}

func NewConversationRepo(client *DbClient) *ConversationRepo {
	return &ConversationRepo{
		collection: client.DB.Collection("conversations"),
	}
}

func (r *ConversationRepo) Create(ctx context.Context, conv *conversation.Conversation) (string, error) {
	conv.CreatedAt = time.Now()
	conv.UpdatedAt = time.Now()
	conv.LastMessageAt = time.Now()

	if conv.ID == "" {
		conv.ID = primitive.NewObjectID().Hex()
	}

	_, err := r.collection.InsertOne(ctx, conv)
	if err != nil {
		return "", err
	}

	return conv.ID, nil
}

func (r *ConversationRepo) GetByID(ctx context.Context, id string) (*conversation.Conversation, error) {
	var conv conversation.Conversation
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&conv)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &conv, nil
}

func (r *ConversationRepo) GetByPhoneNumber(ctx context.Context, phoneNumber string) (*conversation.Conversation, error) {
	var conv conversation.Conversation
	err := r.collection.FindOne(ctx, bson.M{"phone_number": phoneNumber}).Decode(&conv)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &conv, nil
}

func (r *ConversationRepo) List(ctx context.Context, limit, offset int) ([]conversation.Conversation, error) {
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "last_message_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var convs []conversation.Conversation
	if err := cursor.All(ctx, &convs); err != nil {
		return nil, err
	}

	if convs == nil {
		convs = []conversation.Conversation{}
	}

	return convs, nil
}

func (r *ConversationRepo) UpdateLastMessage(ctx context.Context, id string) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{
				"last_message_at": time.Now(),
				"updated_at":      time.Now(),
			},
		},
	)
	return err
}

func (r *ConversationRepo) IncrementMessageCount(ctx context.Context, id string) error {
	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{
			"$inc": bson.M{"message_count": 1},
			"$set": bson.M{"updated_at": time.Now()},
		},
	)
	return err
}

func (r *ConversationRepo) Count(ctx context.Context) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{})
}
