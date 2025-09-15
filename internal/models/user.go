package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User is a pure data structure representing a user
type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FirstName    *string   `json:"first_name,omitempty"`
	LastName     *string   `json:"last_name,omitempty"`
	Age          *int      `json:"age,omitempty"`
	Height       *float64  `json:"height,omitempty"` // Height in cm
	Language     *string   `json:"language,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// SetPassword hashes and sets the user's password
// Business logic method - consider moving to service layer
func (u *User) SetPassword(password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u.PasswordHash = string(hashedPassword)
	return nil
}

// CheckPassword verifies a password against the hash
// Business logic method - consider moving to service layer
func (u *User) CheckPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password))
	return err == nil
}

// ProfileUpdateRequest represents the request to update user profile
type ProfileUpdateRequest struct {
	FirstName *string  `json:"first_name"`
	LastName  *string  `json:"last_name"`
	Email     string   `json:"email" validate:"required,email"`
	Age       *int     `json:"age" validate:"omitempty,min=1,max=150"`
	Height    *float64 `json:"height" validate:"omitempty,min=50,max=300"`
	Weight    *float64 `json:"weight" validate:"omitempty,min=10,max=500"`
	Language  string   `json:"language" validate:"required,oneof=en_US uk_UA ru_UA bg_BG"`
}

// ProfileResponse represents the user profile data returned to the client
type ProfileResponse struct {
	ID        int       `json:"id"`
	FirstName *string   `json:"first_name"`
	LastName  *string   `json:"last_name"`
	Email     string    `json:"email"`
	Age       *int      `json:"age"`
	Height    *float64  `json:"height"`
	Weight    *float64  `json:"weight"` // Latest weight from weight_history
	Language  string    `json:"language"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}