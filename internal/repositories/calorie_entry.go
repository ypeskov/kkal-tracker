package repositories

import (
	"database/sql"
	"log/slog"
	"time"

	"ypeskov/kkal-tracker/internal/models"
)

type CalorieEntryRepositoryImpl struct {
	db        *sql.DB
	logger    *slog.Logger
	sqlLoader *SqlLoaderInstance
}

func NewCalorieEntryRepository(db *sql.DB, logger *slog.Logger, dialect Dialect) *CalorieEntryRepositoryImpl {
	sqlLoader := NewSqlLoader(dialect)

	return &CalorieEntryRepositoryImpl{
		db:        db,
		logger:    logger,
		sqlLoader: sqlLoader,
	}
}

func (r *CalorieEntryRepositoryImpl) Create(userID int,
	food string, calories int, weight float64, kcalPer100g float64, fats, carbs,
	proteins *float64, mealDatetime time.Time) (*models.CalorieEntry, error) {

	r.logger.Debug("Creating calorie entry",
		slog.Int("user_id", userID),
		slog.String("food", food),
		slog.Int("calories", calories))

	query, err := r.sqlLoader.Load(QueryInsertCalorieEntry)
	if err != nil {
		return nil, err
	}

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

// CreateMeal stores all entries of one meal in a single transaction: either every entry
// (and every new ingredient) is saved or none. The result is built from the inserted values
// instead of being read back, so a committed meal can never be reported as failed.
func (r *CalorieEntryRepositoryImpl) CreateMeal(userID int, entries []models.NewMealEntry,
	mealDatetime time.Time) ([]models.CreatedMealEntry, error) {

	r.logger.Debug("Creating meal",
		slog.Int("user_id", userID),
		slog.Int("entries", len(entries)))

	entryQuery, err := r.sqlLoader.Load(QueryInsertCalorieEntry)
	if err != nil {
		return nil, err
	}
	ingredientQuery, err := r.sqlLoader.Load(QueryInsertUserIngredient)
	if err != nil {
		return nil, err
	}

	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	now := time.Now().UTC()
	created := make([]models.CreatedMealEntry, 0, len(entries))

	for _, e := range entries {
		var ingredient *models.UserIngredient
		if e.SaveAsIngredient {
			result, err := tx.Exec(ingredientQuery, userID, e.Food, e.KcalPer100g, e.Fats, e.Carbs, e.Proteins)
			if err != nil {
				return nil, err
			}
			ingredientID, err := result.LastInsertId()
			if err != nil {
				return nil, err
			}
			ingredient = &models.UserIngredient{
				ID:          int(ingredientID),
				UserID:      userID,
				Name:        e.Food,
				KcalPer100g: e.KcalPer100g,
				Fats:        e.Fats,
				Carbs:       e.Carbs,
				Proteins:    e.Proteins,
				CreatedAt:   now,
				UpdatedAt:   now,
			}
		}

		result, err := tx.Exec(entryQuery, userID, e.Food, e.Calories, e.Weight, e.KcalPer100g,
			e.Fats, e.Carbs, e.Proteins, mealDatetime, now)
		if err != nil {
			return nil, err
		}
		entryID, err := result.LastInsertId()
		if err != nil {
			return nil, err
		}

		created = append(created, models.CreatedMealEntry{
			Entry: &models.CalorieEntry{
				ID:           int(entryID),
				UserID:       userID,
				Food:         e.Food,
				Calories:     e.Calories,
				Weight:       e.Weight,
				KcalPer100g:  e.KcalPer100g,
				Fats:         e.Fats,
				Carbs:        e.Carbs,
				Proteins:     e.Proteins,
				MealDatetime: mealDatetime,
				UpdatedAt:    now,
				CreatedAt:    now,
			},
			Ingredient: ingredient,
		})
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	r.logger.Debug("Meal created", slog.Int("user_id", userID), slog.Int("entries", len(created)))
	return created, nil
}

func (r *CalorieEntryRepositoryImpl) GetByID(id int) (*models.CalorieEntry, error) {
	r.logger.Debug("Getting calorie entry by ID", slog.Int("id", id))

	query, err := r.sqlLoader.Load(QueryGetCalorieEntryByID)
	if err != nil {
		return nil, err
	}

	entry := &models.CalorieEntry{}
	err = r.db.QueryRow(query, id).Scan(
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

func (r *CalorieEntryRepositoryImpl) GetByUserID(userID int) ([]*models.CalorieEntry, error) {
	r.logger.Debug("Getting calorie entries by user ID", slog.Int("user_id", userID))

	query, err := r.sqlLoader.Load(QueryGetCalorieEntriesByUserID)
	if err != nil {
		return nil, err
	}

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

func (r *CalorieEntryRepositoryImpl) GetByUserIDAndDateRange(userID int, dateFrom, dateTo string) ([]*models.CalorieEntry, error) {
	query, err := r.sqlLoader.Load(QueryGetCalorieEntriesByDateRange)
	if err != nil {
		return nil, err
	}

	// SQLite query needs the date range passed twice for the two conditions
	// PostgreSQL query only needs it once
	var rows *sql.Rows
	if r.sqlLoader.Dialect == DialectSQLite {
		rows, err = r.db.Query(query, userID, dateFrom, dateTo, dateFrom, dateTo)
	} else {
		rows, err = r.db.Query(query, userID, dateFrom, dateTo)
	}

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

func (r *CalorieEntryRepositoryImpl) Update(id, userID int, food string, calories int, weight float64, kcalPer100g float64, fats, carbs, proteins *float64, mealDatetime time.Time) (*models.CalorieEntry, error) {
	r.logger.Debug("Updating calorie entry",
		slog.Int("id", id),
		slog.Int("user_id", userID),
		slog.String("food", food))

	query, err := r.sqlLoader.Load(QueryUpdateCalorieEntry)
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	result, err := r.db.Exec(query, food, calories, weight, kcalPer100g, fats, carbs, proteins, mealDatetime, now, id, userID)
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

	return r.GetByID(id)
}

func (r *CalorieEntryRepositoryImpl) Delete(id, userID int) error {
	r.logger.Debug("Deleting calorie entry",
		slog.Int("id", id),
		slog.Int("user_id", userID))

	query, err := r.sqlLoader.Load(QueryDeleteCalorieEntry)
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
