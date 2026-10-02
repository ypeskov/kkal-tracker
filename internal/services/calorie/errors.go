package calorie

import (
	"errors"
	"fmt"

	"ypeskov/kkal-tracker/internal/models"
)

var (
	ErrInvalidDate   = errors.New("invalid date format")
	ErrEntryNotFound = errors.New("calorie entry not found")
	ErrEmptyMeal     = errors.New("items must contain at least one item")
	ErrTooManyItems  = fmt.Errorf("items must contain at most %d items", MaxMealItems)
)

// MealItemError is a validation problem of a single meal item. The message is safe to show to the client.
type MealItemError struct {
	Index   int
	Message string
}

func (e *MealItemError) Error() string {
	return fmt.Sprintf("items[%d]: %s", e.Index, e.Message)
}

// DuplicateIngredientError means that a new ingredient has the same name (ignoring case and spacing)
// as an ingredient the user already has
type DuplicateIngredientError struct {
	Index    int
	Existing *models.UserIngredient
}

func (e *DuplicateIngredientError) Error() string {
	return fmt.Sprintf("items[%d]: ingredient %q already exists, use its ingredient_id instead of new_ingredient",
		e.Index, e.Existing.Name)
}
