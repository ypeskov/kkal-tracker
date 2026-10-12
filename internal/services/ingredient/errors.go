package ingredient

import (
	"errors"

	"ypeskov/kkal-tracker/internal/models"
)

var (
	ErrEmptyQuery            = errors.New("search query cannot be empty")
	ErrInvalidName           = errors.New("ingredient name is required")
	ErrInvalidKcalPer100g    = errors.New("kcal per 100g must be greater than or equal to 0")
	ErrInvalidNutritionValue = errors.New("nutrition values must be greater than or equal to 0")
	ErrInvalidIngredientID   = errors.New("ingredient ID must be greater than 0")
)

// ValidationError is a problem of the values of an ingredient change. The message is safe to show to the client.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// DuplicateIngredientError means that the user already has another ingredient with the same name
// (ignoring case, spacing, invisible characters and "ё"/"е"). The message is safe to show to the client.
type DuplicateIngredientError struct {
	Existing *models.UserIngredient
}

func (e *DuplicateIngredientError) Error() string {
	return "an ingredient with this name already exists"
}
