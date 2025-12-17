package ai

import (
	"context"
	"log/slog"

	"ypeskov/kkal-tracker/internal/config"
	"ypeskov/kkal-tracker/internal/models"
	"ypeskov/kkal-tracker/internal/repositories"
)

// Service handles AI-related business logic
type Service struct {
	config    *config.Config
	repo      repositories.AIProviderRepository
	providers map[ProviderType]Provider
	logger    *slog.Logger
}

// New creates a new AI service instance
func New(cfg *config.Config, repo repositories.AIProviderRepository, logger *slog.Logger) *Service {
	svc := &Service{
		config:    cfg,
		repo:      repo,
		providers: make(map[ProviderType]Provider),
		logger:    logger.With("service", "ai"),
	}
	svc.initProviders()
	return svc
}

// initProviders initializes available AI providers based on configuration
func (s *Service) initProviders() {
	// Initialize OpenAI if API key is configured
	if s.config.AI.OpenAI.APIKey != "" {
		s.providers[ProviderOpenAI] = NewOpenAIProvider(
			s.config.AI.OpenAI.APIKey,
			s.config.AI.OpenAI.BaseURL,
			s.config.AI.UseMaxTokens,
			s.config.AI.MaxTokens,
			s.logger,
		)
		// Activate all OpenAI-type providers in the database
		if err := s.repo.SetActiveByType(string(ProviderOpenAI), true); err != nil {
			s.logger.Error("Failed to set OpenAI providers active", slog.String("error", err.Error()))
		} else {
			s.logger.Info("OpenAI providers initialized and activated")
		}
	} else {
		s.logger.Debug("OpenAI provider not configured (no API key)")
	}

	// TODO: Add Anthropic and Ollama initialization when implemented
}

// GetAvailableProviders returns all active providers from the database
func (s *Service) GetAvailableProviders() ([]*models.AIProvider, error) {
	return s.repo.GetActive()
}

// GetAllProviders returns all providers from the database
func (s *Service) GetAllProviders() ([]*models.AIProvider, error) {
	return s.repo.GetAll()
}

// Analyze performs AI analysis using the specified provider
func (s *Service) Analyze(ctx context.Context, providerID string, req AnalysisRequest) (*AnalysisResponse, error) {
	s.logger.Debug("Starting AI analysis",
		slog.String("provider_id", providerID),
		slog.Int("period_days", req.PeriodDays))

	// Get provider config from database for model and provider type
	dbProvider, err := s.repo.GetByID(providerID)
	if err != nil {
		s.logger.Error("Failed to get provider from database", slog.String("error", err.Error()))
		return nil, ErrProviderNotFound
	}

	if !dbProvider.IsActive {
		return nil, ErrProviderNotAvailable
	}

	// Check if provider implementation is available in memory using provider_type
	provider, exists := s.providers[ProviderType(dbProvider.ProviderType)]
	if !exists {
		s.logger.Warn("Provider not available", slog.String("provider_type", dbProvider.ProviderType))
		return nil, ErrProviderNotAvailable
	}

	// Perform analysis
	response, err := provider.Analyze(ctx, dbProvider.Model, req)
	if err != nil {
		s.logger.Error("Analysis failed",
			slog.String("provider_id", providerID),
			slog.String("model", dbProvider.Model),
			slog.String("error", err.Error()))
		return nil, err
	}

	s.logger.Info("AI analysis completed",
		slog.String("provider_id", providerID),
		slog.String("model", dbProvider.Model),
		slog.Int("tokens", response.TokensUsed),
		slog.Int64("duration_ms", response.DurationMs))

	return response, nil
}

// HasActiveProviders checks if any AI providers are active
func (s *Service) HasActiveProviders() bool {
	return len(s.providers) > 0
}
