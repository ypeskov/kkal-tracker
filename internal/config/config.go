package config

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
)

// AIConfig holds configuration for AI
type AIConfig struct {
	APIKey       string
	BaseURL      string
	Model        string
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
		DatabaseType: getEnv("DATABASE_TYPE", "sqlite"), // sqlite is the default database type
		DatabasePath: databasePath,
		PostgresURL:  getEnv("POSTGRES_URL", ""), 
		Port:         getEnv("PORT", "8080"),
		JWTSecret:    getEnv("JWT_SECRET", "default-secret-key"), // default-secret-key is a placeholder for the actual secret key
		LogLevel:     getEnv("LOG_LEVEL", "info"), // info is the default log level
		Environment:  getEnv("ENVIRONMENT", "development"), // development is the default environment
		// SMTP Configuration
		SMTPHost:     getEnv("SMTP_HOST", "smtp.gmail.com"), // smtp.gmail.com is the default SMTP host
		SMTPPort:     getEnvInt("SMTP_PORT", 587),
		SMTPUser:     getEnv("SMTP_USER", ""), // SMTP_USER is the default SMTP user
		SMTPPassword: getEnv("SMTP_PASSWORD", ""), // SMTP_PASSWORD is the default SMTP password
		SMTPFrom:     getEnv("SMTP_FROM", "noreply@kkal-tracker.com"),
		AppURL:       getEnv("APP_URL", "http://localhost:8080"), // http://localhost:8080 is the default app URL
		// AI Configuration
		AI: AIConfig{
			APIKey:       getEnv("OPENAI_API_KEY", ""), // OPENAI_API_KEY is the default OpenAI API key
			BaseURL:      getEnv("OPENAI_BASE_URL", ""), // OPENAI_BASE_URL is the default OpenAI base URL
			Model:        getEnv("OPENAI_MODEL", "gpt-5.2"), // gpt-5.2 is the default OpenAI model
			UseMaxTokens: getEnvBool("AI_USE_MAX_TOKENS", false), // AI_USE_MAX_TOKENS is the default AI use max tokens
			MaxTokens:    getEnvInt("AI_MAX_TOKENS", 2000), // AI_MAX_TOKENS is the default AI max tokens
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
