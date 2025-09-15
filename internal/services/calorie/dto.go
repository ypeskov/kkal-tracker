package calorie

import (
	"time"

	"ypeskov/kkal-tracker/internal/models"
)

type CreateEntryRequest struct {
	UserID       int
	Food         string
	Calories     int
	Weight       float64
	KcalPer100g  float64
	Fats         *float64
	Carbs        *float64
	Proteins     *float64
	MealDatetime time.Time
}

type UpdateEntryRequest struct {
	EntryID      int
	UserID       int
	Food         string
	Calories     int
	Weight       float64
	KcalPer100g  float64
	Fats         *float64
	Carbs        *float64
	Proteins     *float64
	MealDatetime time.Time
}

type CreateEntryResult struct {
	Entry                *models.CalorieEntry `json:"entry"`
	NewIngredientCreated bool                 `json:"new_ingredient_created"`
}