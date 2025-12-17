package models

import "time"

// AIProvider represents an AI provider configuration stored in the database
type AIProvider struct {
	ID           string    `json:"id"`
	ProviderType string    `json:"provider_type"`
	DisplayName  string    `json:"display_name"`
	Model        string    `json:"model"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
