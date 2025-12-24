package mongo

import (
	"context"
	"time"

	"github.com/elprogramadorgt/lucidRAG/internal/domain/user"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepo struct {
	collection *mongo.Collection
}

func NewUserRepo(client *DbClient) *UserRepo {
	return &UserRepo{
		collection: client.DB.Collection("users"),
	}
}

func (r *UserRepo) Create(ctx context.Context, u *user.User) (string, error) {
	u.CreatedAt = time.Now()
	u.UpdatedAt = time.Now()

	if u.ID == "" {
		u.ID = primitive.NewObjectID().Hex()
	}

	_, err := r.collection.InsertOne(ctx, u)
	if err != nil {
		return "", err
	}

	return u.ID, nil
}

func (r *UserRepo) GetByID(ctx context.Context, id string) (*user.User, error) {
	var u user.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&u)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	var u user.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&u)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *UserRepo) Update(ctx context.Context, u *user.User) error {
	u.UpdatedAt = time.Now()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": u.ID},
		bson.M{"$set": u},
	)
	return err
}
