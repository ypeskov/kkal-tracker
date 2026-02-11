package repositories

import (
	"database/sql"
	"log/slog"
	"time"

	"ypeskov/kkal-tracker/internal/models"
)

type APIKeyRepositoryImpl struct {
	db        *sql.DB
	logger    *slog.Logger
	sqlLoader *SqlLoaderInstance
}

// NewAPIKeyRepository creates a new API key repository
func NewAPIKeyRepository(db *sql.DB, dialect Dialect, logger *slog.Logger) *APIKeyRepositoryImpl {
	return &APIKeyRepositoryImpl{
		db:        db,
		logger:    logger.With("repository", "api_key"),
		sqlLoader: NewSqlLoader(dialect),
	}
}

// Create stores a new API key record
func (r *APIKeyRepositoryImpl) Create(userID int, name, keyHash, keyPrefix string, expiresAt *time.Time) (*models.APIKey, error) {
	r.logger.Debug("Creating API key", "user_id", userID, "name", name, "prefix", keyPrefix)

	query, err := r.sqlLoader.Load(QueryCreateAPIKey)
	if err != nil {
		r.logger.Error("Failed to load query", "error", err)
		return nil, err
	}

	result, err := r.db.Exec(query, userID, name, keyHash, keyPrefix, expiresAt)
	if err != nil {
		r.logger.Error("Failed to insert API key", "error", err)
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		r.logger.Error("Failed to get last insert ID", "error", err)
		return nil, err
	}

	apiKey := &models.APIKey{
		ID:        int(id),
		UserID:    userID,
		Name:      name,
		KeyHash:   keyHash,
		KeyPrefix: keyPrefix,
		ExpiresAt: expiresAt,
		IsRevoked: false,
		CreatedAt: time.Now(),
	}

	r.logger.Debug("API key created", "id", id, "user_id", userID, "prefix", keyPrefix)
	return apiKey, nil
}

// GetByKeyHash retrieves an API key by its hash
func (r *APIKeyRepositoryImpl) GetByKeyHash(keyHash string) (*models.APIKey, error) {
	r.logger.Debug("Getting API key by hash")

	query, err := r.sqlLoader.Load(QueryGetAPIKeyByHash)
	if err != nil {
		r.logger.Error("Failed to load query", "error", err)
		return nil, err
	}

	apiKey, err := r.scanAPIKey(r.db.QueryRow(query, keyHash))
	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Debug("API key not found by hash")
			return nil, ErrNotFound
		}
		r.logger.Error("Failed to get API key by hash", "error", err)
		return nil, err
	}

	return apiKey, nil
}

// GetByUserID retrieves all API keys for a user
func (r *APIKeyRepositoryImpl) GetByUserID(userID int) ([]*models.APIKey, error) {
	r.logger.Debug("Getting API keys for user", "user_id", userID)

	query, err := r.sqlLoader.Load(QueryGetAPIKeysByUserID)
	if err != nil {
		r.logger.Error("Failed to load query", "error", err)
		return nil, err
	}

	rows, err := r.db.Query(query, userID)
	if err != nil {
		r.logger.Error("Failed to query API keys", "error", err)
		return nil, err
	}
	defer rows.Close()

	var keys []*models.APIKey
	for rows.Next() {
		apiKey, err := r.scanAPIKey(rows)
		if err != nil {
			r.logger.Error("Failed to scan API key row", "error", err)
			return nil, err
		}
		keys = append(keys, apiKey)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating API key rows", "error", err)
		return nil, err
	}

	if keys == nil {
		keys = []*models.APIKey{}
	}

	r.logger.Debug("API keys retrieved", "user_id", userID, "count", len(keys))
	return keys, nil
}

// Revoke marks an API key as revoked
func (r *APIKeyRepositoryImpl) Revoke(id, userID int) error {
	r.logger.Debug("Revoking API key", "id", id, "user_id", userID)

	query, err := r.sqlLoader.Load(QueryRevokeAPIKey)
	if err != nil {
		r.logger.Error("Failed to load query", "error", err)
		return err
	}

	result, err := r.db.Exec(query, id, userID)
	if err != nil {
		r.logger.Error("Failed to revoke API key", "error", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", "error", err)
		return err
	}

	if rowsAffected == 0 {
		r.logger.Debug("No API key revoked (not found)", "id", id, "user_id", userID)
		return ErrNotFound
	}

	r.logger.Debug("API key revoked", "id", id, "user_id", userID)
	return nil
}

// Delete removes an API key
func (r *APIKeyRepositoryImpl) Delete(id, userID int) error {
	r.logger.Debug("Deleting API key", "id", id, "user_id", userID)

	query, err := r.sqlLoader.Load(QueryDeleteAPIKey)
	if err != nil {
		r.logger.Error("Failed to load query", "error", err)
		return err
	}

	result, err := r.db.Exec(query, id, userID)
	if err != nil {
		r.logger.Error("Failed to delete API key", "error", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", "error", err)
		return err
	}

	if rowsAffected == 0 {
		r.logger.Debug("No API key deleted (not found)", "id", id, "user_id", userID)
		return ErrNotFound
	}

	r.logger.Debug("API key deleted", "id", id, "user_id", userID)
	return nil
}

// scanner is an interface satisfied by both *sql.Row and *sql.Rows
type scanner interface {
	Scan(dest ...any) error
}

// scanAPIKey scans a single API key from a row
func (r *APIKeyRepositoryImpl) scanAPIKey(row scanner) (*models.APIKey, error) {
	var apiKey models.APIKey
	var expiresAt sql.NullTime
	var isRevoked int

	err := row.Scan(
		&apiKey.ID,
		&apiKey.UserID,
		&apiKey.Name,
		&apiKey.KeyHash,
		&apiKey.KeyPrefix,
		&expiresAt,
		&isRevoked,
		&apiKey.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	if expiresAt.Valid {
		apiKey.ExpiresAt = &expiresAt.Time
	}
	apiKey.IsRevoked = isRevoked != 0

	return &apiKey, nil
}
