package user

import "context"

// Repository defines the data access interface for users.
type Repository interface {
	Create(ctx context.Context, user *User) (string, error)
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
}
