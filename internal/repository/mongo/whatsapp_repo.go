package mongo

import (
	"context"
	"errors"
)

// ErrNotFound is returned when a requested record does not exist.
var ErrNotFound = errors.New("record not found")

// WhatsappRepo implements whatsapp.Repository using MongoDB.
type WhatsappRepo struct {
	c *DbClient
}

// NewWhatsappRepo creates a new WhatsappRepo with the given database client.
func NewWhatsappRepo(c *DbClient) *WhatsappRepo {
	return &WhatsappRepo{c: c}
}

func (r *WhatsappRepo) FindByNumber(ctx context.Context, number string) (string, error) {
	// TODO: Implement actual MongoDB query
	return "", ErrNotFound
}
