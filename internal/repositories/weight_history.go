package repositories

import (
	"database/sql"
	"log/slog"
	"time"

	"ypeskov/kkal-tracker/internal/models"
)

type WeightHistoryRepositoryImpl struct {
	db        *sql.DB
	logger    *slog.Logger
	sqlLoader *SqlLoaderInstance
}

func NewWeightHistoryRepository(db *sql.DB, logger *slog.Logger, dialect Dialect) *WeightHistoryRepositoryImpl {
	sqlLoader := NewSqlLoader(dialect)

	return &WeightHistoryRepositoryImpl{
		db:        db,
		logger:    logger,
		sqlLoader: sqlLoader,
	}
}

func (r *WeightHistoryRepositoryImpl) GetByUserID(userID int) ([]*models.WeightHistory, error) {
	r.logger.Debug("Getting weight history", slog.Int("user_id", userID))

	query, err := r.sqlLoader.Load(QueryGetWeightHistory)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*models.WeightHistory
	for rows.Next() {
		var entry models.WeightHistory
		if err := rows.Scan(&entry.ID, &entry.UserID, &entry.Weight, &entry.RecordedAt, &entry.CreatedAt); err != nil {
			return nil, err
		}
		history = append(history, &entry)
	}

	return history, rows.Err()
}

func (r *WeightHistoryRepositoryImpl) GetByUserIDAndDateRange(userID int, dateFrom, dateTo string) ([]*models.WeightHistory, error) {
	r.logger.Debug("Getting weight history by date range",
		slog.Int("user_id", userID),
		slog.String("date_from", dateFrom),
		slog.String("date_to", dateTo))

	query, err := r.sqlLoader.Load(QueryGetWeightHistoryByDateRange)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(query, userID, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []*models.WeightHistory
	for rows.Next() {
		var entry models.WeightHistory
		if err := rows.Scan(&entry.ID, &entry.UserID, &entry.Weight, &entry.RecordedAt, &entry.CreatedAt); err != nil {
			return nil, err
		}
		history = append(history, &entry)
	}

	return history, rows.Err()
}

func (r *WeightHistoryRepositoryImpl) GetLatestByUserID(userID int) (*models.WeightHistory, error) {
	r.logger.Debug("Getting latest weight history", slog.Int("user_id", userID))

	query, err := r.sqlLoader.Load(QueryGetLatestWeightHistory)
	if err != nil {
		return nil, err
	}

	var entry models.WeightHistory
	err = r.db.QueryRow(query, userID).Scan(&entry.ID, &entry.UserID, &entry.Weight, &entry.RecordedAt, &entry.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No weight history yet
		}
		return nil, err
	}

	return &entry, nil
}

func (r *WeightHistoryRepositoryImpl) Create(userID int, weight float64, recordedAt *time.Time) (*models.WeightHistory, error) {
	r.logger.Debug("Creating weight history entry",
		slog.Int("user_id", userID),
		slog.Float64("weight", weight))

	query, err := r.sqlLoader.Load(QueryCreateWeightHistory)
	if err != nil {
		return nil, err
	}

	var result sql.Result
	if recordedAt != nil {
		result, err = r.db.Exec(query, userID, weight, recordedAt)
	} else {
		result, err = r.db.Exec(query, userID, weight, time.Now())
	}

	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &models.WeightHistory{
		ID:         int(id),
		UserID:     userID,
		Weight:     weight,
		RecordedAt: time.Now(),
		CreatedAt:  time.Now(),
	}, nil
}

func (r *WeightHistoryRepositoryImpl) Update(id, userID int, weight float64, recordedAt *time.Time) (*models.WeightHistory, error) {
	r.logger.Debug("Updating weight history entry",
		slog.Int("id", id),
		slog.Int("user_id", userID),
		slog.Float64("weight", weight))

	query, err := r.sqlLoader.Load(QueryUpdateWeightHistory)
	if err != nil {
		return nil, err
	}

	var result sql.Result
	if recordedAt != nil {
		result, err = r.db.Exec(query, weight, recordedAt, id, userID)
	} else {
		result, err = r.db.Exec(query, weight, time.Now(), id, userID)
	}

	if err != nil {
		return nil, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rowsAffected == 0 {
		return nil, ErrNotFound
	}

	return &models.WeightHistory{
		ID:         id,
		UserID:     userID,
		Weight:     weight,
		RecordedAt: *recordedAt,
	}, nil
}

func (r *WeightHistoryRepositoryImpl) Delete(id, userID int) error {
	r.logger.Debug("Deleting weight history entry",
		slog.Int("id", id),
		slog.Int("user_id", userID))

	query, err := r.sqlLoader.Load(QueryDeleteWeightHistory)
	if err != nil {
		return err
	}

	result, err := r.db.Exec(query, id, userID)
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