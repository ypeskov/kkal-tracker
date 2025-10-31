package repositories

import (
	"database/sql"
	"log/slog"

	"ypeskov/kkal-tracker/internal/models"
)

type UserRepositoryImpl struct {
	db        *sql.DB
	logger    *slog.Logger
	sqlLoader *SqlLoaderInstance
}

func NewUserRepository(db *sql.DB, logger *slog.Logger, dialect Dialect) *UserRepositoryImpl {
	sqlLoader := NewSqlLoader(dialect)

	return &UserRepositoryImpl{
		db:        db,
		logger:    logger,
		sqlLoader: sqlLoader,
	}
}

func (r *UserRepositoryImpl) Create(email, passwordHash string) (*models.User, error) {
	r.logger.Debug("Creating user", slog.String("email", email))
	return r.CreateWithLanguage(email, passwordHash, "en_US", true)
}

func (r *UserRepositoryImpl) CreateWithLanguage(email, passwordHash, languageCode string, isActive bool) (*models.User, error) {
	r.logger.Debug("Creating user with language",
		slog.String("email", email),
		slog.String("language", languageCode),
		slog.Bool("is_active", isActive))

	// Start transaction to ensure user creation and ingredient copying are atomic
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Create the user with language preference and activation status
	query, err := r.sqlLoader.Load(QueryInsertUser)
	if err != nil {
		return nil, err
	}

	result, err := tx.Exec(query, email, passwordHash, languageCode, isActive)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	r.logger.Debug("User created in database", slog.Int64("id", id), slog.Bool("is_active", isActive))

	// Copy global ingredients to user_ingredients
	copyQuery, err := r.sqlLoader.Load(QueryCopyGlobalIngredients)
	if err != nil {
		return nil, err
	}

	result, err = tx.Exec(copyQuery, id, languageCode, languageCode)
	if err != nil {
		r.logger.Error("Failed to copy ingredients",
			"error", err,
			slog.Int64("user_id", id),
			slog.String("language", languageCode))
		return nil, err
	}

	rowsInserted, _ := result.RowsAffected()
	r.logger.Debug("Ingredients copied",
		slog.Int64("user_id", id),
		slog.Int64("count", rowsInserted))

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return nil, err
	}

	r.logger.Debug("Transaction of creating user with language committed successfully", slog.Int64("user_id", id))

	return r.GetByID(int(id))
}

func (r *UserRepositoryImpl) GetByID(id int) (*models.User, error) {
	r.logger.Debug("Getting user by ID", slog.Int("id", id))

	query, err := r.sqlLoader.Load(QueryGetUserByID)
	if err != nil {
		return nil, err
	}

	user := &models.User{}
	var language sql.NullString
	err = r.db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.IsActive,
		&user.FirstName,
		&user.LastName,
		&user.Age,
		&user.Height,
		&language,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == nil && language.Valid {
		user.Language = &language.String
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *UserRepositoryImpl) GetByEmail(email string) (*models.User, error) {
	r.logger.Debug("Getting user by email", slog.String("email", email))

	query, err := r.sqlLoader.Load(QueryGetUserByEmail)
	if err != nil {
		return nil, err
	}

	user := &models.User{}
	var language sql.NullString
	err = r.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.IsActive,
		&user.FirstName,
		&user.LastName,
		&user.Age,
		&user.Height,
		&language,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err == nil && language.Valid {
		user.Language = &language.String
	}

	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateProfile updates user profile information
func (r *UserRepositoryImpl) UpdateProfile(userID int, firstName, lastName *string, email string, age *int, height *float64, _ *float64, language string) error {
	r.logger.Debug("Updating user profile", slog.Int("user_id", userID))

	query, err := r.sqlLoader.Load(QueryUpdateUserProfile)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(query,
		firstName,
		lastName,
		email,
		age,
		height,
		language,
		userID,
	)

	return err
}

// AddWeightEntry adds a new weight entry to the history
func (r *UserRepositoryImpl) AddWeightEntry(userID int, weight float64) error {
	r.logger.Debug("Adding weight entry", slog.Int("user_id", userID), slog.Float64("weight", weight))

	query, err := r.sqlLoader.Load(QueryAddWeightEntry)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(query, userID, weight)
	return err
}

// ActivateUser sets the user's is_active status to true
func (r *UserRepositoryImpl) ActivateUser(userID int) error {
	r.logger.Debug("Activating user", slog.Int("user_id", userID))

	query, err := r.sqlLoader.Load(QueryActivateUser)
	if err != nil {
		return err
	}

	result, err := r.db.Exec(query, userID)
	if err != nil {
		r.logger.Error("Failed to activate user", "error", err, "user_id", userID)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		r.logger.Warn("No user activated (user not found)", "user_id", userID)
		return ErrNotFound
	}

	r.logger.Info("User activated successfully", "user_id", userID)
	return nil
}

// Delete removes a user from the database
func (r *UserRepositoryImpl) Delete(userID int) error {
	r.logger.Debug("Deleting user", slog.Int("user_id", userID))

	query, err := r.sqlLoader.Load(QueryDeleteUser)
	if err != nil {
		return err
	}

	result, err := r.db.Exec(query, userID)
	if err != nil {
		r.logger.Error("Failed to delete user", "error", err, "user_id", userID)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		r.logger.Warn("No user deleted (user not found)", "user_id", userID)
		return ErrNotFound
	}

	r.logger.Info("User deleted successfully", "user_id", userID)
	return nil
}
