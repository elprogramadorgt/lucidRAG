package conversation

import "time"

type MessageDirection string

const (
	DirectionIncoming MessageDirection = "incoming"
	DirectionOutgoing MessageDirection = "outgoing"
)

type Conversation struct {
	ID            string    `json:"id" bson:"_id,omitempty"`
	UserID        string    `json:"user_id" bson:"user_id"`
	PhoneNumber   string    `json:"phone_number" bson:"phone_number"`
	ContactName   string    `json:"contact_name" bson:"contact_name"`
	LastMessageAt time.Time `json:"last_message_at" bson:"last_message_at"`
	MessageCount  int       `json:"message_count" bson:"message_count"`
	CreatedAt     time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" bson:"updated_at"`
}

type Message struct {
	ID             string           `json:"id" bson:"_id,omitempty"`
	ConversationID string           `json:"conversation_id" bson:"conversation_id"`
	WhatsAppMsgID  string           `json:"whatsapp_msg_id" bson:"whatsapp_msg_id"`
	Direction      MessageDirection `json:"direction" bson:"direction"`
	Content        string           `json:"content" bson:"content"`
	MessageType    string           `json:"message_type" bson:"message_type"`
	RAGQueryID     string           `json:"rag_query_id,omitempty" bson:"rag_query_id,omitempty"`
	RAGAnswer      string           `json:"rag_answer,omitempty" bson:"rag_answer,omitempty"`
	Timestamp      time.Time        `json:"timestamp" bson:"timestamp"`
	CreatedAt      time.Time        `json:"created_at" bson:"created_at"`
}
