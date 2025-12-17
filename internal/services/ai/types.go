package ai

import "errors"

// ProviderType represents supported AI provider types
type ProviderType string

const (
	ProviderOpenAI    ProviderType = "openai"
	ProviderAnthropic ProviderType = "anthropic"
	ProviderOllama    ProviderType = "ollama"
)

// IsValid checks if the provider type is supported
func (p ProviderType) IsValid() bool {
	switch p {
	case ProviderOpenAI, ProviderAnthropic, ProviderOllama:
		return true
	}
	return false
}

// AllProviders returns all supported provider types
func AllProviders() []ProviderType {
	return []ProviderType{
		ProviderOpenAI,
		ProviderAnthropic,
		ProviderOllama,
	}
}

// Service errors
var (
	ErrProviderNotAvailable = errors.New("AI provider is not available")
	ErrProviderNotFound     = errors.New("AI provider not found")
	ErrNoActiveProviders    = errors.New("no active AI providers configured")
	ErrAnalysisFailed       = errors.New("AI analysis failed")
)

// NutritionDataPoint represents daily nutrition data for AI analysis
type NutritionDataPoint struct {
	Date     string  `json:"date"`
	Calories int     `json:"calories"`
	Fats     float64 `json:"fats"`
	Carbs    float64 `json:"carbs"`
	Proteins float64 `json:"proteins"`
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
	Query         string               `json:"query,omitempty"`
	PeriodDays    int                  `json:"period_days"`
}

// AnalysisResponse represents the AI analysis result
type AnalysisResponse struct {
	Analysis   string `json:"analysis"`
	Provider   string `json:"provider"`
	Model      string `json:"model"`
	TokensUsed int    `json:"tokens_used,omitempty"`
	DurationMs int64  `json:"duration_ms"`
}
