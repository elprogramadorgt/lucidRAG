package config

import (
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	t.Setenv("DB_PASSWORD", "testpassword")
	t.Setenv("WHATSAPP_WEBHOOK_VERIFY_TOKEN", "testtoken")
	t.Setenv("SERVER_PORT", "9090")
	t.Setenv("SERVER_HOST", "127.0.0.1")
	t.Setenv("ENVIRONMENT", "test")

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
	t.Setenv("DB_PASSWORD", "testpassword")
	t.Setenv("WHATSAPP_WEBHOOK_VERIFY_TOKEN", "testtoken")

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
	// t.Setenv clears env vars after test, so we don't set them here
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
}

func TestLoadOAuthConfig(t *testing.T) {
	t.Setenv("DB_PASSWORD", "testpassword")
	t.Setenv("WHATSAPP_WEBHOOK_VERIFY_TOKEN", "testtoken")
	t.Setenv("OAUTH_REDIRECT_BASE_URL", "http://localhost:3000")
	t.Setenv("GOOGLE_OAUTH_ENABLED", "true")
	t.Setenv("GOOGLE_CLIENT_ID", "google-client-id")
	t.Setenv("GOOGLE_CLIENT_SECRET", "google-client-secret")
	t.Setenv("FACEBOOK_OAUTH_ENABLED", "true")
	t.Setenv("FACEBOOK_CLIENT_ID", "facebook-client-id")
	t.Setenv("FACEBOOK_CLIENT_SECRET", "facebook-client-secret")
	t.Setenv("APPLE_OAUTH_ENABLED", "true")
	t.Setenv("APPLE_CLIENT_ID", "apple-client-id")
	t.Setenv("APPLE_TEAM_ID", "apple-team-id")
	t.Setenv("APPLE_KEY_ID", "apple-key-id")
	t.Setenv("APPLE_PRIVATE_KEY", "apple-private-key")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test OAuth redirect base URL
	if cfg.Auth.OAuth.RedirectBaseURL != "http://localhost:3000" {
		t.Errorf("Expected redirect base URL http://localhost:3000, got %s", cfg.Auth.OAuth.RedirectBaseURL)
	}

	// Test Google OAuth config
	if !cfg.Auth.OAuth.Google.Enabled {
		t.Error("Expected Google OAuth to be enabled")
	}
	if cfg.Auth.OAuth.Google.ClientID != "google-client-id" {
		t.Errorf("Expected Google client ID google-client-id, got %s", cfg.Auth.OAuth.Google.ClientID)
	}
	if cfg.Auth.OAuth.Google.ClientSecret != "google-client-secret" {
		t.Errorf("Expected Google client secret google-client-secret, got %s", cfg.Auth.OAuth.Google.ClientSecret)
	}

	// Test Facebook OAuth config
	if !cfg.Auth.OAuth.Facebook.Enabled {
		t.Error("Expected Facebook OAuth to be enabled")
	}
	if cfg.Auth.OAuth.Facebook.ClientID != "facebook-client-id" {
		t.Errorf("Expected Facebook client ID facebook-client-id, got %s", cfg.Auth.OAuth.Facebook.ClientID)
	}

	// Test Apple OAuth config
	if !cfg.Auth.OAuth.Apple.Enabled {
		t.Error("Expected Apple OAuth to be enabled")
	}
	if cfg.Auth.OAuth.Apple.ClientID != "apple-client-id" {
		t.Errorf("Expected Apple client ID apple-client-id, got %s", cfg.Auth.OAuth.Apple.ClientID)
	}
	if cfg.Auth.OAuth.Apple.TeamID != "apple-team-id" {
		t.Errorf("Expected Apple team ID apple-team-id, got %s", cfg.Auth.OAuth.Apple.TeamID)
	}
	if cfg.Auth.OAuth.Apple.KeyID != "apple-key-id" {
		t.Errorf("Expected Apple key ID apple-key-id, got %s", cfg.Auth.OAuth.Apple.KeyID)
	}
}

func TestLoadOAuthDefaults(t *testing.T) {
	t.Setenv("DB_PASSWORD", "testpassword")
	t.Setenv("WHATSAPP_WEBHOOK_VERIFY_TOKEN", "testtoken")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Test OAuth defaults
	if cfg.Auth.OAuth.RedirectBaseURL != "http://localhost:4200" {
		t.Errorf("Expected default redirect base URL http://localhost:4200, got %s", cfg.Auth.OAuth.RedirectBaseURL)
	}
	if cfg.Auth.OAuth.Google.Enabled {
		t.Error("Expected Google OAuth to be disabled by default")
	}
	if cfg.Auth.OAuth.Facebook.Enabled {
		t.Error("Expected Facebook OAuth to be disabled by default")
	}
	if cfg.Auth.OAuth.Apple.Enabled {
		t.Error("Expected Apple OAuth to be disabled by default")
	}
}

func TestLoadCookieConfig(t *testing.T) {
	t.Setenv("DB_PASSWORD", "testpassword")
	t.Setenv("WHATSAPP_WEBHOOK_VERIFY_TOKEN", "testtoken")
	t.Setenv("COOKIE_DOMAIN", "example.com")
	t.Setenv("COOKIE_SECURE", "true")
	t.Setenv("JWT_EXPIRY_HOURS", "48")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Auth.CookieDomain != "example.com" {
		t.Errorf("Expected cookie domain example.com, got %s", cfg.Auth.CookieDomain)
	}
	if !cfg.Auth.CookieSecure {
		t.Error("Expected cookie secure to be true")
	}
	if cfg.Auth.JWTExpiryHours != 48 {
		t.Errorf("Expected JWT expiry hours 48, got %d", cfg.Auth.JWTExpiryHours)
	}
}

func TestLoadInvalidPort(t *testing.T) {
	t.Setenv("DB_PASSWORD", "testpassword")
	t.Setenv("WHATSAPP_WEBHOOK_VERIFY_TOKEN", "testtoken")
	t.Setenv("SERVER_PORT", "invalid")

	_, err := Load()
	if err == nil {
		t.Fatal("Expected error for invalid port")
	}
	if !strings.Contains(err.Error(), "SERVER_PORT") {
		t.Errorf("Expected error to mention SERVER_PORT, got: %v", err)
	}
}

func TestLoadInvalidJWTExpiry(t *testing.T) {
	t.Setenv("DB_PASSWORD", "testpassword")
	t.Setenv("WHATSAPP_WEBHOOK_VERIFY_TOKEN", "testtoken")
	t.Setenv("JWT_EXPIRY_HOURS", "invalid")

	_, err := Load()
	if err == nil {
		t.Fatal("Expected error for invalid JWT expiry hours")
	}
	if !strings.Contains(err.Error(), "JWT_EXPIRY_HOURS") {
		t.Errorf("Expected error to mention JWT_EXPIRY_HOURS, got: %v", err)
	}
}
