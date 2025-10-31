package repositories

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"log/slog"
	"time"

	"ypeskov/kkal-tracker/internal/models"
)

// ActivationTokenRepository defines methods for activation token data access
type ActivationTokenRepository interface {
	Create(userID int, expiresAt time.Time) (*models.ActivationToken, error)
	GetByToken(token string) (*models.ActivationToken, error)
	Delete(token string) error
	DeleteExpired() (int64, error)
}

type ActivationTokenRepositoryImpl struct {
	db        *sql.DB
	logger    *slog.Logger
	sqlLoader *SqlLoaderInstance
}

// NewActivationTokenRepository creates a new activation token repository
func NewActivationTokenRepository(db *sql.DB, dialect Dialect, logger *slog.Logger) *ActivationTokenRepositoryImpl {
	return &ActivationTokenRepositoryImpl{
		db:        db,
		logger:    logger.With("repository", "activation_token"),
		sqlLoader: NewSqlLoader(dialect),
	}
}

// generateToken creates a cryptographically secure random token
func generateToken() (string, error) {
	bytes := make([]byte, 32) // 32 bytes = 64 hex characters
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Create generates and stores a new activation token
func (r *ActivationTokenRepositoryImpl) Create(userID int, expiresAt time.Time) (*models.ActivationToken, error) {
	r.logger.Debug("Creating activation token", "user_id", userID, "expires_at", expiresAt)

	token, err := generateToken()
	if err != nil {
		r.logger.Error("Failed to generate token", "error", err)
		return nil, err
	}

	query, err := r.sqlLoader.Load(QueryCreateActivationToken)
	if err != nil {
		r.logger.Error("Failed to load query", "error", err)
		return nil, err
	}

	result, err := r.db.Exec(query, userID, token, expiresAt)
	if err != nil {
		r.logger.Error("Failed to insert activation token", "error", err)
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		r.logger.Error("Failed to get last insert ID", "error", err)
		return nil, err
	}

	activationToken := &models.ActivationToken{
		ID:        int(id),
		UserID:    userID,
		Token:     token,
		CreatedAt: time.Now(),
		ExpiresAt: expiresAt,
	}

	r.logger.Debug("Activation token created", "id", id, "user_id", userID)
	return activationToken, nil
}

// GetByToken retrieves an activation token by its token string
func (r *ActivationTokenRepositoryImpl) GetByToken(token string) (*models.ActivationToken, error) {
	r.logger.Debug("Getting activation token", "token", token[:8]+"...")

	query, err := r.sqlLoader.Load(QueryGetActivationTokenByToken)
	if err != nil {
		r.logger.Error("Failed to load query", "error", err)
		return nil, err
	}

	var activationToken models.ActivationToken
	err = r.db.QueryRow(query, token).Scan(
		&activationToken.ID,
		&activationToken.UserID,
		&activationToken.Token,
		&activationToken.CreatedAt,
		&activationToken.ExpiresAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			r.logger.Debug("Activation token not found", "token", token[:8]+"...")
			return nil, ErrNotFound
		}
		r.logger.Error("Failed to get activation token", "error", err)
		return nil, err
	}

	r.logger.Debug("Activation token found", "id", activationToken.ID, "user_id", activationToken.UserID)
	return &activationToken, nil
}

// Delete removes an activation token
func (r *ActivationTokenRepositoryImpl) Delete(token string) error {
	r.logger.Debug("Deleting activation token", "token", token[:8]+"...")

	query, err := r.sqlLoader.Load(QueryDeleteActivationToken)
	if err != nil {
		r.logger.Error("Failed to load query", "error", err)
		return err
	}

	result, err := r.db.Exec(query, token)
	if err != nil {
		r.logger.Error("Failed to delete activation token", "error", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", "error", err)
		return err
	}

	if rowsAffected == 0 {
		r.logger.Debug("No activation token deleted (not found)", "token", token[:8]+"...")
		return ErrNotFound
	}

	r.logger.Debug("Activation token deleted", "rows_affected", rowsAffected)
	return nil
}

// DeleteExpired removes all expired activation tokens
func (r *ActivationTokenRepositoryImpl) DeleteExpired() (int64, error) {
	r.logger.Debug("Deleting expired activation tokens")

	query, err := r.sqlLoader.Load(QueryDeleteExpiredActivationTokens)
	if err != nil {
		r.logger.Error("Failed to load query", "error", err)
		return 0, err
	}

	result, err := r.db.Exec(query, time.Now())
	if err != nil {
		r.logger.Error("Failed to delete expired tokens", "error", err)
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", "error", err)
		return 0, err
	}

	r.logger.Info("Expired activation tokens deleted", "count", rowsAffected)
	return rowsAffected, nil
}
