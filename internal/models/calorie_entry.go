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