package whatsapp

import (
	"errors"

	whatsappDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/whatsapp"
)

var (
	ErrInvalidToken = errors.New("invalid verify token")
	ErrInvalidMode  = errors.New("invalid mode, expected 'subscribe'")
)

type service struct {
	repo whatsappDomain.Repository
}

func NewService(repo whatsappDomain.Repository) whatsappDomain.Service {
	return &service{repo: repo}
}

func (s *service) VerifyWebhook(req whatsappDomain.HookInput, expectedToken string) (string, error) {
	if req.Mode != "subscribe" {
		return "", ErrInvalidMode
	}

	if req.VerifyToken != expectedToken {
		return "", ErrInvalidToken
	}

	return req.Challenge, nil
}
