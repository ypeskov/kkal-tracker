package ai

// AnalyzeRequest represents the request body for AI analysis
type AnalyzeRequest struct {
	PeriodDays int    `json:"period_days" validate:"required,min=1,max=365"`
	Query      string `json:"query,omitempty"` // Currently ignored; reserved for future use
}

// StatusResponse represents the AI service status
type StatusResponse struct {
	Available bool   `json:"available"`
	Model     string `json:"model,omitempty"`
}

// AnalysisResultResponse represents the AI analysis result
type AnalysisResultResponse struct {
	Analysis   string `json:"analysis"`
	Model      string `json:"model"`
	TokensUsed int    `json:"tokens_used,omitempty"`
	DurationMs int64  `json:"duration_ms"`
}
