package conversation

import "context"

type Service interface {
	GetOrCreateConversation(ctx context.Context, phoneNumber, contactName string) (*Conversation, error)
	ListConversations(ctx context.Context, limit, offset int) ([]Conversation, int64, error)
	GetConversation(ctx context.Context, id string) (*Conversation, error)

	SaveIncomingMessage(ctx context.Context, phoneNumber, contactName, whatsappMsgID, content, msgType string) (*Message, error)
	SaveOutgoingMessage(ctx context.Context, conversationID, content, ragAnswer string) (*Message, error)
	GetMessages(ctx context.Context, conversationID string, limit, offset int) ([]Message, int64, error)
}
