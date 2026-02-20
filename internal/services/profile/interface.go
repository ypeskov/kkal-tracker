package profile

// Servicer defines the profile service contract used by handlers.
type Servicer interface {
	GetProfile(userID int) (*ProfileResponse, error)
	UpdateProfile(userID int, req *ProfileUpdateRequest) error
	SetWeightGoal(userID int, req *WeightGoalRequest) error
	ClearWeightGoal(userID int) error
	GetWeightGoalProgress(userID int) (*WeightGoalResponse, error)
}
