package repositories

import (
	"database/sql"
	"errors"
	"log/slog"

	"ypeskov/kkal-tracker/internal/models"
)

type UserPreferenceRepositoryImpl struct {
	db        *sql.DB
	logger    *slog.Logger
	sqlLoader *SqlLoaderInstance
}

// NewUserPreferenceRepository creates a new user preference repository
func NewUserPreferenceRepository(db *sql.DB, dialect Dialect, logger *slog.Logger) *UserPreferenceRepositoryImpl {
	return &UserPreferenceRepositoryImpl{
		db:        db,
		logger:    logger.With("repository", "user_preference"),
		sqlLoader: NewSqlLoader(dialect),
	}
}

// Get returns the stored preference of the user for the key, ErrNotFound when there is none
func (r *UserPreferenceRepositoryImpl) Get(userID int, key string) (*models.UserPreference, error) {
	r.logger.Debug("Getting user preference", "user_id", userID, "key", key)

	query, err := r.sqlLoader.Load(QueryGetUserPreference)
	if err != nil {
		r.logger.Error("Failed to load query", "error", err)
		return nil, err
	}

	var preference models.UserPreference
	err = r.db.QueryRow(query, userID, key).Scan(
		&preference.UserID,
		&preference.Key,
		&preference.Value,
		&preference.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		r.logger.Error("Failed to get user preference", "error", err, "user_id", userID, "key", key)
		return nil, err
	}

	return &preference, nil
}

// Upsert stores the value of the key for the user, replacing the previous one
func (r *UserPreferenceRepositoryImpl) Upsert(userID int, key, value string) error {
	r.logger.Debug("Storing user preference", "user_id", userID, "key", key)

	query, err := r.sqlLoader.Load(QueryUpsertUserPreference)
	if err != nil {
		r.logger.Error("Failed to load query", "error", err)
		return err
	}

	if _, err := r.db.Exec(query, userID, key, value); err != nil {
		r.logger.Error("Failed to store user preference", "error", err, "user_id", userID, "key", key)
		return err
	}

	return nil
}
