package dto

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
	FirstName *string  `json:"first_name"`
	LastName  *string  `json:"last_name"`
	Email     string   `json:"email"`
	Age       *int     `json:"age"`
	Height    *float64 `json:"height"`
	Weight    *float64 `json:"weight"`
	Language  string   `json:"language"`
}