// Package mongo provides MongoDB repository implementations.
package mongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// DbClient wraps a MongoDB client and database connection.
type DbClient struct {
	client *mongo.Client
	DB     *mongo.Database
}

// NewClient creates a new MongoDB client and connects to the database.
func NewClient(ctx context.Context, uri, dbName string) (*DbClient, error) {
	mc, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := mc.Ping(ctx, nil); err != nil {
		return nil, err
	}
	return &DbClient{client: mc, DB: mc.Database(dbName)}, nil
}

// Ping checks the database connection health.
func (c *DbClient) Ping(ctx context.Context) error {
	return c.client.Ping(ctx, nil)
}

// Close disconnects from the MongoDB server.
func (c *DbClient) Close(ctx context.Context) error {
	return c.client.Disconnect(ctx)
}

// WithTimeout returns a context with a 10-second timeout for database operations.
func (c *DbClient) WithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, 10*time.Second)
}
