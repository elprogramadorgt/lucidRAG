package logger

import (
	"context"
	"errors"
	"testing"
)

func TestNew(t *testing.T) {
	log := New()
	if log == nil {
		t.Fatal("Expected logger to be created, got nil")
	}
}

func TestNewWithOptions(t *testing.T) {
	log := New(Options{Level: "debug", JSON: true, AddSource: true})
	if log == nil {
		t.Fatal("Expected logger to be created, got nil")
	}
}

func TestLoggerMethods(t *testing.T) {
	log := New(Options{Level: "trace"}) // Enable all levels

	// These should not panic
	log.Trace("Test trace message")
	log.Debug("Test debug message")
	log.Info("Test info message")
	log.Warn("Test warn message")
	log.Error("Test error message")
	log.Critical("Test critical message")
}

func TestLoggerContextMethods(t *testing.T) {
	log := New(Options{Level: "trace"}) // Enable all levels
	ctx := context.Background()

	// These should not panic
	log.TraceContext(ctx, "Test trace message")
	log.DebugContext(ctx, "Test debug message")
	log.InfoContext(ctx, "Test info message")
	log.WarnContext(ctx, "Test warn message")
	log.ErrorContext(ctx, "Test error message")
	log.CriticalContext(ctx, "Test critical message")
}

func TestLoggerStructuredArgs(t *testing.T) {
	log := New()

	// Test structured logging
	log.Info("user action", "user_id", 123, "action", "login")
}

func TestLoggerWith(t *testing.T) {
	log := New()
	childLog := log.With("service", "whatsapp")

	if childLog == nil {
		t.Fatal("Expected child logger to be created")
	}

	childLog.Info("test message")
}

func TestLoggerWithGroup(t *testing.T) {
	log := New()
	groupLog := log.WithGroup("http")

	if groupLog == nil {
		t.Fatal("Expected group logger to be created")
	}

	groupLog.Info("test message", "status", 200)
}

func TestLoggerWithError(t *testing.T) {
	log := New()
	err := errors.New("something went wrong")
	errLog := log.WithError(err)

	if errLog == nil {
		t.Fatal("Expected error logger to be created")
	}

	errLog.Error("operation failed")
}

func TestLoggerWithContext(t *testing.T) {
	log := New()
	ctx := context.WithValue(context.Background(), RequestIDKey, "abc-123")
	ctx = context.WithValue(ctx, UserIDKey, "user-456")
	ctxLog := log.WithContext(ctx)

	if ctxLog == nil {
		t.Fatal("Expected context logger to be created")
	}
}

func TestSetLevel(t *testing.T) {
	log := New(Options{Level: "info"})

	if log.GetLevel() != "info" {
		t.Fatalf("Expected level info, got %s", log.GetLevel())
	}

	log.SetLevel("debug")
	if log.GetLevel() != "debug" {
		t.Fatalf("Expected level debug, got %s", log.GetLevel())
	}

	log.SetLevel("ERROR") // case insensitive
	if log.GetLevel() != "error" {
		t.Fatalf("Expected level error, got %s", log.GetLevel())
	}
}

func TestParseLevelCaseInsensitive(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"trace", "trace"},
		{"TRACE", "trace"},
		{"Trace", "trace"},
		{"debug", "debug"},
		{"DEBUG", "debug"},
		{"Debug", "debug"},
		{"info", "info"},
		{"INFO", "info"},
		{"warn", "warn"},
		{"WARNING", "warn"},
		{"error", "error"},
		{"ERROR", "error"},
		{"critical", "critical"},
		{"CRITICAL", "critical"},
		{"fatal", "critical"}, // alias for critical
		{"invalid", "info"},   // defaults to info
	}

	for _, tt := range tests {
		log := New(Options{Level: tt.input})
		if log.GetLevel() != tt.expected {
			t.Errorf("For input %q: expected %q, got %q", tt.input, tt.expected, log.GetLevel())
		}
	}
}
