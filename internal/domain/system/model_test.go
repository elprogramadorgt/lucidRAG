package system

import (
	"testing"
	"time"
)

func TestLogEntryStruct(t *testing.T) {
	now := time.Now()
	attrs := map[string]any{"key": "value", "count": 10}
	entry := LogEntry{
		ID:        "log-123",
		Level:     "INFO",
		Message:   "Test message",
		Timestamp: now,
		Source:    "test",
		RequestID: "req-456",
		UserID:    "user-789",
		Attrs:     attrs,
	}

	if entry.ID != "log-123" {
		t.Errorf("Expected ID 'log-123', got '%s'", entry.ID)
	}
	if entry.Level != "INFO" {
		t.Errorf("Expected Level 'INFO', got '%s'", entry.Level)
	}
	if entry.Message != "Test message" {
		t.Errorf("Expected Message 'Test message', got '%s'", entry.Message)
	}
	if entry.Source != "test" {
		t.Errorf("Expected Source 'test', got '%s'", entry.Source)
	}
	if entry.RequestID != "req-456" {
		t.Errorf("Expected RequestID 'req-456', got '%s'", entry.RequestID)
	}
	if entry.UserID != "user-789" {
		t.Errorf("Expected UserID 'user-789', got '%s'", entry.UserID)
	}
	if entry.Attrs["key"] != "value" {
		t.Errorf("Expected Attrs[key] 'value', got '%v'", entry.Attrs["key"])
	}
}

func TestLogFilterStruct(t *testing.T) {
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()
	filter := LogFilter{
		Level:     "ERROR",
		StartTime: start,
		EndTime:   end,
		Search:    "error message",
		RequestID: "req-123",
		Source:    "api",
		Limit:     50,
		Offset:    10,
	}

	if filter.Level != "ERROR" {
		t.Errorf("Expected Level 'ERROR', got '%s'", filter.Level)
	}
	if filter.Search != "error message" {
		t.Errorf("Expected Search 'error message', got '%s'", filter.Search)
	}
	if filter.RequestID != "req-123" {
		t.Errorf("Expected RequestID 'req-123', got '%s'", filter.RequestID)
	}
	if filter.Source != "api" {
		t.Errorf("Expected Source 'api', got '%s'", filter.Source)
	}
	if filter.Limit != 50 {
		t.Errorf("Expected Limit 50, got %d", filter.Limit)
	}
	if filter.Offset != 10 {
		t.Errorf("Expected Offset 10, got %d", filter.Offset)
	}
}

func TestLogStatsStruct(t *testing.T) {
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()
	levelCounts := map[string]int64{
		"INFO":  100,
		"ERROR": 10,
		"DEBUG": 50,
	}
	stats := LogStats{
		TotalCount:  160,
		LevelCounts: levelCounts,
		StartTime:   start,
		EndTime:     end,
	}

	if stats.TotalCount != 160 {
		t.Errorf("Expected TotalCount 160, got %d", stats.TotalCount)
	}
	if stats.LevelCounts["INFO"] != 100 {
		t.Errorf("Expected LevelCounts[INFO] 100, got %d", stats.LevelCounts["INFO"])
	}
	if stats.LevelCounts["ERROR"] != 10 {
		t.Errorf("Expected LevelCounts[ERROR] 10, got %d", stats.LevelCounts["ERROR"])
	}
	if stats.LevelCounts["DEBUG"] != 50 {
		t.Errorf("Expected LevelCounts[DEBUG] 50, got %d", stats.LevelCounts["DEBUG"])
	}
}

func TestLogEntryZeroValue(t *testing.T) {
	var entry LogEntry
	if entry.ID != "" {
		t.Errorf("Expected empty ID, got '%s'", entry.ID)
	}
	if entry.Attrs != nil {
		t.Error("Expected nil Attrs by default")
	}
}

func TestLogFilterZeroValue(t *testing.T) {
	var filter LogFilter
	if filter.Level != "" {
		t.Errorf("Expected empty Level, got '%s'", filter.Level)
	}
	if filter.Limit != 0 {
		t.Errorf("Expected Limit 0, got %d", filter.Limit)
	}
	if !filter.StartTime.IsZero() {
		t.Error("Expected zero StartTime")
	}
}

func TestLogStatsZeroValue(t *testing.T) {
	var stats LogStats
	if stats.TotalCount != 0 {
		t.Errorf("Expected TotalCount 0, got %d", stats.TotalCount)
	}
	if stats.LevelCounts != nil {
		t.Error("Expected nil LevelCounts by default")
	}
}
