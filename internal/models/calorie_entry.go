package models

import (
	"database/sql"
	"log/slog"
	"time"
)

type CalorieEntry struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	Food         string    `json:"food"`
	Calories     int       `json:"calories"`
	Weight       float64   `json:"weight"`
	KcalPer100g  float64   `json:"kcalPer100g"`
	Fats         *float64  `json:"fats,omitempty"`
	Carbs        *float64  `json:"carbs,omitempty"`
	Proteins     *float64  `json:"proteins,omitempty"`
	MealDatetime time.Time `json:"meal_datetime"`
	UpdatedAt    time.Time `json:"updated_at"`
	CreatedAt    time.Time `json:"created_at"`
}

type CalorieEntryRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewCalorieEntryRepository(db *sql.DB, log *slog.Logger) *CalorieEntryRepository {
	return &CalorieEntryRepository{
		db:     db,
		logger: log,
	}
}

func (r *CalorieEntryRepository) Create(userID int,
	food string, calories int, weight float64, kcalPer100g float64, fats, carbs,
	proteins *float64, mealDatetime string) (*CalorieEntry, error) {

	query := `
		INSERT INTO calorie_entries (user_id, food, calories, weight, kcal_per_100g, fats, carbs, proteins, meal_datetime, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	result, err := r.db.Exec(query, userID, food, calories, weight, kcalPer100g, fats, carbs, proteins, mealDatetime, now)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return r.GetByID(int(id))
}

func (r *CalorieEntryRepository) GetByID(id int) (*CalorieEntry, error) {
	query := `
		SELECT id, user_id, food, calories, weight, kcal_per_100g, fats, carbs, proteins, meal_datetime, updated_at, created_at
		FROM calorie_entries
		WHERE id = ?
	`

	entry := &CalorieEntry{}
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

func (r *CalorieEntryRepository) GetByUserID(userID int) ([]*CalorieEntry, error) {
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

	var entries []*CalorieEntry
	for rows.Next() {
		entry := &CalorieEntry{}
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

func (r *CalorieEntryRepository) Update(id, userID int, food string, calories int, weight float64, kcalPer100g float64, fats, carbs, proteins *float64, mealDatetime time.Time) (*CalorieEntry, error) {
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

func (r *CalorieEntryRepository) GetByUserIDAndDate(userID int, date string) ([]*CalorieEntry, error) {
	query := `
		SELECT id, user_id, food, calories, weight, kcal_per_100g, fats, carbs, proteins, meal_datetime, updated_at, created_at
		FROM calorie_entries
		WHERE user_id = ? AND strftime('%Y-%m-%d', meal_datetime) = ?
		ORDER BY meal_datetime DESC
	`
	r.logger.Debug("GetByUserIDAndDate", slog.Int("userID", userID), slog.String("date", date))

	rows, err := r.db.Query(query, userID, date)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*CalorieEntry
	for rows.Next() {
		entry := &CalorieEntry{}
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

func (r *CalorieEntryRepository) Delete(id, userID int) error {
	query := `
		DELETE FROM calorie_entries
		WHERE id = ? AND user_id = ?
	`

	_, err := r.db.Exec(query, id, userID)
	return err
}

func (r *CalorieEntryRepository) GetByUserIDAndDateRange(userID int, dateFrom, dateTo string) ([]*CalorieEntry, error) {
	query := `
		SELECT id, user_id, food, calories, weight, kcal_per_100g, fats, carbs, proteins, meal_datetime, updated_at, created_at
		FROM calorie_entries
		WHERE user_id = ? AND strftime('%Y-%m-%d', meal_datetime) BETWEEN ? AND ?
		ORDER BY meal_datetime DESC
	`

	rows, err := r.db.Query(query, userID, dateFrom, dateTo)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []*CalorieEntry
	for rows.Next() {
		entry := &CalorieEntry{}
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
