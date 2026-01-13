package logger

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/elprogramadorgt/lucidRAG/internal/domain/system"
)

// mockLogStore implements LogStore for testing
type mockLogStore struct {
	entries  []*system.LogEntry
	insertFn func(ctx context.Context, entry *system.LogEntry) error
}

func (m *mockLogStore) Insert(ctx context.Context, entry *system.LogEntry) error {
	m.entries = append(m.entries, entry)
	if m.insertFn != nil {
		return m.insertFn(ctx, entry)
	}
	return nil
}

// mockHandler implements slog.Handler for testing
type mockHandler struct {
	enabled bool
	records []slog.Record
}

func (m *mockHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return m.enabled
}

func (m *mockHandler) Handle(_ context.Context, r slog.Record) error {
	m.records = append(m.records, r)
	return nil
}

func (m *mockHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return m
}

func (m *mockHandler) WithGroup(name string) slog.Handler {
	return m
}

func TestNewMultiHandler(t *testing.T) {
	store := &mockLogStore{}
	handlers := []slog.Handler{&mockHandler{enabled: true}}

	mh := NewMultiHandler(handlers, store)

	if mh == nil {
		t.Fatal("Expected non-nil MultiHandler")
	}
	if len(mh.handlers) != 1 {
		t.Errorf("Expected 1 handler, got %d", len(mh.handlers))
	}
	if mh.store != store {
		t.Error("Expected store to be set")
	}
}

