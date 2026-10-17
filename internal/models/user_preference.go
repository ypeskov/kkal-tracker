package models

import "time"

// UserPreference is a stored UI preference of a user: one value per key, JSON encoded by the server
type UserPreference struct {
	UserID    int       `json:"user_id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}
