package mongo

import (
	"context"
	"time"

	user "github.com/elprogramadorgt/lucidRAG/internal/domain/user"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepo struct {
	c *DbClient
}

func NewUserRepo(c *DbClient) *UserRepo {
	return &UserRepo{c: c}
}

func (r *UserRepo) FindByEmail(
	ctx context.Context, email string) (*user.User, error) {

	ctx, cancel := r.c.WithTimeout(ctx)
	defer cancel()
	collection := r.c.DB.Collection("users")
	filter := bson.D{{Key: "email", Value: email}}

	var usr user.User
	err := collection.FindOne(ctx, filter).Decode(&usr)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &usr, nil
}

func (r *UserRepo) Create(
	ctx context.Context, usr *user.User) error {

	usr.CreateAt = time.Now()
	usr.UpdateAt = time.Now()
	usr.IsActive = true

	ctx, cancel := r.c.WithTimeout(ctx)
	defer cancel()
	collection := r.c.DB.Collection("users")

	_, err := collection.InsertOne(ctx, usr)
	if err != nil {
		return err
	}
	return nil
}
