package user

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID    primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email string             `bson:"email" json:"email"`
	//TODO: what does the - meamn in json tag?
	PasswordHash string    `bson:"password_hash" json:"-"`
	FirstName    string    `bson:"first_name" json:"first_name"`
	LastName     string    `bson:"last_name" json:"last_name"`
	Role         string    `bson:"role" json:"role"`
	IsActive     bool      `bson:"is_active" json:"is_active"`
	CreateAt     time.Time `bson:"create_at" json:"create_at"`
	UpdateAt     time.Time `bson:"update_at" json:"update_at"`
}
