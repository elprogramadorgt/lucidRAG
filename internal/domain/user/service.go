package user

import "context"

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

type Service interface {
	Register(ctx context.Context, email, password, name string) (*User, error)
	Login(ctx context.Context, email, password string) (string, error)
	GetUser(ctx context.Context, id string) (*User, error)
	ValidateToken(token string) (*Claims, error)
}
