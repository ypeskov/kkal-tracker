package calorie

import "ypeskov/kkal-tracker/internal/models"

type CreateEntryResult struct {
	Entry                *models.CalorieEntry `json:"entry"`
	NewIngredientCreated bool                 `json:"new_ingredient_created"`
}