func TestMultiHandlerEnabled(t *testing.T) {
	tests := []struct {
		name     string
		handlers []slog.Handler
		expected bool
	}{
		{
			name:     "all enabled",
			handlers: []slog.Handler{&mockHandler{enabled: true}, &mockHandler{enabled: true}},
			expected: true,
		},
		{
			name:     "one enabled",
			handlers: []slog.Handler{&mockHandler{enabled: false}, &mockHandler{enabled: true}},
			expected: true,
		},
		{
			name:     "none enabled",
			handlers: []slog.Handler{&mockHandler{enabled: false}, &mockHandler{enabled: false}},
			expected: false,
		},
		{
			name:     "empty handlers",
			handlers: []slog.Handler{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mh := NewMultiHandler(tt.handlers, nil)
			result := mh.Enabled(context.Background(), slog.LevelInfo)
			if result != tt.expected {
				t.Errorf("Expected Enabled() = %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMultiHandlerHandle(t *testing.T) {
	handler1 := &mockHandler{enabled: true}
	handler2 := &mockHandler{enabled: true}
	store := &mockLogStore{}

	mh := NewMultiHandler([]slog.Handler{handler1, handler2}, store)

	record := slog.Record{
		Time:    time.Now(),
		Message: "test message",
		Level:   slog.LevelInfo,
	}

	err := mh.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(handler1.records) != 1 {
		t.Errorf("Expected handler1 to receive 1 record, got %d", len(handler1.records))
	}
	if len(handler2.records) != 1 {
		t.Errorf("Expected handler2 to receive 1 record, got %d", len(handler2.records))
	}
}

func TestMultiHandlerHandleSkipsDisabled(t *testing.T) {
	handler1 := &mockHandler{enabled: true}
	handler2 := &mockHandler{enabled: false}

	mh := NewMultiHandler([]slog.Handler{handler1, handler2}, nil)

	record := slog.Record{
		Time:    time.Now(),
		Message: "test message",
		Level:   slog.LevelInfo,
	}

	err := mh.Handle(context.Background(), record)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(handler1.records) != 1 {
		t.Errorf("Expected handler1 to receive 1 record, got %d", len(handler1.records))
	}
	if len(handler2.records) != 0 {
		t.Errorf("Expected handler2 to receive 0 records, got %d", len(handler2.records))
	}
}

func TestMultiHandlerWithAttrs(t *testing.T) {
	handler := &mockHandler{enabled: true}
	store := &mockLogStore{}
	mh := NewMultiHandler([]slog.Handler{handler}, store)

	attrs := []slog.Attr{slog.String("key", "value")}
	newHandler := mh.WithAttrs(attrs)

	if newHandler == nil {
		t.Fatal("Expected non-nil handler")
	}

	newMH, ok := newHandler.(*MultiHandler)
	if !ok {
		t.Fatal("Expected *MultiHandler type")
	}

	if len(newMH.attrs) != 1 {
		t.Errorf("Expected 1 attr, got %d", len(newMH.attrs))
	}
}

func TestMultiHandlerWithGroup(t *testing.T) {
	handler := &mockHandler{enabled: true}
	store := &mockLogStore{}
	mh := NewMultiHandler([]slog.Handler{handler}, store)

	newHandler := mh.WithGroup("mygroup")

	if newHandler == nil {
		t.Fatal("Expected non-nil handler")
	}

	newMH, ok := newHandler.(*MultiHandler)
	if !ok {
		t.Fatal("Expected *MultiHandler type")
	}

	if len(newMH.groups) != 1 {
		t.Errorf("Expected 1 group, got %d", len(newMH.groups))
	}
	if newMH.groups[0] != "mygroup" {
		t.Errorf("Expected group 'mygroup', got '%s'", newMH.groups[0])
	}
}

func TestLevelToString(t *testing.T) {
	tests := []struct {
		level    slog.Level
		expected string
	}{
		{LevelTrace, "TRACE"},
		{slog.LevelDebug, "DEBUG"},
		{slog.LevelInfo, "INFO"},
		{slog.LevelWarn, "WARN"},
		{slog.LevelError, "ERROR"},
		{LevelCritical, "CRITICAL"},
		{slog.Level(99), "INFO"}, // default case
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := levelToString(tt.level)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestPersistLogWithAttributes(t *testing.T) {
	store := &mockLogStore{}
	handler := &mockHandler{enabled: true}
	mh := NewMultiHandler([]slog.Handler{handler}, store)

	// Add base attributes
	mh = mh.WithAttrs([]slog.Attr{slog.String("source", "test")}).(*MultiHandler)

	record := slog.Record{
		Time:    time.Now(),
		Message: "test message",
		Level:   slog.LevelInfo,
	}
	record.AddAttrs(slog.String("request_id", "req-123"))
	record.AddAttrs(slog.String("user_id", "user-456"))
	record.AddAttrs(slog.String("custom_key", "custom_value"))

	// Call persistLog directly
	mh.persistLog(record)

	// Wait a bit for goroutine to complete
	time.Sleep(100 * time.Millisecond)

	if len(store.entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(store.entries))
	}

	entry := store.entries[0]
	if entry.Message != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", entry.Message)
	}
	if entry.Level != "INFO" {
		t.Errorf("Expected level 'INFO', got '%s'", entry.Level)
	}
	if entry.RequestID != "req-123" {
		t.Errorf("Expected request_id 'req-123', got '%s'", entry.RequestID)
	}
	if entry.UserID != "user-456" {
		t.Errorf("Expected user_id 'user-456', got '%s'", entry.UserID)
	}
}

func TestAddAttr(t *testing.T) {
	mh := &MultiHandler{}

	tests := []struct {
		name     string
		attr     slog.Attr
		checkFn  func(entry *system.LogEntry) bool
	}{
		{
			name: "request_id",
			attr: slog.String("request_id", "req-123"),
			checkFn: func(e *system.LogEntry) bool {
				return e.RequestID == "req-123"
			},
		},
		{
			name: "user_id",
			attr: slog.String("user_id", "user-456"),
			checkFn: func(e *system.LogEntry) bool {
				return e.UserID == "user-456"
			},
		},
		{
			name: "source",
			attr: slog.String("source", "api"),
			checkFn: func(e *system.LogEntry) bool {
				return e.Source == "api"
			},
		},
		{
			name: "handler",
			attr: slog.String("handler", "auth"),
			checkFn: func(e *system.LogEntry) bool {
				return e.Source == "auth"
			},
		},
		{
			name: "custom_attr",
			attr: slog.String("custom", "value"),
			checkFn: func(e *system.LogEntry) bool {
				return e.Attrs["custom"] == "value"
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &system.LogEntry{Attrs: make(map[string]any)}
			mh.addAttr(entry, tt.attr)
			if !tt.checkFn(entry) {
				t.Errorf("Attribute check failed for %s", tt.name)
			}
		})
	}
}
