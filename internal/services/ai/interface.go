package ai

import "context"

// Servicer defines the AI service contract used by handlers.
type Servicer interface {
	Analyze(ctx context.Context, req AnalysisRequest) (*AnalysisResponse, error)
	IsAvailable() bool
	GetModel() string
}
