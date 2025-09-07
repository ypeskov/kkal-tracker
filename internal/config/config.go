package config

import (
	"os"
	"strconv"
)

type Config struct {
	DatabasePath string
	Port         string
	JWTSecret    string
	LogLevel     string
	Environment  string
}

func New() *Config {
	return &Config{
		DatabasePath: getEnv("DATABASE_PATH", "./data/kkal_tracker.db"),
		Port:         getEnv("PORT", "8080"),
		JWTSecret:    getEnv("JWT_SECRET", "default-secret-key"),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		Environment:  getEnv("ENVIRONMENT", "development"),
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