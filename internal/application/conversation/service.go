package conversation

import (
	"context"
	"errors"
	"time"

	conversationDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/conversation"
)

var (
	ErrConversationNotFound = errors.New("conversation not found")
	ErrForbidden            = errors.New("access denied")
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

func (s *service) GetOrCreateConversation(ctx context.Context, userID, phoneNumber, contactName string) (*conversationDomain.Conversation, error) {
	conv, err := s.convRepo.GetByPhoneNumber(ctx, phoneNumber)
	if err != nil {
		return nil, err
	}

	if conv != nil {
		return conv, nil
	}

	newConv := &conversationDomain.Conversation{
		UserID:       userID,
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

func (s *service) ListConversations(ctx context.Context, userCtx conversationDomain.UserContext, limit, offset int) ([]conversationDomain.Conversation, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	var convs []conversationDomain.Conversation
	var total int64
	var err error

	if userCtx.IsAdmin {
		convs, err = s.convRepo.List(ctx, limit, offset)
		if err != nil {
			return nil, 0, err
		}
		total, err = s.convRepo.Count(ctx)
	} else {
		convs, err = s.convRepo.ListByUser(ctx, userCtx.UserID, limit, offset)
		if err != nil {
			return nil, 0, err
		}
		total, err = s.convRepo.CountByUser(ctx, userCtx.UserID)
	}

	if err != nil {
		return nil, 0, err
	}

	return convs, total, nil
}

func (s *service) GetConversation(ctx context.Context, userCtx conversationDomain.UserContext, id string) (*conversationDomain.Conversation, error) {
	conv, err := s.convRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if conv == nil {
		return nil, ErrConversationNotFound
	}

	if !userCtx.IsAdmin && conv.UserID != userCtx.UserID {
		return nil, ErrForbidden
	}

	return conv, nil
}

func (s *service) SaveIncomingMessage(ctx context.Context, phoneNumber, contactName, whatsappMsgID, content, msgType string) (*conversationDomain.Message, error) {
	// For incoming WhatsApp messages, use empty userID (system-created conversations)
	conv, err := s.GetOrCreateConversation(ctx, "", phoneNumber, contactName)
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

	_ = s.convRepo.UpdateLastMessage(ctx, conv.ID)
	_ = s.convRepo.IncrementMessageCount(ctx, conv.ID)

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

	_ = s.convRepo.UpdateLastMessage(ctx, conversationID)
	_ = s.convRepo.IncrementMessageCount(ctx, conversationID)

	return msg, nil
}

func (s *service) GetMessages(ctx context.Context, userCtx conversationDomain.UserContext, conversationID string, limit, offset int) ([]conversationDomain.Message, int64, error) {
	conv, err := s.convRepo.GetByID(ctx, conversationID)
	if err != nil {
		return nil, 0, err
	}
	if conv == nil {
		return nil, 0, ErrConversationNotFound
	}

	if !userCtx.IsAdmin && conv.UserID != userCtx.UserID {
		return nil, 0, ErrForbidden
	}

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
