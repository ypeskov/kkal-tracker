package reports

import (
	"log/slog"
	"time"

	"ypeskov/kkal-tracker/internal/models"
	calorieservice "ypeskov/kkal-tracker/internal/services/calorie"
	weightservice "ypeskov/kkal-tracker/internal/services/weight"
)

const defaultPeriodDays = 30

type Service struct {
	calorieService *calorieservice.Service
	weightService  *weightservice.Service
	logger         *slog.Logger
}

func New(calorieService *calorieservice.Service, weightService *weightservice.Service, logger *slog.Logger) *Service {
	return &Service{
		calorieService: calorieService,
		weightService:  weightService,
		logger:         logger.With("service", "reports"),
	}
}

// GetAggregatedMetrics retrieves and aggregates weight and calorie metrics for the specified date range
func (s *Service) GetAggregatedMetrics(userID int, dateFrom, dateTo string) (*ReportDataResponse, error) {
	s.logger.Debug("GetAggregatedMetrics called", "user_id", userID, "from", dateFrom, "to", dateTo)

	// Apply default date range if not provided
	if dateFrom == "" || dateTo == "" {
		dateTo = time.Now().Format("2006-01-02")
		dateFrom = time.Now().AddDate(0, 0, -defaultPeriodDays).Format("2006-01-02")
	}

	// Get weight history
	weightHistory, err := s.weightService.GetWeightHistoryByDateRange(userID, dateFrom, dateTo)
	if err != nil {
		s.logger.Error("failed to get weight history", "error", err)
		return nil, err
	}

	// Get calorie entries
	calorieEntries, err := s.calorieService.GetEntriesByDateRange(userID, dateFrom, dateTo)
	if err != nil {
		s.logger.Error("failed to get calorie entries", "error", err)
		return nil, err
	}

	// Process weight data
	weightPoints := s.aggregateWeightData(weightHistory)

	// Process calorie data
	caloriePoints := s.aggregateCalorieData(calorieEntries)

	return &ReportDataResponse{
		WeightHistory:  weightPoints,
		CalorieHistory: caloriePoints,
	}, nil
}

// aggregateWeightData processes weight history and calculates daily averages
func (s *Service) aggregateWeightData(weightHistory []*models.WeightHistory) []WeightDataPoint {
	// Group weights by day
	weightByDay := make(map[string][]float64)
	for _, w := range weightHistory {
		day := w.RecordedAt.Format("2006-01-02")
		weightByDay[day] = append(weightByDay[day], w.Weight)
	}

	// Calculate average weight per day
	weightPoints := make([]WeightDataPoint, 0, len(weightByDay))
	for date, weights := range weightByDay {
		avgWeight := s.calculateAverage(weights)
		weightPoints = append(weightPoints, WeightDataPoint{
			Date:   date,
			Weight: avgWeight,
		})
	}

	return weightPoints
}

// aggregateCalorieData processes calorie entries and calculates daily totals
func (s *Service) aggregateCalorieData(calorieEntries []*models.CalorieEntry) []CalorieDataPoint {
	// Sum calories by day
	caloriesByDay := make(map[string]int)
	for _, entry := range calorieEntries {
		day := entry.MealDatetime.Format("2006-01-02")
		caloriesByDay[day] += entry.Calories
	}

	// Convert to slice
	caloriePoints := make([]CalorieDataPoint, 0, len(caloriesByDay))
	for date, calories := range caloriesByDay {
		caloriePoints = append(caloriePoints, CalorieDataPoint{
			Date:     date,
			Calories: calories,
		})
	}

	return caloriePoints
}

// calculateAverage calculates the average of a slice of float64 values
func (s *Service) calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	var sum float64
	for _, value := range values {
		sum += value
	}
	return sum / float64(len(values))
}