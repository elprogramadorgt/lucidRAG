package conversation

import (
	"context"
	"errors"
	"time"

	conversationDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/conversation"
)

var (
	ErrConversationNotFound = errors.New("conversation not found")
)

type service struct {
	convRepo conversationDomain.ConversationRepository
	msgRepo  conversationDomain.MessageRepository
}

type ServiceConfig struct {
	ConvRepo conversationDomain.ConversationRepository
	MsgRepo  conversationDomain.MessageRepository
}

func NewService(cfg ServiceConfig) conversationDomain.Service {
	return &service{
		convRepo: cfg.ConvRepo,
		msgRepo:  cfg.MsgRepo,
	}
}

func (s *service) GetOrCreateConversation(ctx context.Context, phoneNumber, contactName string) (*conversationDomain.Conversation, error) {
	conv, err := s.convRepo.GetByPhoneNumber(ctx, phoneNumber)
	if err != nil {
		return nil, err
	}

	if conv != nil {
		return conv, nil
	}

	newConv := &conversationDomain.Conversation{
		PhoneNumber:  phoneNumber,
		ContactName:  contactName,
		MessageCount: 0,
	}

	id, err := s.convRepo.Create(ctx, newConv)
	if err != nil {
		return nil, err
	}
	newConv.ID = id

	return newConv, nil
}

func (s *service) ListConversations(ctx context.Context, limit, offset int) ([]conversationDomain.Conversation, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	convs, err := s.convRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.convRepo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	return convs, total, nil
}

func (s *service) GetConversation(ctx context.Context, id string) (*conversationDomain.Conversation, error) {
	conv, err := s.convRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if conv == nil {
		return nil, ErrConversationNotFound
	}
	return conv, nil
}

func (s *service) SaveIncomingMessage(ctx context.Context, phoneNumber, contactName, whatsappMsgID, content, msgType string) (*conversationDomain.Message, error) {
	conv, err := s.GetOrCreateConversation(ctx, phoneNumber, contactName)
	if err != nil {
		return nil, err
	}

	msg := &conversationDomain.Message{
		ConversationID: conv.ID,
		WhatsAppMsgID:  whatsappMsgID,
		Direction:      conversationDomain.DirectionIncoming,
		Content:        content,
		MessageType:    msgType,
		Timestamp:      time.Now(),
	}

	id, err := s.msgRepo.Create(ctx, msg)
	if err != nil {
		return nil, err
	}
	msg.ID = id

	s.convRepo.UpdateLastMessage(ctx, conv.ID)
	s.convRepo.IncrementMessageCount(ctx, conv.ID)

	return msg, nil
}

func (s *service) SaveOutgoingMessage(ctx context.Context, conversationID, content, ragAnswer string) (*conversationDomain.Message, error) {
	msg := &conversationDomain.Message{
		ConversationID: conversationID,
		Direction:      conversationDomain.DirectionOutgoing,
		Content:        content,
		MessageType:    "text",
		RAGAnswer:      ragAnswer,
		Timestamp:      time.Now(),
	}

	id, err := s.msgRepo.Create(ctx, msg)
	if err != nil {
		return nil, err
	}
	msg.ID = id

	s.convRepo.UpdateLastMessage(ctx, conversationID)
	s.convRepo.IncrementMessageCount(ctx, conversationID)

	return msg, nil
}

func (s *service) GetMessages(ctx context.Context, conversationID string, limit, offset int) ([]conversationDomain.Message, int64, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}

	msgs, err := s.msgRepo.GetByConversationID(ctx, conversationID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.msgRepo.CountByConversation(ctx, conversationID)
	if err != nil {
		return nil, 0, err
	}

	return msgs, total, nil
}
