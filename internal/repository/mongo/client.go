package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DbClient struct{ DB *mongo.Database }

func NewClient(ctx context.Context, uri, dbName string) (*DbClient, error) {
	mc, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := mc.Ping(ctx, nil); err != nil {
		return nil, err
	}
	return &DbClient{DB: mc.Database(dbName)}, nil
}

func (c *DbClient) WithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, 10*time.Second)
}
