package profile

import "time"

// ProfileUpdateRequest represents the request to update user profile
type ProfileUpdateRequest struct {
	FirstName     *string  `json:"first_name"`
	LastName      *string  `json:"last_name"`
	Email         string   `json:"email" validate:"required,email"`
	Age           *int     `json:"age" validate:"omitempty,min=1,max=150"`
	Height        *float64 `json:"height" validate:"omitempty,min=50,max=300"`
	Gender        *string  `json:"gender" validate:"omitempty,oneof=male female"`
	Language      string   `json:"language" validate:"required,oneof=en_US uk_UA ru_UA bg_BG"`
	ActivityLevel *string  `json:"activity_level" validate:"omitempty,oneof=sedentary lightly_active moderate very_active extra_active"`
}

// ProfileResponse represents the user profile data returned to the client
type ProfileResponse struct {
	FirstName     *string  `json:"first_name"`
	LastName      *string  `json:"last_name"`
	Email         string   `json:"email"`
	Age           *int     `json:"age"`
	Height        *float64 `json:"height"`
	Weight        *float64 `json:"weight"`
	Gender        *string  `json:"gender"`
	Language      string   `json:"language"`
	ActivityLevel *string  `json:"activity_level"`

	// Weight goal fields
	TargetWeight        *float64   `json:"target_weight,omitempty"`
	TargetDate          *time.Time `json:"target_date,omitempty"`
	GoalSetAt           *time.Time `json:"goal_set_at,omitempty"`
	InitialWeightAtGoal *float64   `json:"initial_weight_at_goal,omitempty"`
}

// WeightGoalRequest represents the request to set a weight goal
type WeightGoalRequest struct {
	TargetWeight float64 `json:"target_weight" validate:"required,min=30,max=300"`
	TargetDate   *string `json:"target_date" validate:"omitempty"` // Format: YYYY-MM-DD, optional
}

// WeightGoalResponse represents the weight goal progress data
type WeightGoalResponse struct {
	TargetWeight        float64    `json:"target_weight"`
	TargetDate          *time.Time `json:"target_date,omitempty"`
	GoalSetAt           time.Time  `json:"goal_set_at"`
	InitialWeightAtGoal float64    `json:"initial_weight_at_goal"`
	CurrentWeight       float64    `json:"current_weight"`
	ProgressPercent     float64    `json:"progress_percent"`     // 0-100
	WeightToGo          float64    `json:"weight_to_go"`         // Remaining weight to lose/gain
	WeightLost          float64    `json:"weight_lost"`          // Weight already lost/gained
	DaysRemaining       *int       `json:"days_remaining"`       // Days until target date (if set)
	DailyDeficitNeeded  *float64   `json:"daily_deficit_needed"` // Required daily deficit (if target date set)
	EstimatedCompletion *time.Time `json:"estimated_completion"` // If no target date, estimated at 0.5kg/week
	IsGaining           bool       `json:"is_gaining"`           // True if gaining weight, false if losing
}
