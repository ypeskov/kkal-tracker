package ai

import (
	"context"
	"log/slog"

	"ypeskov/kkal-tracker/internal/config"
)

// Service handles AI-related business logic
type Service struct {
	config   *config.Config
	provider Provider
	logger   *slog.Logger
}

// New creates a new AI service instance
func New(cfg *config.Config, logger *slog.Logger) *Service {
	svc := &Service{
		config: cfg,
		logger: logger.With("service", "ai"),
	}
	svc.initProvider()
	return svc
}

// initProvider initializes the OpenAI provider based on configuration
func (s *Service) initProvider() {
	if s.config.AI.APIKey != "" {
		s.provider = NewOpenAIProvider(
			s.config.AI.APIKey,
			s.config.AI.BaseURL,
			s.config.AI.UseMaxTokens,
			s.config.AI.MaxTokens,
			s.logger,
		)
		s.logger.Info("OpenAI provider initialized",
			slog.String("model", s.config.AI.Model))
	} else {
		s.logger.Debug("OpenAI provider not configured (no API key)")
	}
}

// Analyze performs AI analysis using the configured provider
func (s *Service) Analyze(ctx context.Context, req AnalysisRequest) (*AnalysisResponse, error) {
	s.logger.Debug("Starting AI analysis",
		slog.Int("period_days", req.PeriodDays))

	if s.provider == nil {
		return nil, ErrProviderNotAvailable
	}

	// Perform analysis using the model from config
	response, err := s.provider.Analyze(ctx, s.config.AI.Model, req)
	if err != nil {
		s.logger.Error("Analysis failed",
			slog.String("model", s.config.AI.Model),
			slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.Info("AI analysis completed",
		slog.String("model", s.config.AI.Model),
		slog.Int("tokens", response.TokensUsed),
		slog.Int64("duration_ms", response.DurationMs))

	return response, nil
}

// IsAvailable checks if the AI provider is configured and available
func (s *Service) IsAvailable() bool {
	return s.provider != nil
}

// GetModel returns the configured model name
func (s *Service) GetModel() string {
	return s.config.AI.Model
}
