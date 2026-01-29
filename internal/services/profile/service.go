package profile

import (
	"database/sql"
	"errors"
	"log/slog"
	"math"
	"time"

	"ypeskov/kkal-tracker/internal/repositories"
)

// Constants for weight goal calculations
const (
	KcalPerKg             = 7700.0 // Calories per kg of body weight
	SafeWeeklyLossKg      = 0.5    // Safe rate: 0.5 kg per week
	MaxDailyDeficit       = 1000.0 // Maximum recommended daily deficit
	MinDailyCaloriesMale  = 1500.0 // Minimum daily intake for men
	MinDailyCaloriesFemale = 1200.0 // Minimum daily intake for women
)

var (
	ErrNoWeightData     = errors.New("no weight data available")
	ErrInvalidGoal      = errors.New("invalid weight goal")
	ErrGoalNotSet       = errors.New("weight goal is not set")
	ErrTargetDateInPast = errors.New("target date must be in the future")
)

type Service struct {
	db             *sql.DB
	userRepo       repositories.UserRepository
	weightHistRepo repositories.WeightHistoryRepository
	logger         *slog.Logger
}

func New(db *sql.DB, userRepo repositories.UserRepository, weightHistRepo repositories.WeightHistoryRepository, logger *slog.Logger) *Service {
	return &Service{
		db:             db,
		userRepo:       userRepo,
		weightHistRepo: weightHistRepo,
		logger:         logger.With("service", "profile"),
	}
}

// GetProfile retrieves the user profile and returns a DTO
func (s *Service) GetProfile(userID int) (*ProfileResponse, error) {
	s.logger.Debug("GetProfile called", "user_id", userID)

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		s.logger.Error("Failed to get user", "user_id", userID, "error", err)
		return nil, err
	}

	// Default language if not set
	language := "en_US"
	if user.Language != nil {
		language = *user.Language
	}

	// Get latest weight from weight history
	var weight *float64
	latestWeight, err := s.weightHistRepo.GetLatestByUserID(userID)
	if err != nil {
		s.logger.Warn("Failed to get latest weight", "user_id", userID, "error", err)
		// Continue without weight data
	} else if latestWeight != nil {
		weight = &latestWeight.Weight
	}

	// Convert domain model to DTO
	response := &ProfileResponse{
		FirstName:           user.FirstName,
		LastName:            user.LastName,
		Email:               user.Email,
		Age:                 user.Age,
		Height:              user.Height,
		Weight:              weight,
		Gender:              user.Gender,
		Language:            language,
		ActivityLevel:       user.ActivityLevel,
		TargetWeight:        user.TargetWeight,
		TargetDate:          user.TargetDate,
		GoalSetAt:           user.GoalSetAt,
		InitialWeightAtGoal: user.InitialWeightAtGoal,
	}

	s.logger.Debug("GetProfile completed successfully", "user_id", userID, "email", response.Email)
	return response, nil
}

// UpdateProfile updates the user profile, passing DTO directly to repository
func (s *Service) UpdateProfile(userID int, req *ProfileUpdateRequest) error {
	s.logger.Debug("UpdateProfile called", "user_id", userID, "email", req.Email, "first_name", req.FirstName, "last_name", req.LastName)

	// Start transaction for atomic updates
	tx, err := s.db.Begin()
	if err != nil {
		s.logger.Error("Failed to begin transaction", "user_id", userID, "error", err)
		return err
	}
	defer tx.Rollback()

	// Update user profile (without weight - now managed only in weight history)
	if err := s.userRepo.UpdateProfile(userID, req.FirstName, req.LastName, req.Email, req.Age, req.Height, req.Gender, nil, req.Language, req.ActivityLevel); err != nil {
		s.logger.Error("Failed to update profile", "user_id", userID, "error", err)
		return err
	}

	// Weight is now managed only through weight history, not profile updates

	// Commit transaction
	if err := tx.Commit(); err != nil {
		s.logger.Error("Failed to commit transaction", "user_id", userID, "error", err)
		return err
	}

	s.logger.Debug("UpdateProfile completed successfully", "user_id", userID)
	return nil
}

// SetWeightGoal sets a weight goal for the user
func (s *Service) SetWeightGoal(userID int, req *WeightGoalRequest) error {
	s.logger.Debug("SetWeightGoal called", "user_id", userID, "target_weight", req.TargetWeight)

	// Validate target date if provided
	if req.TargetDate != nil && *req.TargetDate != "" {
		targetDate, err := time.Parse("2006-01-02", *req.TargetDate)
		if err != nil {
			s.logger.Error("Invalid target date format", "user_id", userID, "error", err)
			return ErrInvalidGoal
		}
		if targetDate.Before(time.Now().Truncate(24 * time.Hour)) {
			s.logger.Error("Target date is in the past", "user_id", userID)
			return ErrTargetDateInPast
		}
	}

	// Get current weight
	latestWeight, err := s.weightHistRepo.GetLatestByUserID(userID)
	if err != nil {
		s.logger.Error("Failed to get latest weight", "user_id", userID, "error", err)
		return ErrNoWeightData
	}
	if latestWeight == nil {
		s.logger.Error("No weight data available", "user_id", userID)
		return ErrNoWeightData
	}

	// Set the goal
	if err := s.userRepo.SetWeightGoal(userID, req.TargetWeight, req.TargetDate, latestWeight.Weight); err != nil {
		s.logger.Error("Failed to set weight goal", "user_id", userID, "error", err)
		return err
	}

	s.logger.Info("Weight goal set successfully", "user_id", userID, "target_weight", req.TargetWeight)
	return nil
}

