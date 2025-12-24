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

type MessageRepo struct {
	collection *mongo.Collection
}

func NewMessageRepo(client *DbClient) *MessageRepo {
	return &MessageRepo{
		collection: client.DB.Collection("messages"),
	}
}

func (r *MessageRepo) Create(ctx context.Context, msg *conversation.Message) (string, error) {
	msg.CreatedAt = time.Now()

	if msg.ID == "" {
		msg.ID = primitive.NewObjectID().Hex()
	}

	_, err := r.collection.InsertOne(ctx, msg)
	if err != nil {
		return "", err
	}

	return msg.ID, nil
}

func (r *MessageRepo) GetByID(ctx context.Context, id string) (*conversation.Message, error) {
	var msg conversation.Message
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&msg)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &msg, nil
}

func (r *MessageRepo) GetByConversationID(ctx context.Context, conversationID string, limit, offset int) ([]conversation.Message, error) {
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "timestamp", Value: -1}})

	cursor, err := r.collection.Find(ctx, bson.M{"conversation_id": conversationID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var msgs []conversation.Message
	if err := cursor.All(ctx, &msgs); err != nil {
		return nil, err
	}

	if msgs == nil {
		msgs = []conversation.Message{}
	}

	return msgs, nil
}

func (r *MessageRepo) CountByConversation(ctx context.Context, conversationID string) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"conversation_id": conversationID})
}
