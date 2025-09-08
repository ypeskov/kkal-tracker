package models

import (
	"time"
)

// GlobalIngredient represents admin-managed ingredients
type GlobalIngredient struct {
	ID          int               `json:"id"`
	KcalPer100g float64           `json:"kcalPer100g"`
	Fats        *float64          `json:"fats,omitempty"`
	Carbs       *float64          `json:"carbs,omitempty"`
	Proteins    *float64          `json:"proteins,omitempty"`
	Names       map[string]string `json:"names"` // language_code -> name
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// UserIngredient represents user-specific ingredients (copied from global on registration)
type UserIngredient struct {
	ID                 int       `json:"id"`
	UserID             int       `json:"user_id"`
	Name               string    `json:"name"`
	KcalPer100g        float64   `json:"kcalPer100g"`
	Fats               *float64  `json:"fats,omitempty"`
	Carbs              *float64  `json:"carbs,omitempty"`
	Proteins           *float64  `json:"proteins,omitempty"`
	GlobalIngredientID *int      `json:"global_ingredient_id,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}