package models

import (
	"database/sql"
	"time"
)

// Global ingredients (admin managed)
type GlobalIngredient struct {
	ID          int               `json:"id"`
	KcalPer100g float64           `json:"kcalPer100g"`
	Fats        *float64          `json:"fats,omitempty"`
	Carbs       *float64          `json:"carbs,omitempty"`
	Proteins    *float64          `json:"proteins,omitempty"`
	Names       map[string]string `json:"names"` // language_code -> name
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

// User-specific ingredients (copied from global on registration)
type UserIngredient struct {
	ID                 int       `json:"id"`
	UserID             int       `json:"user_id"`
	Name               string    `json:"name"`
	KcalPer100g        float64   `json:"kcalPer100g"`
	Fats               *float64  `json:"fats,omitempty"`
	Carbs              *float64  `json:"carbs,omitempty"`
	Proteins           *float64  `json:"proteins,omitempty"`
	GlobalIngredientID *int      `json:"global_ingredient_id,omitempty"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type IngredientRepository struct {
	db *sql.DB
}

func NewIngredientRepository(db *sql.DB) *IngredientRepository {
	return &IngredientRepository{db: db}
}

// Search user ingredients for autocomplete
func (r *IngredientRepository) SearchUserIngredients(userID int, query string, limit int) ([]*UserIngredient, error) {
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

	var ingredients []*UserIngredient
	for rows.Next() {
		ingredient := &UserIngredient{}
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
func (r *IngredientRepository) GetAllUserIngredients(userID int) ([]*UserIngredient, error) {
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

	var ingredients []*UserIngredient
	for rows.Next() {
		ingredient := &UserIngredient{}
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

// Copy global ingredients to user ingredients table in specified language
func (r *IngredientRepository) CopyGlobalIngredientsToUser(userID int, languageCode string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO user_ingredients (user_id, name, kcal_per_100g, fats, carbs, proteins, global_ingredient_id)
		SELECT ?, gin.name, gi.kcal_per_100g, gi.fats, gi.carbs, gi.proteins, gi.id
		FROM global_ingredients gi
		INNER JOIN global_ingredient_names gin ON gi.id = gin.ingredient_id
		WHERE gin.language_code = ?
	`

	_, err = tx.Exec(query, userID, languageCode)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Admin functions for global ingredients
func (r *IngredientRepository) CreateGlobalIngredient(kcalPer100g float64, fats, carbs, proteins *float64, names map[string]string) (*GlobalIngredient, error) {
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

func (r *IngredientRepository) GetGlobalIngredientByID(id int) (*GlobalIngredient, error) {
	query := `
		SELECT id, kcal_per_100g, fats, carbs, proteins, created_at, updated_at
		FROM global_ingredients
		WHERE id = ?
	`
	
	ingredient := &GlobalIngredient{}
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