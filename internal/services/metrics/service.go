package metrics

import (
	"fmt"
	"log/slog"
	"math"

	"ypeskov/kkal-tracker/internal/repositories"
)

type Service struct {
	userRepo   repositories.UserRepository
	weightRepo repositories.WeightHistoryRepository
	logger     *slog.Logger
}

func New(userRepo repositories.UserRepository, weightRepo repositories.WeightHistoryRepository, logger *slog.Logger) *Service {
	return &Service{
		userRepo:   userRepo,
		weightRepo: weightRepo,
		logger:     logger.With("service", "metrics"),
	}
}

// GetHealthMetrics calculates and returns health metrics for a user
func (s *Service) GetHealthMetrics(userID int) (*HealthMetrics, error) {
	s.logger.Debug("GetHealthMetrics called", "user_id", userID)

	// Get user data
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		s.logger.Error("Failed to get user", "user_id", userID, "error", err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Get latest weight
	weights, err := s.weightRepo.GetByUserID(userID)
	if err != nil {
		s.logger.Error("Failed to get weight history", "user_id", userID, "error", err)
		return nil, fmt.Errorf("failed to get weight history: %w", err)
	}

	var currentWeight *float64
	if len(weights) > 0 {
		currentWeight = &weights[0].Weight
	}

	metrics := &HealthMetrics{}

	// Calculate BMI if we have height and weight
	if user.Height != nil && currentWeight != nil {
		bmi := s.calculateBMI(*currentWeight, *user.Height)
		metrics.BMI = &bmi
		metrics.BMICategory = s.getBMICategory(bmi)
	}

	// Calculate BMR if we have all required data
	if user.Age != nil && user.Height != nil && currentWeight != nil && user.Gender != nil {
		bmr := s.calculateBMR(*currentWeight, *user.Height, *user.Age, *user.Gender)
		metrics.BMR = &bmr

		// Calculate TDEE with sedentary activity level as default
		tdee := bmr * ActivityMultipliers[ActivitySedentary]
		metrics.TDEE = &tdee
	}

	// Generate health status message
	metrics.HealthStatus = s.generateHealthStatus(metrics)

	s.logger.Debug("Health metrics calculated", "user_id", userID, "bmi", metrics.BMI, "bmr", metrics.BMR)
	return metrics, nil
}

// calculateBMI calculates Body Mass Index
// BMI = weight (kg) / (height (m))^2
func (s *Service) calculateBMI(weightKg, heightCm float64) float64 {
	heightM := heightCm / 100.0
	bmi := weightKg / (heightM * heightM)
	return math.Round(bmi*10) / 10 // Round to 1 decimal place
}

// getBMICategory returns the BMI category based on WHO classification
func (s *Service) getBMICategory(bmi float64) string {
	switch {
	case bmi < 16.0:
		return "severely_underweight"
	case bmi < 18.5:
		return "underweight"
	case bmi < 25.0:
		return "normal"
	case bmi < 30.0:
		return "overweight"
	case bmi < 35.0:
		return "obese_class_1"
	case bmi < 40.0:
		return "obese_class_2"
	default:
		return "obese_class_3"
	}
}

// calculateBMR calculates Basal Metabolic Rate using Mifflin-St Jeor Equation
// Men: BMR = (10 × weight in kg) + (6.25 × height in cm) - (5 × age in years) + 5
// Women: BMR = (10 × weight in kg) + (6.25 × height in cm) - (5 × age in years) - 161
func (s *Service) calculateBMR(weightKg, heightCm float64, age int, gender string) float64 {
	bmr := (10 * weightKg) + (6.25 * heightCm) - (5 * float64(age))

	if gender == "male" {
		bmr += 5
	} else {
		bmr -= 161
	}

	return math.Round(bmr)
}

// CalculateTDEE calculates Total Daily Energy Expenditure
func (s *Service) CalculateTDEE(userID int, activityLevel ActivityLevel) (*float64, error) {
	metrics, err := s.GetHealthMetrics(userID)
	if err != nil {
		return nil, err
	}

	if metrics.BMR == nil {
		return nil, fmt.Errorf("cannot calculate TDEE: BMR not available")
	}

	multiplier, ok := ActivityMultipliers[activityLevel]
	if !ok {
		multiplier = ActivityMultipliers[ActivitySedentary]
	}

	tdee := *metrics.BMR * multiplier
	tdee = math.Round(tdee)
	return &tdee, nil
}

// generateHealthStatus generates a health status message based on metrics
func (s *Service) generateHealthStatus(metrics *HealthMetrics) string {
	if metrics.BMI == nil {
		return "incomplete_data"
	}

	category := metrics.BMICategory
	switch category {
	case "severely_underweight", "underweight":
		return "below_healthy_weight"
	case "normal":
		return "healthy_weight"
	case "overweight":
		return "above_healthy_weight"
	default:
		return "significantly_above_healthy_weight"
	}
}

// GetBMIHistory calculates BMI history based on weight history
func (s *Service) GetBMIHistory(userID int) ([]map[string]interface{}, error) {
	s.logger.Debug("GetBMIHistory called", "user_id", userID)

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if user.Height == nil {
		return nil, fmt.Errorf("user height not set")
	}

	weights, err := s.weightRepo.GetByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get weight history: %w", err)
	}

	history := make([]map[string]interface{}, 0, len(weights))
	for _, weight := range weights {
		bmi := s.calculateBMI(weight.Weight, *user.Height)
		history = append(history, map[string]interface{}{
			"date":         weight.RecordedAt,
			"weight":       weight.Weight,
			"bmi":          bmi,
			"bmi_category": s.getBMICategory(bmi),
		})
	}

	return history, nil
}
