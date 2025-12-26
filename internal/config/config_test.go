package config

import (
	"os"
	"strings"
	"testing"
)

func setRequiredEnvVars() {
	os.Setenv("DB_PASSWORD", "testpassword")
	os.Setenv("WHATSAPP_WEBHOOK_VERIFY_TOKEN", "testtoken")
	os.Setenv("JWT_SECRET", "test-jwt-secret-must-be-32-chars!")
}

func clearRequiredEnvVars() {
	os.Unsetenv("DB_PASSWORD")
	os.Unsetenv("WHATSAPP_WEBHOOK_VERIFY_TOKEN")
	os.Unsetenv("JWT_SECRET")
}

func TestLoad(t *testing.T) {
	setRequiredEnvVars()
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("SERVER_HOST", "127.0.0.1")
	os.Setenv("ENVIRONMENT", "test")

	defer func() {
		clearRequiredEnvVars()
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("SERVER_HOST")
		os.Unsetenv("ENVIRONMENT")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("Expected port 9090, got %d", cfg.Server.Port)
	}

	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("Expected host 127.0.0.1, got %s", cfg.Server.Host)
	}

	if cfg.Server.Environment != "test" {
		t.Errorf("Expected environment test, got %s", cfg.Server.Environment)
	}
}

func TestLoadDefaults(t *testing.T) {
	setRequiredEnvVars()
	defer clearRequiredEnvVars()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("Expected default port 8080, got %d", cfg.Server.Port)
	}

	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Expected default host 0.0.0.0, got %s", cfg.Server.Host)
	}

	if cfg.RAG.ChunkSize != 512 {
		t.Errorf("Expected default chunk size 512, got %d", cfg.RAG.ChunkSize)
	}
}

func TestLoadMissingRequiredEnvVars(t *testing.T) {
	clearRequiredEnvVars()

	_, err := Load()
	if err == nil {
		t.Fatal("Expected error for missing required env vars")
	}

	if !strings.Contains(err.Error(), "DB_PASSWORD") {
		t.Errorf("Expected error to mention DB_PASSWORD, got: %v", err)
	}

	if !strings.Contains(err.Error(), "WHATSAPP_WEBHOOK_VERIFY_TOKEN") {
		t.Errorf("Expected error to mention WHATSAPP_WEBHOOK_VERIFY_TOKEN, got: %v", err)
	}

	if !strings.Contains(err.Error(), "JWT_SECRET") {
		t.Errorf("Expected error to mention JWT_SECRET, got: %v", err)
	}
}

func TestLoadJWTSecretTooShort(t *testing.T) {
	setRequiredEnvVars()
	os.Setenv("JWT_SECRET", "short") // Less than 32 chars

	defer func() {
		clearRequiredEnvVars()
	}()

	_, err := Load()
	if err == nil {
		t.Fatal("Expected error for short JWT_SECRET")
	}

	if !strings.Contains(err.Error(), "JWT_SECRET must be at least 32 characters") {
		t.Errorf("Expected error about JWT_SECRET length, got: %v", err)
	}
}
