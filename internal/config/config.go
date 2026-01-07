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
	Auth      AuthConfig
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	JWTSecret      string
	JWTExpiryHours int
	CookieDomain   string
	CookieSecure   bool
	OAuth          OAuthConfig
}

// OAuthConfig holds OAuth provider configurations
type OAuthConfig struct {
	RedirectBaseURL    string
	Google             OAuthProviderConfig
	Facebook           OAuthProviderConfig
	Apple              AppleOAuthConfig
}

// OAuthProviderConfig holds standard OAuth provider settings
type OAuthProviderConfig struct {
	ClientID     string
	ClientSecret string
	Enabled      bool
}

// AppleOAuthConfig holds Apple Sign In specific settings
type AppleOAuthConfig struct {
	ClientID   string
	TeamID     string
	KeyID      string
	PrivateKey string
	Enabled    bool
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port        int
	Host        string
	Environment string
}

// LogLevel returns the appropriate log level for the environment.
func (s ServerConfig) LogLevel() string {
	if s.Environment == "development" {
		return "debug"
	}
	return "info"
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
	OpenAIAPIKey   string
	ModelName      string
	EmbeddingModel string
	ChunkSize      int
	ChunkOverlap   int
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

// MongoURI returns the MongoDB connection URI.
func (d DatabaseConfig) MongoURI() string {
	return fmt.Sprintf("mongodb://%s:%s@%s:%d/%s?authSource=admin",
		d.User, d.Password, d.Host, d.Port, d.Name)
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	port, err := strconv.Atoi(getEnv("SERVER_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid SERVER_PORT: %w", err)
	}

	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "27017"))
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

	jwtExpiry, err := strconv.Atoi(getEnv("JWT_EXPIRY_HOURS", "24"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRY_HOURS: %w", err)
	}

	cookieSecure := getEnv("COOKIE_SECURE", "false") == "true"

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
			OpenAIAPIKey:   getEnv("OPENAI_API_KEY", ""),
			ModelName:      getEnv("RAG_MODEL_NAME", "gpt-3.5-turbo"),
			EmbeddingModel: getEnv("RAG_EMBEDDING_MODEL", "text-embedding-ada-002"),
			ChunkSize:      chunkSize,
			ChunkOverlap:   chunkOverlap,
		},
		Database: DatabaseConfig{
			Type:     getEnv("DB_TYPE", "mongodb"),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			Name:     getEnv("DB_NAME", "lucidrag"),
			User:     getEnv("DB_USER", "lucidrag"),
			Password: getEnv("DB_PASSWORD", ""),
		},
		Auth: AuthConfig{
			JWTSecret:      getEnv("JWT_SECRET", ""),
			JWTExpiryHours: jwtExpiry,
			CookieDomain:   getEnv("COOKIE_DOMAIN", ""),
			CookieSecure:   cookieSecure,
			OAuth: OAuthConfig{
				RedirectBaseURL: getEnv("OAUTH_REDIRECT_BASE_URL", "http://localhost:4200"),
				Google: OAuthProviderConfig{
					ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
					ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
					Enabled:      getEnv("GOOGLE_OAUTH_ENABLED", "false") == "true",
				},
				Facebook: OAuthProviderConfig{
					ClientID:     getEnv("FACEBOOK_CLIENT_ID", ""),
					ClientSecret: getEnv("FACEBOOK_CLIENT_SECRET", ""),
					Enabled:      getEnv("FACEBOOK_OAUTH_ENABLED", "false") == "true",
				},
				Apple: AppleOAuthConfig{
					ClientID:   getEnv("APPLE_CLIENT_ID", ""),
					TeamID:     getEnv("APPLE_TEAM_ID", ""),
					KeyID:      getEnv("APPLE_KEY_ID", ""),
					PrivateKey: getEnv("APPLE_PRIVATE_KEY", ""),
					Enabled:    getEnv("APPLE_OAUTH_ENABLED", "false") == "true",
				},
			},
		},
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) Validate() error {
	var missing []string

	if c.Database.Password == "" {
		missing = append(missing, "DB_PASSWORD")
	}

	if c.WhatsApp.WebhookVerifyToken == "" {
		missing = append(missing, "WHATSAPP_WEBHOOK_VERIFY_TOKEN")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %v", missing)
	}

	return nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
