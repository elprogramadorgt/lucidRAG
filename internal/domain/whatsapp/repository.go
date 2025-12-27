package whatsapp

import "context"

// Repository defines the data access interface for WhatsApp contacts.
type Repository interface {
	FindByNumber(ctx context.Context, number string) (string, error)
}
