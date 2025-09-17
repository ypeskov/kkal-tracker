package repositories

import (
	"database/sql"
	"log/slog"
	"time"

	"ypeskov/kkal-tracker/internal/models"
)

type IngredientRepositoryImpl struct {
	db        *sql.DB
	logger    *slog.Logger
	sqlLoader *SqlLoaderInstance
}

func NewIngredientRepository(db *sql.DB, logger *slog.Logger, dialect Dialect) *IngredientRepositoryImpl {
	sqlLoader := NewSqlLoader(dialect)

	return &IngredientRepositoryImpl{
		db:        db,
		logger:    logger.With(slog.String("repo", "IngredientRepository")),
		sqlLoader: sqlLoader,
	}
}

// Search user ingredients for autocomplete
func (r *IngredientRepositoryImpl) SearchUserIngredients(userID int, query string, limit int) ([]*models.UserIngredient, error) {
	r.logger.Debug("searching user ingredients", slog.Int("user_id", userID), slog.String("query", query), slog.Int("limit", limit))

	sqlQuery, err := r.sqlLoader.Load(QuerySearchUserIngredients)
	if err != nil {
		r.logger.Error("failed to load SQL query", "error", err)
		return nil, err
	}

	rows, err := r.db.Query(sqlQuery, userID, "%"+query+"%", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ingredients []*models.UserIngredient
	for rows.Next() {
		ingredient := &models.UserIngredient{}
		err := rows.Scan(
			&ingredient.ID,
			&ingredient.UserID,
			&ingredient.Name,
			&ingredient.KcalPer100g,
			&ingredient.Fats,
			&ingredient.Carbs,
			&ingredient.Proteins,
			&ingredient.GlobalIngredientID,
			&ingredient.CreatedAt,
			&ingredient.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		ingredients = append(ingredients, ingredient)
	}

	return ingredients, nil
}

// Get all user ingredients for session storage
func (r *IngredientRepositoryImpl) GetAllUserIngredients(userID int) ([]*models.UserIngredient, error) {
	r.logger.Debug("getting all user ingredients", slog.Int("user_id", userID))

	sqlQuery, err := r.sqlLoader.Load(QueryGetAllUserIngredients)
	if err != nil {
		r.logger.Error("failed to load SQL query QueryGetAllUserIngredients", "error", err)
		return nil, err
	}

	rows, err := r.db.Query(sqlQuery, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ingredients []*models.UserIngredient
	for rows.Next() {
		ingredient := &models.UserIngredient{}
		err := rows.Scan(
			&ingredient.ID,
			&ingredient.UserID,
			&ingredient.Name,
			&ingredient.KcalPer100g,
			&ingredient.Fats,
			&ingredient.Carbs,
			&ingredient.Proteins,
			&ingredient.GlobalIngredientID,
			&ingredient.CreatedAt,
			&ingredient.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		ingredients = append(ingredients, ingredient)
	}

	return ingredients, nil
}

// GetUserIngredientByName Get user ingredient by name
func (r *IngredientRepositoryImpl) GetUserIngredientByName(userID int, name string) (*models.UserIngredient, error) {
	query, err := r.sqlLoader.Load(QueryGetUserIngredientByName)
	if err != nil {
		return nil, err
	}

	ingredient := &models.UserIngredient{}
	err = r.db.QueryRow(query, userID, name).Scan(
		&ingredient.ID,
		&ingredient.UserID,
		&ingredient.Name,
		&ingredient.KcalPer100g,
		&ingredient.Fats,
		&ingredient.Carbs,
		&ingredient.Proteins,
		&ingredient.GlobalIngredientID,
		&ingredient.CreatedAt,
		&ingredient.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return ingredient, nil
}

// CreateOrUpdateUserIngredient Create or update user ingredient (check first, then insert/update)
func (r *IngredientRepositoryImpl) CreateOrUpdateUserIngredient(userID int, name string, kcalPer100g float64, fats, carbs, proteins *float64) (*models.UserIngredient, error) {
	r.logger.Debug("Creating or updating user ingredient",
		slog.Int("user_id", userID),
		slog.String("name", name),
		slog.Float64("kcal_per_100g", kcalPer100g))

	// Check if ingredient already exists
	_, err := r.GetUserIngredientByName(userID, name)
	if err == nil {
		// Ingredient exists, update it
		// Note: This uses a special simplified update query
		updateQuery := `
			UPDATE user_ingredients
			SET kcal_per_100g = ?, fats = ?, carbs = ?, proteins = ?, updated_at = CURRENT_TIMESTAMP
			WHERE user_id = ? AND name = ?
		`
		_, updateErr := r.db.Exec(updateQuery, kcalPer100g, fats, carbs, proteins, userID, name)
		if updateErr != nil {
			return nil, updateErr
		}
		return r.GetUserIngredientByName(userID, name)
	}

	// Ingredient doesn't exist, create it
	insertQuery, err := r.sqlLoader.Load(QueryInsertUserIngredient)
	if err != nil {
		return nil, err
	}
	_, err = r.db.Exec(insertQuery, userID, name, kcalPer100g, fats, carbs, proteins)
	if err != nil {
		return nil, err
	}

	// Get the newly created ingredient
	return r.GetUserIngredientByName(userID, name)
}

// Copy global ingredients to user ingredients table in specified language
func (r *IngredientRepositoryImpl) CopyGlobalIngredientsToUser(userID int, languageCode string) error {
	r.logger.Debug("Copying global ingredients to user",
		slog.Int("user_id", userID),
		slog.String("language", languageCode))

	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// First check if user already has ingredients
	var count int
	countQuery, err := r.sqlLoader.Load(QueryCountUserIngredients)
	if err != nil {
		return err
	}
	err = tx.QueryRow(countQuery, userID).Scan(&count)
	if err != nil {
		return err
	}

	if count > 0 {
		r.logger.Debug("User already has ingredients, skipping copy",
			slog.Int("user_id", userID),
			slog.Int("existing_count", count))
		return tx.Commit()
	}

	// Insert only if user has no ingredients yet (prevents duplicates)
	query, err := r.sqlLoader.Load(QueryCopyGlobalIngredientsToUser)
	if err != nil {
		return err
	}

	// For SQLite, we need to pass userID twice (for the main query and the NOT EXISTS)
	if r.sqlLoader.Dialect == DialectSQLite {
		_, err = tx.Exec(query, userID, languageCode, userID)
	} else {
		// PostgreSQL uses numbered placeholders, userID is $1 and referenced twice
		_, err = tx.Exec(query, userID, languageCode)
	}
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Admin functions for global ingredients
func (r *IngredientRepositoryImpl) CreateGlobalIngredient(kcalPer100g float64, fats, carbs, proteins *float64, names map[string]string) (*models.GlobalIngredient, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Insert global ingredient
	now := time.Now().UTC()
	insertQuery, err := r.sqlLoader.Load(QueryInsertGlobalIngredient)
	if err != nil {
		return nil, err
	}
	result, err := tx.Exec(insertQuery, kcalPer100g, fats, carbs, proteins, now, now)
	if err != nil {
		return nil, err
	}

	ingredientID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Insert names
	nameQuery, err := r.sqlLoader.Load(QueryInsertGlobalIngredientName)
	if err != nil {
		return nil, err
	}
	for langCode, name := range names {
		_, err = tx.Exec(nameQuery, ingredientID, langCode, name)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetGlobalIngredientByID(int(ingredientID))
}

func (r *IngredientRepositoryImpl) GetGlobalIngredientByID(id int) (*models.GlobalIngredient, error) {
	query, err := r.sqlLoader.Load(QueryGetGlobalIngredientByID)
	if err != nil {
		return nil, err
	}

	ingredient := &models.GlobalIngredient{}
	err = r.db.QueryRow(query, id).Scan(
		&ingredient.ID,
		&ingredient.KcalPer100g,
		&ingredient.Fats,
		&ingredient.Carbs,
		&ingredient.Proteins,
		&ingredient.CreatedAt,
		&ingredient.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Load names for this ingredient
	names, err := r.getGlobalIngredientNames(ingredient.ID)
	if err != nil {
		return nil, err
	}
	ingredient.Names = names

	return ingredient, nil
}

func (r *IngredientRepositoryImpl) getGlobalIngredientNames(ingredientID int) (map[string]string, error) {
	query, err := r.sqlLoader.Load(QueryGetGlobalIngredientNames)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(query, ingredientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	names := make(map[string]string)
	for rows.Next() {
		var langCode, name string
		err := rows.Scan(&langCode, &name)
		if err != nil {
			return nil, err
		}
		names[langCode] = name
	}

	return names, nil
}

// GetUserIngredientByID Get user ingredient by ID
func (r *IngredientRepositoryImpl) GetUserIngredientByID(userID int, ingredientID int) (*models.UserIngredient, error) {
	query, err := r.sqlLoader.Load(QueryGetUserIngredientByID)
	if err != nil {
		return nil, err
	}

	ingredient := &models.UserIngredient{}
	err = r.db.QueryRow(query, userID, ingredientID).Scan(
		&ingredient.ID,
		&ingredient.UserID,
		&ingredient.Name,
		&ingredient.KcalPer100g,
		&ingredient.Fats,
		&ingredient.Carbs,
		&ingredient.Proteins,
		&ingredient.GlobalIngredientID,
		&ingredient.CreatedAt,
		&ingredient.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return ingredient, nil
}

// CreateUserIngredient Create a new user ingredient
func (r *IngredientRepositoryImpl) CreateUserIngredient(userID int, name string, kcalPer100g float64, fats, carbs, proteins *float64) (*models.UserIngredient, error) {
	r.logger.Debug("Creating user ingredient",
		slog.Int("user_id", userID),
		slog.String("name", name),
		slog.Float64("kcal_per_100g", kcalPer100g))

	insertQuery, err := r.sqlLoader.Load(QueryInsertUserIngredient)
	if err != nil {
		return nil, err
	}
	result, err := r.db.Exec(insertQuery, userID, name, kcalPer100g, fats, carbs, proteins)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Get the newly created ingredient
	return r.GetUserIngredientByID(userID, int(id))
}

// UpdateUserIngredient Update an existing user ingredient
func (r *IngredientRepositoryImpl) UpdateUserIngredient(userID int, ingredientID int, name string, kcalPer100g float64, fats, carbs, proteins *float64) (*models.UserIngredient, error) {
	r.logger.Debug("Updating user ingredient",
		slog.Int("user_id", userID),
		slog.Int("ingredient_id", ingredientID),
		slog.String("name", name),
		slog.Float64("kcal_per_100g", kcalPer100g))

	updateQuery, err := r.sqlLoader.Load(QueryUpdateUserIngredient)
	if err != nil {
		return nil, err
	}
	_, err = r.db.Exec(updateQuery, name, kcalPer100g, fats, carbs, proteins, userID, ingredientID)
	if err != nil {
		return nil, err
	}

	return r.GetUserIngredientByID(userID, ingredientID)
}

// DeleteUserIngredient Delete a user ingredient
func (r *IngredientRepositoryImpl) DeleteUserIngredient(userID int, ingredientID int) error {
	r.logger.Debug("Deleting user ingredient",
		slog.Int("user_id", userID),
		slog.Int("ingredient_id", ingredientID))

	deleteQuery, err := r.sqlLoader.Load(QueryDeleteUserIngredient)
	if err != nil {
		return err
	}
	result, err := r.db.Exec(deleteQuery, userID, ingredientID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
