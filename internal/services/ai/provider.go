package ai

import "context"

// Provider defines the interface for AI providers
type Provider interface {
	// Analyze performs nutrition and weight analysis
	Analyze(ctx context.Context, model string, req AnalysisRequest) (*AnalysisResponse, error)

	// GetProviderName returns the provider identifier
	GetProviderName() string

	// IsAvailable checks if the provider is properly configured and accessible
	IsAvailable() bool
}
