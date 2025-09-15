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
	Weight       *float64  `json:"weight,omitempty"` // Weight in kg
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


