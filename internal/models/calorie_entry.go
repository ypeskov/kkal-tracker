package models

import (
	"time"
)

// CalorieEntry is a pure data structure representing a calorie entry
type CalorieEntry struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	Food         string    `json:"food"`
	Calories     int       `json:"calories"`
	Weight       float64   `json:"weight"`
	KcalPer100g  float64   `json:"kcalPer100g"`
	Fats         *float64  `json:"fats,omitempty"`
	Carbs        *float64  `json:"carbs,omitempty"`
	Proteins     *float64  `json:"proteins,omitempty"`
	MealDatetime time.Time `json:"meal_datetime"`
	UpdatedAt    time.Time `json:"updated_at"`
	CreatedAt    time.Time `json:"created_at"`
}

// NewMealEntry is one diary entry to be stored as part of a meal
type NewMealEntry struct {
	Food        string
	Calories    int
	Weight      float64
	KcalPer100g float64
	Fats        *float64
	Carbs       *float64
	Proteins    *float64
	// SaveAsIngredient also adds the food to the user's ingredient list
	SaveAsIngredient bool
}

// CreatedMealEntry is a stored diary entry together with the ingredient created for it (if any)
type CreatedMealEntry struct {
	Entry      *CalorieEntry
	Ingredient *UserIngredient
}
