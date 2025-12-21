package ai

import "errors"

// Service errors
var (
	ErrProviderNotAvailable = errors.New("AI provider is not available")
	ErrAnalysisFailed       = errors.New("AI analysis failed")
)

// FoodItem represents a single food entry for AI analysis
type FoodItem struct {
	Name     string  `json:"name"`
	Weight   float64 `json:"weight"`
	Calories int     `json:"calories"`
}

// NutritionDataPoint represents daily nutrition data for AI analysis
type NutritionDataPoint struct {
	Date      string     `json:"date"`
	Calories  int        `json:"calories"`
	Fats      float64    `json:"fats"`
	Carbs     float64    `json:"carbs"`
	Proteins  float64    `json:"proteins"`
	FoodItems []FoodItem `json:"food_items"`
}

// WeightDataPoint represents a weight measurement for AI analysis
type WeightDataPoint struct {
	Date   string  `json:"date"`
	Weight float64 `json:"weight"`
}

// UserContext provides user profile information for personalized analysis
type UserContext struct {
	Age      *int     `json:"age,omitempty"`
	Height   *float64 `json:"height,omitempty"`
	Language string   `json:"language"`
}

// AnalysisRequest represents a request for AI analysis
type AnalysisRequest struct {
	UserContext   UserContext          `json:"user_context"`
	NutritionData []NutritionDataPoint `json:"nutrition_data"`
	WeightData    []WeightDataPoint    `json:"weight_data"`
	Query         string               `json:"query,omitempty"` // Currently ignored; reserved for future use
	PeriodDays    int                  `json:"period_days"`
}

// AnalysisResponse represents the AI analysis result
type AnalysisResponse struct {
	Analysis   string `json:"analysis"`
	Model      string `json:"model"`
	TokensUsed int    `json:"tokens_used,omitempty"`
	DurationMs int64  `json:"duration_ms"`
}
