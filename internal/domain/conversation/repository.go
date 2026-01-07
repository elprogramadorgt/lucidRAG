package conversation

import "context"

type ConversationRepository interface {
	Create(ctx context.Context, conv *Conversation) (string, error)
	GetByID(ctx context.Context, id string) (*Conversation, error)
	GetByPhoneNumber(ctx context.Context, phoneNumber string) (*Conversation, error)
	List(ctx context.Context, limit, offset int) ([]Conversation, error)
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]Conversation, error)
	UpdateLastMessage(ctx context.Context, id string) error
	IncrementMessageCount(ctx context.Context, id string) error
	Count(ctx context.Context) (int64, error)
	CountByUser(ctx context.Context, userID string) (int64, error)
}

type MessageRepository interface {
	Create(ctx context.Context, msg *Message) (string, error)
	GetByID(ctx context.Context, id string) (*Message, error)
	GetByConversationID(ctx context.Context, conversationID string, limit, offset int) ([]Message, error)
	CountByConversation(ctx context.Context, conversationID string) (int64, error)
}
