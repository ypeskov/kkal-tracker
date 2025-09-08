package config

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	DatabaseType string
	DatabasePath string
	PostgresURL  string
	Port         string
	JWTSecret    string
	LogLevel     string
	Environment  string
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