package auth

import "ypeskov/kkal-tracker/internal/models"

// Servicer defines the auth service contract used by handlers.
type Servicer interface {
	Login(email, password string) (*models.User, string, error)
	Register(email, password, languageCode string, skipActivation bool) (*models.User, string, error)
	GetCurrentUser(userID int) (*models.User, error)
	ActivateUser(token string) error
}
