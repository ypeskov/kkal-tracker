package ai

// AnalyzeRequest represents the request body for AI analysis
type AnalyzeRequest struct {
	Provider   string `json:"provider" validate:"required"`
	PeriodDays int    `json:"period_days" validate:"required,min=1,max=365"`
	Query      string `json:"query,omitempty"`
}

// ProviderResponse represents a single AI provider in API responses
type ProviderResponse struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Model       string `json:"model"`
}

// ProvidersResponse represents the response for listing providers
type ProvidersResponse struct {
	Providers []ProviderResponse `json:"providers"`
}

// AnalysisResultResponse represents the AI analysis result
type AnalysisResultResponse struct {
	Analysis   string `json:"analysis"`
	Provider   string `json:"provider"`
	Model      string `json:"model"`
	TokensUsed int    `json:"tokens_used,omitempty"`
	DurationMs int64  `json:"duration_ms"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error"`
}
