package user

import "context"

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
}

type Service interface {
	Register(ctx context.Context, newUser User) (*User, error)
	RegisterOAuth(ctx context.Context, newUser User, provider, providerID string) (*User, error)
	Login(ctx context.Context, email, password string) (string, *User, error)
	GetUser(ctx context.Context, id string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	ValidateToken(token string) (*Claims, error)
	GenerateToken(user *User) (string, error)
}
