package logger

import (
	"testing"
)

func TestNew(t *testing.T) {
	log := New()
	if log == nil {
		t.Fatal("Expected logger to be created, got nil")
	}
}

func TestLoggerMethods(t *testing.T) {
	log := New()

	// These should not panic
	log.Info("Test info message")
	log.Warn("Test warn message")
	log.Error("Test error message")
	log.Debug("Test debug message")
}
