package config

import (
	"strings"
	"testing"
)

func TestResolveJWTSecret(t *testing.T) {
	strongSecret := strings.Repeat("s", minJWTSecretLength)
	shortSecret := strings.Repeat("s", minJWTSecretLength-1)

	tests := []struct {
		name        string
		environment string
		secret      string
		want        string
		wantErr     bool
	}{
		{"development falls back to dev-only secret when empty", environmentDevelopment, "", devOnlyJWTSecret, false},
		{"development keeps placeholder secret", environmentDevelopment, "a-very-secret-key", "a-very-secret-key", false},
		{"development keeps short secret", environmentDevelopment, shortSecret, shortSecret, false},
		{"development keeps strong secret", environmentDevelopment, strongSecret, strongSecret, false},
		{"production accepts strong secret", environmentProduction, strongSecret, strongSecret, false},
		{"production rejects empty secret", environmentProduction, "", "", true},
		{"production rejects placeholder secret", environmentProduction, "your-jwt-secret-key-change-this-in-production", "", true},
		{"production rejects short secret", environmentProduction, shortSecret, "", true},
		{"unknown environment rejects empty secret", "staging", "", "", true},
		{"unknown environment rejects short secret", "staging", shortSecret, "", true},
		{"unknown environment accepts strong secret", "staging", strongSecret, strongSecret, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveJWTSecret(tt.environment, tt.secret)
			if (err != nil) != tt.wantErr {
				t.Fatalf("resolveJWTSecret() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("resolveJWTSecret() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewDefaultsToProductionEnvironment(t *testing.T) {
	t.Setenv("ENVIRONMENT", "")
	t.Setenv("JWT_SECRET", strings.Repeat("s", minJWTSecretLength))
	t.Setenv("DATABASE_PATH", "app.db")

	cfg := New()

	if !cfg.IsProduction() {
		t.Errorf("Environment = %q, want %q when ENVIRONMENT is unset", cfg.Environment, environmentProduction)
	}
}
