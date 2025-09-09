package sqlite

import (
	"database/sql"
	"log/slog"
	"time"

	"ypeskov/kkal-tracker/internal/models"
)

type IngredientRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewIngredientRepository(db *sql.DB, logger *slog.Logger) *IngredientRepository {
	return &IngredientRepository{
		db:     db,
		logger: logger,
	}
}

// Search user ingredients for autocomplete
func (r *IngredientRepository) SearchUserIngredients(userID int, query string, limit int) ([]*models.UserIngredient, error) {
	r.logger.Debug("Searching user ingredients",
		slog.Int("user_id", userID),
		slog.String("query", query),
		slog.Int("limit", limit))

	sqlQuery := `
		SELECT id, user_id, name, kcal_per_100g, fats, carbs, proteins, global_ingredient_id, created_at, updated_at
		FROM user_ingredients
		WHERE user_id = ? AND name LIKE ?
		ORDER BY name
		LIMIT ?
	`

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
func (r *IngredientRepository) GetAllUserIngredients(userID int) ([]*models.UserIngredient, error) {
	r.logger.Debug("Getting all user ingredients", slog.Int("user_id", userID))

	sqlQuery := `
		SELECT id, user_id, name, kcal_per_100g, fats, carbs, proteins, global_ingredient_id, created_at, updated_at
		FROM user_ingredients
		WHERE user_id = ?
		ORDER BY name
	`

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
func (r *IngredientRepository) GetUserIngredientByName(userID int, name string) (*models.UserIngredient, error) {
	query := `
		SELECT id, user_id, name, kcal_per_100g, fats, carbs, proteins, global_ingredient_id, created_at, updated_at
		FROM user_ingredients
		WHERE user_id = ? AND name = ?
	`

	ingredient := &models.UserIngredient{}
	err := r.db.QueryRow(query, userID, name).Scan(
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
func (r *IngredientRepository) CreateOrUpdateUserIngredient(userID int, name string, kcalPer100g float64, fats, carbs, proteins *float64) (*models.UserIngredient, error) {
	r.logger.Debug("Creating or updating user ingredient",
		slog.Int("user_id", userID),
		slog.String("name", name),
		slog.Float64("kcal_per_100g", kcalPer100g))

	// Check if ingredient already exists
	_, err := r.GetUserIngredientByName(userID, name)
	if err == nil {
		// Ingredient exists, update it
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
	insertQuery := `
		INSERT INTO user_ingredients (user_id, name, kcal_per_100g, fats, carbs, proteins)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err = r.db.Exec(insertQuery, userID, name, kcalPer100g, fats, carbs, proteins)
	if err != nil {
		return nil, err
	}

	// Get the newly created ingredient
	return r.GetUserIngredientByName(userID, name)
}

// Copy global ingredients to user ingredients table in specified language
func (r *IngredientRepository) CopyGlobalIngredientsToUser(userID int, languageCode string) error {
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
	err = tx.QueryRow("SELECT COUNT(*) FROM user_ingredients WHERE user_id = ?", userID).Scan(&count)
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
	query := `
		INSERT INTO user_ingredients (user_id, name, kcal_per_100g, fats, carbs, proteins, global_ingredient_id)
		SELECT ?, gin.name, gi.kcal_per_100g, gi.fats, gi.carbs, gi.proteins, gi.id
		FROM global_ingredients gi
		INNER JOIN global_ingredient_names gin ON gi.id = gin.ingredient_id
		WHERE gin.language_code = ?
		AND NOT EXISTS (
			SELECT 1 FROM user_ingredients ui 
			WHERE ui.user_id = ? AND ui.name = gin.name
		)
	`

	_, err = tx.Exec(query, userID, languageCode, userID)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Admin functions for global ingredients
func (r *IngredientRepository) CreateGlobalIngredient(kcalPer100g float64, fats, carbs, proteins *float64, names map[string]string) (*models.GlobalIngredient, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Insert global ingredient
	now := time.Now().UTC()
	result, err := tx.Exec(
		"INSERT INTO global_ingredients (kcal_per_100g, fats, carbs, proteins, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		kcalPer100g, fats, carbs, proteins, now, now,
	)
	if err != nil {
		return nil, err
	}

	ingredientID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	// Insert names
	for langCode, name := range names {
		_, err = tx.Exec(
			"INSERT INTO global_ingredient_names (ingredient_id, language_code, name) VALUES (?, ?, ?)",
			ingredientID, langCode, name,
		)
		if err != nil {
			return nil, err
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetGlobalIngredientByID(int(ingredientID))
}

func (r *IngredientRepository) GetGlobalIngredientByID(id int) (*models.GlobalIngredient, error) {
	query := `
		SELECT id, kcal_per_100g, fats, carbs, proteins, created_at, updated_at
		FROM global_ingredients
		WHERE id = ?
	`

	ingredient := &models.GlobalIngredient{}
	err := r.db.QueryRow(query, id).Scan(
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

func (r *IngredientRepository) getGlobalIngredientNames(ingredientID int) (map[string]string, error) {
	query := `
		SELECT language_code, name
		FROM global_ingredient_names
		WHERE ingredient_id = ?
	`

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
func (r *IngredientRepository) GetUserIngredientByID(userID int, ingredientID int) (*models.UserIngredient, error) {
	query := `
		SELECT id, user_id, name, kcal_per_100g, fats, carbs, proteins, global_ingredient_id, created_at, updated_at
		FROM user_ingredients
		WHERE user_id = ? AND id = ?
	`

	ingredient := &models.UserIngredient{}
	err := r.db.QueryRow(query, userID, ingredientID).Scan(
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
func (r *IngredientRepository) CreateUserIngredient(userID int, name string, kcalPer100g float64, fats, carbs, proteins *float64) (*models.UserIngredient, error) {
	r.logger.Debug("Creating user ingredient",
		slog.Int("user_id", userID),
		slog.String("name", name),
		slog.Float64("kcal_per_100g", kcalPer100g))

	insertQuery := `
		INSERT INTO user_ingredients (user_id, name, kcal_per_100g, fats, carbs, proteins)
		VALUES (?, ?, ?, ?, ?, ?)
	`
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
func (r *IngredientRepository) UpdateUserIngredient(userID int, ingredientID int, name string, kcalPer100g float64, fats, carbs, proteins *float64) (*models.UserIngredient, error) {
	r.logger.Debug("Updating user ingredient",
		slog.Int("user_id", userID),
		slog.Int("ingredient_id", ingredientID),
		slog.String("name", name),
		slog.Float64("kcal_per_100g", kcalPer100g))

	updateQuery := `
		UPDATE user_ingredients 
		SET name = ?, kcal_per_100g = ?, fats = ?, carbs = ?, proteins = ?, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = ? AND id = ?
	`
	_, err := r.db.Exec(updateQuery, name, kcalPer100g, fats, carbs, proteins, userID, ingredientID)
	if err != nil {
		return nil, err
	}

	return r.GetUserIngredientByID(userID, ingredientID)
}

// DeleteUserIngredient Delete a user ingredient
func (r *IngredientRepository) DeleteUserIngredient(userID int, ingredientID int) error {
	r.logger.Debug("Deleting user ingredient",
		slog.Int("user_id", userID),
		slog.Int("ingredient_id", ingredientID))

	deleteQuery := `
		DELETE FROM user_ingredients
		WHERE user_id = ? AND id = ?
	`
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
