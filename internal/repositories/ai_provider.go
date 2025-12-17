package repositories

import (
	"database/sql"
	"log/slog"

	"ypeskov/kkal-tracker/internal/models"
)

type AIProviderRepositoryImpl struct {
	db        *sql.DB
	logger    *slog.Logger
	sqlLoader *SqlLoaderInstance
}

func NewAIProviderRepository(db *sql.DB, logger *slog.Logger, dialect Dialect) *AIProviderRepositoryImpl {
	return &AIProviderRepositoryImpl{
		db:        db,
		logger:    logger.With("repository", "ai_provider"),
		sqlLoader: NewSqlLoader(dialect),
	}
}

func (r *AIProviderRepositoryImpl) GetAll() ([]*models.AIProvider, error) {
	r.logger.Debug("Getting all AI providers")

	query, err := r.sqlLoader.Load(QueryGetAllAIProviders)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []*models.AIProvider
	for rows.Next() {
		provider := &models.AIProvider{}
		err := rows.Scan(
			&provider.ID,
			&provider.ProviderType,
			&provider.DisplayName,
			&provider.Model,
			&provider.IsActive,
			&provider.CreatedAt,
			&provider.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		providers = append(providers, provider)
	}

	return providers, rows.Err()
}

func (r *AIProviderRepositoryImpl) GetActive() ([]*models.AIProvider, error) {
	r.logger.Debug("Getting active AI providers")

	query, err := r.sqlLoader.Load(QueryGetActiveAIProviders)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []*models.AIProvider
	for rows.Next() {
		provider := &models.AIProvider{}
		err := rows.Scan(
			&provider.ID,
			&provider.ProviderType,
			&provider.DisplayName,
			&provider.Model,
			&provider.IsActive,
			&provider.CreatedAt,
			&provider.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		providers = append(providers, provider)
	}

	return providers, rows.Err()
}

func (r *AIProviderRepositoryImpl) GetByID(id string) (*models.AIProvider, error) {
	r.logger.Debug("Getting AI provider by ID", slog.String("id", id))

	query, err := r.sqlLoader.Load(QueryGetAIProviderByID)
	if err != nil {
		return nil, err
	}

	provider := &models.AIProvider{}
	err = r.db.QueryRow(query, id).Scan(
		&provider.ID,
		&provider.ProviderType,
		&provider.DisplayName,
		&provider.Model,
		&provider.IsActive,
		&provider.CreatedAt,
		&provider.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return provider, nil
}

func (r *AIProviderRepositoryImpl) UpdateModel(id string, model string) error {
	r.logger.Debug("Updating AI provider model", slog.String("id", id), slog.String("model", model))

	query, err := r.sqlLoader.Load(QueryUpdateAIProviderModel)
	if err != nil {
		return err
	}

	result, err := r.db.Exec(query, model, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *AIProviderRepositoryImpl) SetActive(id string, isActive bool) error {
	r.logger.Debug("Setting AI provider active status", slog.String("id", id), slog.Bool("is_active", isActive))

	query, err := r.sqlLoader.Load(QuerySetAIProviderActive)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(query, isActive, id)
	return err
}

func (r *AIProviderRepositoryImpl) SetActiveByType(providerType string, isActive bool) error {
	r.logger.Debug("Setting AI providers active by type", slog.String("provider_type", providerType), slog.Bool("is_active", isActive))

	query, err := r.sqlLoader.Load(QuerySetAIProviderActiveByType)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(query, isActive, providerType)
	return err
}