// ClearWeightGoal clears the weight goal for the user
func (s *Service) ClearWeightGoal(userID int) error {
	s.logger.Debug("ClearWeightGoal called", "user_id", userID)

	if err := s.userRepo.ClearWeightGoal(userID); err != nil {
		s.logger.Error("Failed to clear weight goal", "user_id", userID, "error", err)
		return err
	}

	s.logger.Info("Weight goal cleared successfully", "user_id", userID)
	return nil
}

// GetWeightGoalProgress calculates and returns the weight goal progress
func (s *Service) GetWeightGoalProgress(userID int) (*WeightGoalResponse, error) {
	s.logger.Debug("GetWeightGoalProgress called", "user_id", userID)

	// Get user to check if goal is set
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		s.logger.Error("Failed to get user", "user_id", userID, "error", err)
		return nil, err
	}

	if user.TargetWeight == nil || user.GoalSetAt == nil || user.InitialWeightAtGoal == nil {
		return nil, ErrGoalNotSet
	}

	// Get current weight
	latestWeight, err := s.weightHistRepo.GetLatestByUserID(userID)
	if err != nil {
		s.logger.Error("Failed to get latest weight", "user_id", userID, "error", err)
		return nil, ErrNoWeightData
	}
	if latestWeight == nil {
		return nil, ErrNoWeightData
	}

	currentWeight := latestWeight.Weight
	targetWeight := *user.TargetWeight
	initialWeight := *user.InitialWeightAtGoal

	// Determine if gaining or losing weight
	isGaining := targetWeight > initialWeight

	// Calculate progress
	var progressPercent float64
	var weightLost float64
	var weightToGo float64

	if isGaining {
		// Gaining weight
		totalToGain := targetWeight - initialWeight
		gained := currentWeight - initialWeight
		weightLost = -gained // Negative indicates weight gained
		weightToGo = targetWeight - currentWeight

		if totalToGain > 0 {
			progressPercent = (gained / totalToGain) * 100
		}
	} else {
		// Losing weight
		totalToLose := initialWeight - targetWeight
		lost := initialWeight - currentWeight
		weightLost = lost
		weightToGo = currentWeight - targetWeight

		if totalToLose > 0 {
			progressPercent = (lost / totalToLose) * 100
		}
	}

	// Clamp progress to 0-100
	progressPercent = math.Max(0, math.Min(100, progressPercent))

	response := &WeightGoalResponse{
		TargetWeight:        targetWeight,
		TargetDate:          user.TargetDate,
		GoalSetAt:           *user.GoalSetAt,
		InitialWeightAtGoal: initialWeight,
		CurrentWeight:       currentWeight,
		ProgressPercent:     math.Round(progressPercent*10) / 10, // Round to 1 decimal
		WeightToGo:          math.Round(weightToGo*100) / 100,    // Round to 2 decimals
		WeightLost:          math.Round(weightLost*100) / 100,    // Round to 2 decimals
		IsGaining:           isGaining,
	}

	// Calculate days remaining and daily deficit if target date is set
	if user.TargetDate != nil {
		daysRemaining := int(math.Ceil(time.Until(*user.TargetDate).Hours() / 24))
		if daysRemaining < 0 {
			daysRemaining = 0
		}
		response.DaysRemaining = &daysRemaining

		if daysRemaining > 0 && weightToGo > 0 {
			// Calculate required daily deficit
			totalKcal := math.Abs(weightToGo) * KcalPerKg
			dailyDeficit := totalKcal / float64(daysRemaining)
			dailyDeficit = math.Round(dailyDeficit)
			response.DailyDeficitNeeded = &dailyDeficit
		}
	} else {
		// No target date - estimate completion at safe rate
		if weightToGo > 0 {
			weeksNeeded := math.Abs(weightToGo) / SafeWeeklyLossKg
			daysNeeded := int(math.Ceil(weeksNeeded * 7))
			estimatedCompletion := time.Now().AddDate(0, 0, daysNeeded)
			response.EstimatedCompletion = &estimatedCompletion
		}
	}

	s.logger.Debug("GetWeightGoalProgress completed", "user_id", userID, "progress", progressPercent)
	return response, nil
}
