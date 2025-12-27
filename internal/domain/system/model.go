package system

import "time"

// LogEntry represents a single log record stored in the database.
type LogEntry struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	Level     string    `json:"level" bson:"level"`
	Message   string    `json:"message" bson:"message"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp"`
	Source    string    `json:"source,omitempty" bson:"source,omitempty"`
	RequestID string    `json:"request_id,omitempty" bson:"request_id,omitempty"`
	UserID    string    `json:"user_id,omitempty" bson:"user_id,omitempty"`
	Attrs     map[string]any `json:"attrs,omitempty" bson:"attrs,omitempty"`
}

// LogFilter contains criteria for querying log entries.
type LogFilter struct {
	Level     string
	StartTime time.Time
	EndTime   time.Time
	Search    string
	RequestID string
	Source    string
	Limit     int
	Offset    int
}

// LogStats contains aggregated statistics about log entries.
type LogStats struct {
	TotalCount  int64            `json:"total_count"`
	LevelCounts map[string]int64 `json:"level_counts"`
	StartTime   time.Time        `json:"start_time"`
	EndTime     time.Time        `json:"end_time"`
}
