package whatsapp

import "context"

type Repository interface {
	//TODO: Pending to validate the data type
	FindByNumber(ctx context.Context, number string) (string, error)
}
