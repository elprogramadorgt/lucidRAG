package conversation

import (
	"context"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ConversationRepository interface {
	Create(ctx context.Context, conv *Conversation) (primitive.ObjectID, error)
	GetByID(ctx context.Context, id string) (*Conversation, error)
	GetByPhoneNumber(ctx context.Context, phoneNumber string) (*Conversation, error)
	List(ctx context.Context, limit, offset int) ([]Conversation, error)
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]Conversation, error)
	UpdateLastMessage(ctx context.Context, id primitive.ObjectID) error
	IncrementMessageCount(ctx context.Context, id primitive.ObjectID) error
	Count(ctx context.Context) (int64, error)
	CountByUser(ctx context.Context, userID string) (int64, error)
}

type MessageRepository interface {
	Create(ctx context.Context, msg *Message) (primitive.ObjectID, error)
	GetByID(ctx context.Context, id string) (*Message, error)
	GetByConversationID(ctx context.Context, conversationID primitive.ObjectID, limit, offset int) ([]Message, error)
	CountByConversation(ctx context.Context, conversationID primitive.ObjectID) (int64, error)
}
