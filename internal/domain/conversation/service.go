package conversation

import "context"

type UserContext struct {
	UserID  string
	IsAdmin bool
}

type Service interface {
	GetOrCreateConversation(ctx context.Context, userID, phoneNumber, contactName string) (*Conversation, error)
	ListConversations(ctx context.Context, userCtx UserContext, limit, offset int) ([]Conversation, int64, error)
	GetConversation(ctx context.Context, userCtx UserContext, id string) (*Conversation, error)

	SaveIncomingMessage(ctx context.Context, phoneNumber, contactName, whatsappMsgID, content, msgType string) (*Message, error)
	SaveOutgoingMessage(ctx context.Context, conversationID, content, ragAnswer string) (*Message, error)
	GetMessages(ctx context.Context, userCtx UserContext, conversationID string, limit, offset int) ([]Message, int64, error)
}
