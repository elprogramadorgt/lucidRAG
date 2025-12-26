package user

import "time"

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type User struct {
	ID           string    `json:"id" bson:"_id,omitempty"`
	Email        string    `json:"email" bson:"email"`
	PasswordHash string    `json:"-" bson:"password_hash"`
	FirstName    string    `json:"first_name" bson:"first_name"`
	LastName     string    `json:"last_name" bson:"last_name"`
	Role         Role      `json:"role" bson:"role"`
	IsActive     bool      `json:"is_active" bson:"is_active"`
	OAuthProvider   string `json:"oauth_provider,omitempty" bson:"oauth_provider,omitempty"`
	OAuthProviderID string `json:"-" bson:"oauth_provider_id,omitempty"`
	CreatedAt    time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" bson:"updated_at"`
}
