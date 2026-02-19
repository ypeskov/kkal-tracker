package apikey

import "ypeskov/kkal-tracker/internal/models"

// Servicer defines the API key service contract used by handlers and middleware.
type Servicer interface {
	CreateKey(userID int, name string, expiryDays *int) (*models.APIKey, string, error)
	ValidateKey(rawKey string) (*models.APIKey, error)
	GetUserKeys(userID int) ([]*models.APIKey, error)
	RevokeKey(id, userID int) error
	DeleteKey(id, userID int) error
}
