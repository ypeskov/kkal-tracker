package models

import "time"

// ActivationToken represents an email activation token
type ActivationToken struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// IsExpired checks if the token has expired
func (t *ActivationToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}
