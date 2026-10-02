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

// NewIngredient describes a food that is not in the user's ingredient list yet
type NewIngredient struct {
	Name        string
	KcalPer100g float64
	Fats        *float64
	Carbs       *float64
	Proteins    *float64
}

// MealItem is one food of a meal: either an existing ingredient or a new one
type MealItem struct {
	IngredientID  *int
	NewIngredient *NewIngredient
	// OneOff keeps a new food out of the ingredient list: it is written to the diary only
	OneOff bool
	Weight float64
	// KcalPer100g overrides the calorie value of an existing ingredient for this entry only
	KcalPer100g *float64
}

type CreateMealRequest struct {
	UserID       int
	MealDatetime time.Time
	Items        []MealItem
}

// MealEntryResult is a stored diary entry and the ingredient it is based on
type MealEntryResult struct {
	Entry *models.CalorieEntry
	// IngredientID is nil for one-off foods
	IngredientID      *int
	IngredientCreated bool
}

type CreateMealResult struct {
	Entries       []MealEntryResult
	TotalCalories int
	// Day is the diary day (YYYY-MM-DD, UTC) the meal belongs to
	Day string
	// DayTotalCalories is nil when the total could not be calculated after the meal was saved
	DayTotalCalories *int
}
