package ingredient

import "ypeskov/kkal-tracker/internal/models"

// Servicer defines the ingredient service contract used by handlers.
type Servicer interface {
	GetAllIngredients(userID int) ([]*models.UserIngredient, error)
	GetIngredientByID(userID, ingredientID int) (*models.UserIngredient, error)
	CreateIngredient(req *CreateIngredientRequest) (*models.UserIngredient, error)
	UpdateIngredient(req *UpdateIngredientRequest) (*models.UserIngredient, error)
	DeleteIngredient(userID, ingredientID int) error
}
