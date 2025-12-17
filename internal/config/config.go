package config

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
)

// AIProviderConfig holds configuration for a single AI provider
type AIProviderConfig struct {
	APIKey  string
	BaseURL string
}

// AIConfig holds configuration for all AI providers
type AIConfig struct {
	OpenAI       AIProviderConfig
	Anthropic    AIProviderConfig
	Ollama       AIProviderConfig
	UseMaxTokens bool // Whether to limit response tokens
	MaxTokens    int  // Maximum completion tokens (if UseMaxTokens is true)
}

type Config struct {
	DatabaseType string
	DatabasePath string
	PostgresURL  string
	Port         string
	JWTSecret    string
	LogLevel     string
	Environment  string
	// SMTP Configuration
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string
	AppURL       string
	// AI Configuration
	AI AIConfig
}

func New() *Config {
	databasePath := getEnv("DATABASE_PATH", "./data/kkal_tracker.db")

	// Ensure the database directory exists
	if dir := filepath.Dir(databasePath); dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create database directory %s: %v", dir, err)
		}
	}

	return &Config{
		DatabaseType: getEnv("DATABASE_TYPE", "sqlite"),
		DatabasePath: databasePath,
		PostgresURL:  getEnv("POSTGRES_URL", ""),
		Port:         getEnv("PORT", "8080"),
		JWTSecret:    getEnv("JWT_SECRET", "default-secret-key"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		Environment:  getEnv("ENVIRONMENT", "development"),
		// SMTP Configuration
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"),
		SMTPPort:     getEnvInt("SMTP_PORT", 587),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnv("SMTP_PASSWORD", ""),
		SMTPFrom:     getEnv("SMTP_FROM", "noreply@kkal-tracker.com"),
		AppURL:       getEnv("APP_URL", "http://localhost:8080"),
		// AI Configuration
		AI: AIConfig{
			OpenAI: AIProviderConfig{
				APIKey:  getEnv("OPENAI_API_KEY", ""),
				BaseURL: getEnv("OPENAI_BASE_URL", ""),
			},
			Anthropic: AIProviderConfig{
				APIKey:  getEnv("ANTHROPIC_API_KEY", ""),
				BaseURL: getEnv("ANTHROPIC_BASE_URL", ""),
			},
			Ollama: AIProviderConfig{
				BaseURL: getEnv("OLLAMA_BASE_URL", "http://localhost:11434"),
			},
			UseMaxTokens: getEnvBool("AI_USE_MAX_TOKENS", false),
			MaxTokens:    getEnvInt("AI_MAX_TOKENS", 2000),
		},
	}
}

func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
