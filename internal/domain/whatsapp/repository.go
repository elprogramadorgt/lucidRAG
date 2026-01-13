package whatsapp

import "context"

type Repository interface {
	FindByNumber(ctx context.Context, number string) (string, error)
}
