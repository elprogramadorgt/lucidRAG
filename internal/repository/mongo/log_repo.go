package mongo

import (
	"context"
	"time"

	"github.com/elprogramadorgt/lucidRAG/internal/domain/system"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LogRepo struct {
	col *mongo.Collection
}

func NewLogRepo(client *DbClient) *LogRepo {
	col := client.DB.Collection("logs")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_, _ = col.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{Keys: bson.D{{Key: "timestamp", Value: -1}}},
		{Keys: bson.D{{Key: "level", Value: 1}}},
		{Keys: bson.D{{Key: "request_id", Value: 1}}},
	})
	return &LogRepo{col: col}
}

func (r *LogRepo) Insert(ctx context.Context, entry *system.LogEntry) error {
	if entry.ID == "" {
		entry.ID = primitive.NewObjectID().Hex()
	}
	_, err := r.col.InsertOne(ctx, entry)
	return err
}

func (r *LogRepo) List(ctx context.Context, filter system.LogFilter) ([]system.LogEntry, int64, error) {
	query := bson.M{}

	if filter.Level != "" {
		query["level"] = filter.Level
	}
	if !filter.StartTime.IsZero() {
		if query["timestamp"] == nil {
			query["timestamp"] = bson.M{}
		}
		query["timestamp"].(bson.M)["$gte"] = filter.StartTime
	}
	if !filter.EndTime.IsZero() {
		if query["timestamp"] == nil {
			query["timestamp"] = bson.M{}
		}
		query["timestamp"].(bson.M)["$lte"] = filter.EndTime
	}
	if filter.Search != "" {
		query["message"] = bson.M{"$regex": filter.Search, "$options": "i"}
	}
	if filter.RequestID != "" {
		query["request_id"] = filter.RequestID
	}
	if filter.Source != "" {
		query["source"] = filter.Source
	}

	total, err := r.col.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(filter.Offset))

	cursor, err := r.col.Find(ctx, query, opts)
	if err != nil {
		return nil, 0, err
	}
	defer func() { _ = cursor.Close(ctx) }()

	var entries []system.LogEntry
	if err := cursor.All(ctx, &entries); err != nil {
		return nil, 0, err
	}
	return entries, total, nil
}

func (r *LogRepo) Stats(ctx context.Context) (*system.LogStats, error) {
	total, err := r.col.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	pipeline := []bson.M{
		{"$group": bson.M{"_id": "$level", "count": bson.M{"$sum": 1}}},
	}
	cursor, err := r.col.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer func() { _ = cursor.Close(ctx) }()

	levelCounts := make(map[string]int64)
	for cursor.Next(ctx) {
		var result struct {
			ID    string `bson:"_id"`
			Count int64  `bson:"count"`
		}
		if err := cursor.Decode(&result); err == nil {
			levelCounts[result.ID] = result.Count
		}
	}

	var oldest, newest system.LogEntry
	_ = r.col.FindOne(ctx, bson.M{}, options.FindOne().SetSort(bson.D{{Key: "timestamp", Value: 1}})).Decode(&oldest)
	_ = r.col.FindOne(ctx, bson.M{}, options.FindOne().SetSort(bson.D{{Key: "timestamp", Value: -1}})).Decode(&newest)

	return &system.LogStats{
		TotalCount:  total,
		LevelCounts: levelCounts,
		StartTime:   oldest.Timestamp,
		EndTime:     newest.Timestamp,
	}, nil
}

func (r *LogRepo) DeleteOlderThan(ctx context.Context, days int) (int64, error) {
	cutoff := time.Now().AddDate(0, 0, -days)
	result, err := r.col.DeleteMany(ctx, bson.M{"timestamp": bson.M{"$lt": cutoff}})
	if err != nil {
		return 0, err
	}
	return result.DeletedCount, nil
}
