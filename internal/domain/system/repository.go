package system

import "context"

// LogRepository defines the data access interface for system logs.
type LogRepository interface {
	Insert(ctx context.Context, entry *LogEntry) error
	List(ctx context.Context, filter LogFilter) ([]LogEntry, int64, error)
	Stats(ctx context.Context) (*LogStats, error)
	DeleteOlderThan(ctx context.Context, days int) (int64, error)
}
