package models

import (
	"time"
)

// WeightHistory represents a weight measurement for a user at a specific point in time
type WeightHistory struct {
	ID         int       `json:"id"`
	UserID     int       `json:"user_id"`
	Weight     float64   `json:"weight"` // Weight in kg
	RecordedAt time.Time `json:"recorded_at"`
	CreatedAt  time.Time `json:"created_at"`
}