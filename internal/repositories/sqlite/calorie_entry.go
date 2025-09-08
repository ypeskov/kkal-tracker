package sqlite

import (
	"database/sql"
	"log/slog"
	"time"

	"ypeskov/kkal-tracker/internal/models"
)

type CalorieEntryRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewCalorieEntryRepository(db *sql.DB, logger *slog.Logger) *CalorieEntryRepository {
	return &CalorieEntryRepository{
		db:     db,
		logger: logger,
	}
}

func (r *CalorieEntryRepository) Create(userID int,
	food string, calories int, weight float64, kcalPer100g float64, fats, carbs,
	proteins *float64, mealDatetime time.Time) (*models.CalorieEntry, error) {

	r.logger.Debug("Creating calorie entry", 
		slog.Int("user_id", userID),
		slog.String("food", food),
		slog.Int("calories", calories))

	query := `
		INSERT INTO calorie_entries (user_id, food, calories, weight, kcal_per_100g, fats, carbs, proteins, meal_datetime, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now().UTC()
	result, err := r.db.Exec(query, userID, food, calories, weight, kcalPer100g, fats, carbs, proteins, mealDatetime, now)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	r.logger.Debug("Calorie entry created", slog.Int64("id", id))
	return r.GetByID(int(id))
}

func (r *CalorieEntryRepository) GetByID(id int) (*models.CalorieEntry, error) {
	r.logger.Debug("Getting calorie entry by ID", slog.Int("id", id))

	query := `
		SELECT id, user_id, food, calories, weight, kcal_per_100g, fats, carbs, proteins, meal_datetime, updated_at, created_at
		FROM calorie_entries
		WHERE id = ?
	`

	entry := &models.CalorieEntry{}
	err := r.db.QueryRow(query, id).Scan(
		&entry.ID,
		&entry.UserID,
		&entry.Food,
		&entry.Calories,
		&entry.Weight,
		&entry.KcalPer100g,
		&entry.Fats,
		&entry.Carbs,
		&entry.Proteins,
		&entry.MealDatetime,
		&entry.UpdatedAt,
		&entry.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return entry, nil
}

func (r *CalorieEntryRepository) GetByUserID(userID int) ([]*models.CalorieEntry, error) {
	r.logger.Debug("Getting calorie entries by user ID", slog.Int("user_id", userID))

	query := `
		SELECT id, user_id, food, calories, weight, kcal_per_100g, fats, carbs, proteins, meal_datetime, updated_at, created_at
		FROM calorie_entries
		WHERE user_id = ?
		ORDER BY meal_datetime DESC, created_at DESC
	`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*models.CalorieEntry
	for rows.Next() {
		entry := &models.CalorieEntry{}
		err := rows.Scan(
			&entry.ID,
			&entry.UserID,
			&entry.Food,
			&entry.Calories,
			&entry.Weight,
			&entry.KcalPer100g,
			&entry.Fats,
			&entry.Carbs,
			&entry.Proteins,
			&entry.MealDatetime,
			&entry.UpdatedAt,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (r *CalorieEntryRepository) GetByUserIDAndDate(userID int, date string) ([]*models.CalorieEntry, error) {
	// SQLite-specific date formatting using strftime
	query := `
		SELECT id, user_id, food, calories, weight, kcal_per_100g, fats, carbs, proteins, meal_datetime, updated_at, created_at
		FROM calorie_entries
		WHERE user_id = ? AND (
			strftime('%Y-%m-%d', meal_datetime) = ?
			OR strftime('%Y-%m-%d', substr(meal_datetime, 1, 19)) = ?
		)
		ORDER BY meal_datetime DESC
	`

	r.logger.Debug("GetByUserIDAndDate", slog.Int("userID", userID), slog.String("date", date))

	rows, err := r.db.Query(query, userID, date, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*models.CalorieEntry
	for rows.Next() {
		entry := &models.CalorieEntry{}
		err := rows.Scan(
			&entry.ID,
			&entry.UserID,
			&entry.Food,
			&entry.Calories,
			&entry.Weight,
			&entry.KcalPer100g,
			&entry.Fats,
			&entry.Carbs,
			&entry.Proteins,
			&entry.MealDatetime,
			&entry.UpdatedAt,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (r *CalorieEntryRepository) GetByUserIDAndDateRange(userID int, dateFrom, dateTo string) ([]*models.CalorieEntry, error) {
	// SQLite-specific date formatting using strftime
	query := `
		SELECT id, user_id, food, calories, weight, kcal_per_100g, fats, carbs, proteins, meal_datetime, updated_at, created_at
		FROM calorie_entries
		WHERE user_id = ? AND (
			strftime('%Y-%m-%d', meal_datetime) BETWEEN ? AND ?
			OR strftime('%Y-%m-%d', substr(meal_datetime, 1, 19)) BETWEEN ? AND ?
		)
		ORDER BY meal_datetime DESC
	`

	rows, err := r.db.Query(query, userID, dateFrom, dateTo, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*models.CalorieEntry
	for rows.Next() {
		entry := &models.CalorieEntry{}
		err := rows.Scan(
			&entry.ID,
			&entry.UserID,
			&entry.Food,
			&entry.Calories,
			&entry.Weight,
			&entry.KcalPer100g,
			&entry.Fats,
			&entry.Carbs,
			&entry.Proteins,
			&entry.MealDatetime,
			&entry.UpdatedAt,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (r *CalorieEntryRepository) Update(id, userID int, food string, calories int, weight float64, kcalPer100g float64, fats, carbs, proteins *float64, mealDatetime time.Time) (*models.CalorieEntry, error) {
	r.logger.Debug("Updating calorie entry", 
		slog.Int("id", id),
		slog.Int("user_id", userID),
		slog.String("food", food))

	query := `
		UPDATE calorie_entries 
		SET food = ?, calories = ?, weight = ?, kcal_per_100g = ?, fats = ?, carbs = ?, proteins = ?, meal_datetime = ?, updated_at = ?
		WHERE id = ? AND user_id = ?
	`

	now := time.Now().UTC()
	_, err := r.db.Exec(query, food, calories, weight, kcalPer100g, fats, carbs, proteins, mealDatetime, now, id, userID)
	if err != nil {
		return nil, err
	}

	return r.GetByID(id)
}

func (r *CalorieEntryRepository) Delete(id, userID int) error {
	r.logger.Debug("Deleting calorie entry", 
		slog.Int("id", id),
		slog.Int("user_id", userID))

	query := `
		DELETE FROM calorie_entries
		WHERE id = ? AND user_id = ?
	`

	_, err := r.db.Exec(query, id, userID)
	return err
}
