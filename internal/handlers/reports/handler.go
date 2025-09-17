package reports

import (
	"log/slog"
	"net/http"
	"time"

	calorieservice "ypeskov/kkal-tracker/internal/services/calorie"
	weightservice "ypeskov/kkal-tracker/internal/services/weight"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	calorieService *calorieservice.Service
	weightService  *weightservice.Service
	logger         *slog.Logger
}

func New(calorieService *calorieservice.Service, weightService *weightservice.Service, logger *slog.Logger) *Handler {
	return &Handler{
		calorieService: calorieService,
		weightService:  weightService,
		logger:         logger.With("handler", "reports"),
	}
}

// GetReportData returns combined weight and calorie data for the specified date range
func (h *Handler) GetReportData(c echo.Context) error {
	h.logger.Debug("GetReportData called")
	userID := c.Get("user_id").(int)
	dateFrom := c.QueryParam("from")
	dateTo := c.QueryParam("to")

	// Default to last 30 days if no dates provided
	if dateFrom == "" || dateTo == "" {
		dateTo = time.Now().Format("2006-01-02")
		dateFrom = time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	}

	h.logger.Debug("GetReportData called",
		"user_id", userID,
		"from", dateFrom,
		"to", dateTo)

	// Get weight history
	weightHistory, err := h.weightService.GetWeightHistoryByDateRange(userID, dateFrom, dateTo)
	if err != nil {
		h.logger.Error("failed to get weight history", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get report data")
	}

	// Get calorie entries
	calorieEntries, err := h.calorieService.GetEntriesByDateRange(userID, dateFrom, dateTo)
	if err != nil {
		h.logger.Error("failed to get calorie entries", "error", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get report data")
	}

	// Aggregate weight data by day (calculate average if multiple entries per day)
	weightByDay := make(map[string][]float64)
	for _, w := range weightHistory {
		day := w.RecordedAt.Format("2006-01-02")
		weightByDay[day] = append(weightByDay[day], w.Weight)
	}

	// Calculate average weight per day
	weightPoints := make([]WeightDataPoint, 0, len(weightByDay))
	for date, weights := range weightByDay {
		var sum float64
		for _, weight := range weights {
			sum += weight
		}
		avgWeight := sum / float64(len(weights))

		weightPoints = append(weightPoints, WeightDataPoint{
			Date:   date,
			Weight: avgWeight,
		})
	}

	// Aggregate calories by day
	caloriesByDay := make(map[string]int)
	for _, entry := range calorieEntries {
		day := entry.MealDatetime.Format("2006-01-02")
		caloriesByDay[day] += entry.Calories
	}

	// Convert to sorted slice
	caloriePoints := make([]CalorieDataPoint, 0, len(caloriesByDay))
	for date, calories := range caloriesByDay {
		caloriePoints = append(caloriePoints, CalorieDataPoint{
			Date:     date,
			Calories: calories,
		})
	}

	response := ReportDataResponse{
		WeightHistory:  weightPoints,
		CalorieHistory: caloriePoints,
	}

	h.logger.Debug("GetReportData end")
	return c.JSON(http.StatusOK, response)
}

// RegisterRoutes registers all report-related routes
func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.GET("/data", h.GetReportData)
}
