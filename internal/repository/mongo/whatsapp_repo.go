package mongo

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("record not found")

type WhatsappRepo struct {
	c *DbClient
}

func NewWhatsappRepo(c *DbClient) *WhatsappRepo {
	return &WhatsappRepo{c: c}
}

func (r *WhatsappRepo) FindByNumber(ctx context.Context, number string) (string, error) {
	// TODO: Implement actual MongoDB query
	return "", ErrNotFound
}
