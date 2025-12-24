package whatsapp

import (
	whatsappDomain "github.com/elprogramadorgt/lucidRAG/internal/domain/whatsapp"
)

type service struct {
	repo whatsappDomain.Repository
}

func NewService(repo whatsappDomain.Repository) whatsappDomain.Service {
	return &service{repo: repo}
}

func (s *service) VerifyWebhook(req whatsappDomain.HookInput) (string, error) {

	return req.Challenge, nil

}
