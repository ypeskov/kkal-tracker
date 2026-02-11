package models

import "time"

// APIKey represents a user's API key for external data access
type APIKey struct {
	ID        int        `json:"id"`
	UserID    int        `json:"user_id"`
	Name      string     `json:"name"`
	KeyHash   string     `json:"-"`
	KeyPrefix string     `json:"key_prefix"`
	ExpiresAt *time.Time `json:"expires_at"`
	IsRevoked bool       `json:"is_revoked"`
	CreatedAt time.Time  `json:"created_at"`
}

// IsExpired checks if the API key has expired (permanent keys never expire)
func (k *APIKey) IsExpired() bool {
	if k.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*k.ExpiresAt)
}

// IsValid checks if the API key is neither expired nor revoked
func (k *APIKey) IsValid() bool {
	return !k.IsRevoked && !k.IsExpired()
}
