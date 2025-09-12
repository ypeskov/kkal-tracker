package sqlite

import (
	"database/sql"
	"log/slog"

	"ypeskov/kkal-tracker/internal/models"
)

type UserRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewUserRepository(db *sql.DB, logger *slog.Logger) *UserRepository {
	return &UserRepository{
		db:     db,
		logger: logger,
	}
}

func (r *UserRepository) Create(email, passwordHash string) (*models.User, error) {
	r.logger.Debug("Creating user", slog.String("email", email))
	return r.CreateWithLanguage(email, passwordHash, "en_US")
}

func (r *UserRepository) CreateWithLanguage(email, passwordHash, languageCode string) (*models.User, error) {
	r.logger.Debug("Creating user with language",
		slog.String("email", email),
		slog.String("language", languageCode))

	// Start transaction to ensure user creation and ingredient copying are atomic
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Create the user
	query := `
		INSERT INTO users (email, password_hash)
		VALUES (?, ?)
	`

	result, err := tx.Exec(query, email, passwordHash)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	r.logger.Debug("User created in database", slog.Int64("id", id))

	// Copy global ingredients to user_ingredients
	copyQuery := `
		INSERT INTO user_ingredients (user_id, name, kcal_per_100g, fats, carbs, proteins, global_ingredient_id)
		SELECT ?, gin.name, gi.kcal_per_100g, gi.fats, gi.carbs, gi.proteins, gi.id
		FROM global_ingredients gi
		JOIN global_ingredient_names gin ON gi.id = gin.ingredient_id
		WHERE gin.language_code = ?
	`

	_, err = tx.Exec(copyQuery, id, languageCode)
	if err != nil {
		return nil, err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	r.logger.Debug("Transaction of creating user with language committed successfully", slog.Int64("user_id", id))

	return r.GetByID(int(id))
}

func (r *UserRepository) GetByID(id int) (*models.User, error) {
	r.logger.Debug("Getting user by ID", slog.Int("id", id))

	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE id = ?
	`

	user := &models.User{}
	err := r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	r.logger.Debug("Getting user by email", slog.String("email", email))

	query := `
		SELECT id, email, password_hash, created_at, updated_at
		FROM users
		WHERE email = ?
	`

	user := &models.User{}
	err := r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return user, nil
}
