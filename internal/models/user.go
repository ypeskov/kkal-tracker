package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
)

// User is a pure data structure representing a user
type User struct {
	ID            int       `json:"id"`
	Email         string    `json:"email"`
	PasswordHash  string    `json:"-"`
	IsActive      bool      `json:"is_active"`
	FirstName     *string   `json:"first_name,omitempty"`
	LastName      *string   `json:"last_name,omitempty"`
	Age           *int      `json:"age,omitempty"`
	Height        *float64  `json:"height,omitempty"` // Height in cm
	Gender        *string   `json:"gender,omitempty"` // Gender: "male" or "female"
	Language      *string   `json:"language,omitempty"`
	ActivityLevel *string   `json:"activity_level,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// Weight goal fields
	TargetWeight        *float64   `json:"target_weight,omitempty"`         // Target weight in kg
	TargetDate          *time.Time `json:"target_date,omitempty"`           // Optional target date
	GoalSetAt           *time.Time `json:"goal_set_at,omitempty"`           // When the goal was set
	InitialWeightAtGoal *float64   `json:"initial_weight_at_goal,omitempty"` // Weight when goal was set
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
