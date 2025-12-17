package ai

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"ypeskov/kkal-tracker/internal/models"
	"ypeskov/kkal-tracker/internal/repositories"
	aiservice "ypeskov/kkal-tracker/internal/services/ai"
	calorieservice "ypeskov/kkal-tracker/internal/services/calorie"
	weightservice "ypeskov/kkal-tracker/internal/services/weight"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	aiService      *aiservice.Service
	calorieService *calorieservice.Service
	weightService  *weightservice.Service
	userRepo       repositories.UserRepository
	logger         *slog.Logger
}

func New(
	aiService *aiservice.Service,
	calorieService *calorieservice.Service,
	weightService *weightservice.Service,
	userRepo repositories.UserRepository,
	logger *slog.Logger,
) *Handler {
	return &Handler{
		aiService:      aiService,
		calorieService: calorieService,
		weightService:  weightService,
		userRepo:       userRepo,
		logger:         logger.With("handler", "ai"),
	}
}

// GetProviders returns the list of available (active) AI providers
func (h *Handler) GetProviders(c echo.Context) error {
	providers, err := h.aiService.GetAvailableProviders()
	if err != nil {
		h.logger.Error("Failed to get providers", slog.String("error", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get providers")
	}

	response := ProvidersResponse{
		Providers: make([]ProviderResponse, 0, len(providers)),
	}

	for _, p := range providers {
		response.Providers = append(response.Providers, ProviderResponse{
			ID:          p.ID,
			DisplayName: p.DisplayName,
			Model:       p.Model,
		})
	}

	return c.JSON(http.StatusOK, response)
}

// Analyze performs AI analysis on user's nutrition and weight data
func (h *Handler) Analyze(c echo.Context) error {
	userID := c.Get("user_id").(int)

	// Get user profile for context
	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		h.logger.Error("Failed to get user", slog.String("error", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get user profile")
	}

	var req AnalyzeRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	h.logger.Debug("AI analysis requested",
		slog.Int("user_id", userID),
		slog.String("provider", req.Provider),
		slog.Int("period_days", req.PeriodDays))

	// Calculate date range
	dateTo := time.Now().Format("2006-01-02")
	dateFrom := time.Now().AddDate(0, 0, -req.PeriodDays).Format("2006-01-02")

	// Fetch nutrition data
	calorieEntries, err := h.calorieService.GetEntriesByDateRange(userID, dateFrom, dateTo)
	if err != nil {
		h.logger.Error("Failed to get calorie entries", slog.String("error", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch nutrition data")
	}

	// Fetch weight data
	weightHistory, err := h.weightService.GetWeightHistoryByDateRange(userID, dateFrom, dateTo)
	if err != nil {
		h.logger.Error("Failed to get weight history", slog.String("error", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch weight data")
	}

	// Aggregate nutrition data by day
	nutritionData := h.aggregateNutritionData(calorieEntries)

	// Convert weight data
	weightData := h.convertWeightData(weightHistory)

	// Build user context
	userContext := aiservice.UserContext{
		Age:      user.Age,
		Height:   user.Height,
		Language: "en_US",
	}
	if user.Language != nil {
		userContext.Language = *user.Language
	}

	// Build analysis request
	analysisReq := aiservice.AnalysisRequest{
		UserContext:   userContext,
		NutritionData: nutritionData,
		WeightData:    weightData,
		Query:         req.Query,
		PeriodDays:    req.PeriodDays,
	}

	// Perform analysis
	result, err := h.aiService.Analyze(c.Request().Context(), req.Provider, analysisReq)
	if err != nil {
		if errors.Is(err, aiservice.ErrProviderNotAvailable) {
			return echo.NewHTTPError(http.StatusBadRequest, "Selected AI provider is not available")
		}
		if errors.Is(err, aiservice.ErrProviderNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "AI provider not found")
		}
		h.logger.Error("AI analysis failed", slog.String("error", err.Error()))
		return echo.NewHTTPError(http.StatusInternalServerError, "AI analysis failed")
	}

	return c.JSON(http.StatusOK, AnalysisResultResponse{
		Analysis:   result.Analysis,
		Provider:   result.Provider,
		Model:      result.Model,
		TokensUsed: result.TokensUsed,
		DurationMs: result.DurationMs,
	})
}

// aggregateNutritionData groups calorie entries by day
func (h *Handler) aggregateNutritionData(entries []*models.CalorieEntry) []aiservice.NutritionDataPoint {
	// Group by date
	byDate := make(map[string]*aiservice.NutritionDataPoint)

	for _, entry := range entries {
		date := entry.MealDatetime.Format("2006-01-02")

		if _, exists := byDate[date]; !exists {
			byDate[date] = &aiservice.NutritionDataPoint{
				Date: date,
			}
		}

		point := byDate[date]
		point.Calories += entry.Calories

		// Calculate macros based on weight
		if entry.Fats != nil {
			point.Fats += *entry.Fats * entry.Weight / 100
		}
		if entry.Carbs != nil {
			point.Carbs += *entry.Carbs * entry.Weight / 100
		}
		if entry.Proteins != nil {
			point.Proteins += *entry.Proteins * entry.Weight / 100
		}
	}

	// Convert map to slice
	result := make([]aiservice.NutritionDataPoint, 0, len(byDate))
	for _, point := range byDate {
		result = append(result, *point)
	}

	return result
}

// convertWeightData converts weight history to AI service format
func (h *Handler) convertWeightData(history []*models.WeightHistory) []aiservice.WeightDataPoint {
	result := make([]aiservice.WeightDataPoint, 0, len(history))

	for _, w := range history {
		result = append(result, aiservice.WeightDataPoint{
			Date:   w.RecordedAt.Format("2006-01-02"),
			Weight: w.Weight,
		})
	}

	return result
}

// RegisterRoutes registers AI handler routes
func (h *Handler) RegisterRoutes(g *echo.Group) {
	g.GET("/providers", h.GetProviders)
	g.POST("/analyze", h.Analyze)
}
