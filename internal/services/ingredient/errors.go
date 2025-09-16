package ingredient

import "errors"

var (
	ErrEmptyQuery            = errors.New("search query cannot be empty")
	ErrInvalidName           = errors.New("ingredient name is required")
	ErrInvalidKcalPer100g    = errors.New("kcal per 100g must be greater than or equal to 0")
	ErrInvalidNutritionValue = errors.New("nutrition values must be greater than or equal to 0")
	ErrInvalidIngredientID   = errors.New("ingredient ID must be greater than 0")
)