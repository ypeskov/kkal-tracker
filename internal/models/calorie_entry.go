package models

import (
	"database/sql"
	"time"
)

type CalorieEntry struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Food      string    `json:"food"`
	Calories  int       `json:"calories"`
	Date      string    `json:"date"`
	CreatedAt time.Time `json:"created_at"`
}

type CalorieEntryRepository struct {
	db *sql.DB
}

func NewCalorieEntryRepository(db *sql.DB) *CalorieEntryRepository {
	return &CalorieEntryRepository{db: db}
}

func (r *CalorieEntryRepository) Create(userID int, food string, calories int, date string) (*CalorieEntry, error) {
	query := `
		INSERT INTO calorie_entries (user_id, food, calories, date)
		VALUES (?, ?, ?, ?)
	`
	
	result, err := r.db.Exec(query, userID, food, calories, date)
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
		SELECT id, user_id, food, calories, date, created_at
		FROM calorie_entries
		WHERE id = ?
	`
	
	entry := &CalorieEntry{}
	err := r.db.QueryRow(query, id).Scan(
		&entry.ID,
		&entry.UserID,
		&entry.Food,
		&entry.Calories,
		&entry.Date,
		&entry.CreatedAt,
	)
	
	if err != nil {
		return nil, err
	}
	
	return entry, nil
}

func (r *CalorieEntryRepository) GetByUserID(userID int) ([]*CalorieEntry, error) {
	query := `
		SELECT id, user_id, food, calories, date, created_at
		FROM calorie_entries
		WHERE user_id = ?
		ORDER BY date DESC, created_at DESC
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
			&entry.Date,
			&entry.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (r *CalorieEntryRepository) GetByUserIDAndDate(userID int, date string) ([]*CalorieEntry, error) {
	query := `
		SELECT id, user_id, food, calories, date, created_at
		FROM calorie_entries
		WHERE user_id = ? AND date = ?
		ORDER BY created_at DESC
	`
	
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
			&entry.Date,
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