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

	// Check if this user_id already has ingredients (shouldn't happen, but let's verify)
	var existingIngredientsCount int
	checkQuery := `SELECT COUNT(*) FROM user_ingredients WHERE user_id = ?`
	if checkErr := tx.QueryRow(checkQuery, id).Scan(&existingIngredientsCount); checkErr == nil {
		if existingIngredientsCount > 0 {
			r.logger.Warn("User already has ingredients before copying",
				slog.Int64("user_id", id),
				slog.Int("existing_count", existingIngredientsCount))
		}
	}

	// Copy global ingredients to user_ingredients
	copyQuery, err := r.sqlLoader.Load(QueryCopyGlobalIngredients)
	if err != nil {
		r.logger.Error("Failed to load copy query", "error", err)
		return nil, err
	}

	// Log duplicate ingredient names before copying
	checkDuplicatesQuery := `
		SELECT gin.name, COUNT(DISTINCT gi.id) as cnt
		FROM global_ingredients gi
		JOIN global_ingredient_names gin ON gi.id = gin.ingredient_id
		WHERE gin.language_code = ?
		GROUP BY gin.name
		HAVING cnt > 1
	`
	rows, queryErr := tx.Query(checkDuplicatesQuery, languageCode)
	if queryErr == nil {
		var duplicates []string
		for rows.Next() {
			var name string
			var count int
			if scanErr := rows.Scan(&name, &count); scanErr == nil {
				duplicates = append(duplicates, name)
			}
		}
		rows.Close()
		if len(duplicates) > 0 {
			r.logger.Warn("Found duplicate ingredient names",
				slog.String("language", languageCode),
				slog.Int("duplicate_count", len(duplicates)),
				slog.Any("duplicate_names", duplicates))
		}
	}

	r.logger.Debug("Executing ingredient copy query",
		slog.Int64("user_id", id),
		slog.String("language", languageCode),
		slog.String("query", copyQuery))

	result, err = tx.Exec(copyQuery, id, languageCode, languageCode)
	if err != nil {
		// Log what would have been inserted to debug the constraint violation
		debugQuery := `
			SELECT gin.name, gi.id, gi.kcal_per_100g
			FROM global_ingredients gi
			JOIN global_ingredient_names gin ON gi.id = gin.ingredient_id
			WHERE gin.language_code = ?
			ORDER BY gin.name
		`
		debugRows, debugErr := tx.Query(debugQuery, languageCode)
		if debugErr == nil {
			var ingredientNames []string
			for debugRows.Next() {
				var name string
				var ingredientID int
				var kcal float64
				if scanErr := debugRows.Scan(&name, &ingredientID, &kcal); scanErr == nil {
					ingredientNames = append(ingredientNames, name)
				}
			}
			debugRows.Close()
			r.logger.Error("Failed to copy ingredients - ingredient list",
				"error", err,
				slog.Int64("user_id", id),
				slog.String("language", languageCode),
				slog.Int("ingredient_count", len(ingredientNames)),
				slog.Any("ingredient_names", ingredientNames))
		}

		// Check if user already has ingredients (shouldn't happen in a new user transaction)
		checkExistingQuery := `SELECT COUNT(*) FROM user_ingredients WHERE user_id = ?`
		var existingCount int
		if countErr := tx.QueryRow(checkExistingQuery, id).Scan(&existingCount); countErr == nil {
			r.logger.Error("Existing user_ingredients check",
				slog.Int64("user_id", id),
				slog.Int("existing_count", existingCount))
		}

		r.logger.Error("Failed to copy ingredients",
			"error", err,
			slog.Int64("user_id", id),
			slog.String("language", languageCode))
		return nil, err
	}

	rowsInserted, _ := result.RowsAffected()
	r.logger.Debug("Ingredients copied successfully",
		slog.Int64("user_id", id),
		slog.Int64("rows_inserted", rowsInserted))

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
