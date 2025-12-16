package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	Server    ServerConfig
	WhatsApp  WhatsAppConfig
	RAG       RAGConfig
	Database  DatabaseConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port         int
	Host         string
	Environment  string
}

// WhatsAppConfig holds WhatsApp API configuration
type WhatsAppConfig struct {
	APIKey      string
	PhoneNumberID string
	BusinessAccountID string
	WebhookVerifyToken string
	APIVersion  string
}

// RAGConfig holds RAG-related configuration
type RAGConfig struct {
	ModelName       string
	EmbeddingModel  string
	ChunkSize       int
	ChunkOverlap    int
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type     string
	Host     string
	Port     int
	Name     string
	User     string
	Password string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	port, err := strconv.Atoi(getEnv("SERVER_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid SERVER_PORT: %w", err)
	}

	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	chunkSize, err := strconv.Atoi(getEnv("RAG_CHUNK_SIZE", "512"))
	if err != nil {
		return nil, fmt.Errorf("invalid RAG_CHUNK_SIZE: %w", err)
	}

	chunkOverlap, err := strconv.Atoi(getEnv("RAG_CHUNK_OVERLAP", "50"))
	if err != nil {
		return nil, fmt.Errorf("invalid RAG_CHUNK_OVERLAP: %w", err)
	}

	config := &Config{
		Server: ServerConfig{
			Port:        port,
			Host:        getEnv("SERVER_HOST", "0.0.0.0"),
			Environment: getEnv("ENVIRONMENT", "development"),
		},
		WhatsApp: WhatsAppConfig{
			APIKey:             getEnv("WHATSAPP_API_KEY", ""),
			PhoneNumberID:      getEnv("WHATSAPP_PHONE_NUMBER_ID", ""),
			BusinessAccountID:  getEnv("WHATSAPP_BUSINESS_ACCOUNT_ID", ""),
			WebhookVerifyToken: getEnv("WHATSAPP_WEBHOOK_VERIFY_TOKEN", ""),
			APIVersion:         getEnv("WHATSAPP_API_VERSION", "v17.0"),
		},
		RAG: RAGConfig{
			ModelName:      getEnv("RAG_MODEL_NAME", "gpt-3.5-turbo"),
			EmbeddingModel: getEnv("RAG_EMBEDDING_MODEL", "text-embedding-ada-002"),
			ChunkSize:      chunkSize,
			ChunkOverlap:   chunkOverlap,
		},
		Database: DatabaseConfig{
			Type:     getEnv("DB_TYPE", "postgres"),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			Name:     getEnv("DB_NAME", "lucidrag"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
		},
	}

	return config, nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
