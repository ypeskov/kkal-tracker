package repositories

import "fmt"

const (
	// User queries
	QueryInsertUser            = "insertUser"
	QueryGetUserByEmail        = "getUserByEmail"
	QueryGetUserByID           = "getUserByID"
	QueryCopyGlobalIngredients = "copyGlobalIngredients"
	QueryUpdateUserProfile     = "updateUserProfile"
	QueryAddWeightEntry        = "addWeightEntry"
	QueryActivateUser          = "activateUser"
	QueryDeleteUser            = "deleteUser"

	// Group of CalorieEntries queries
	QueryInsertCalorieEntry           = "insertCalorieEntry"
	QueryGetCalorieEntryByID          = "getCalorieEntryByID"
	QueryGetCalorieEntriesByUserID    = "getCalorieEntriesByUserID"
	QueryGetCalorieEntriesByDateRange = "getCalorieEntriesByDateRange"
	QueryUpdateCalorieEntry           = "updateCalorieEntry"
	QueryDeleteCalorieEntry           = "deleteCalorieEntry"

	// Weight History queries
	QueryGetWeightHistory            = "getWeightHistory"
	QueryGetWeightHistoryByDateRange = "getWeightHistoryByDateRange"
	QueryGetLatestWeightHistory      = "getLatestWeightHistory"
	QueryCreateWeightHistory         = "createWeightHistory"
	QueryUpdateWeightHistory         = "updateWeightHistory"
	QueryDeleteWeightHistory         = "deleteWeightHistory"

	// Ingredient queries
	QueryGetAllUserIngredients       = "getAllUserIngredients"
	QueryGetUserIngredientByName     = "getUserIngredientByName"
	QueryGetUserIngredientByID       = "getUserIngredientByID"
	QueryUpdateUserIngredient        = "updateUserIngredient"
	QueryInsertUserIngredient        = "insertUserIngredient"
	QueryDeleteUserIngredient        = "deleteUserIngredient"
	QueryCopyGlobalIngredientsToUser = "copyGlobalIngredientsToUser"
	QueryCountUserIngredients        = "countUserIngredients"
	QueryInsertGlobalIngredient      = "insertGlobalIngredient"
	QueryInsertGlobalIngredientName  = "insertGlobalIngredientName"
	QueryGetGlobalIngredientByID     = "getGlobalIngredientByID"
	QueryGetGlobalIngredientNames    = "getGlobalIngredientNames"

	// Activation Token queries
	QueryCreateActivationToken       = "createActivationToken"
	QueryGetActivationTokenByToken   = "getActivationTokenByToken"
	QueryDeleteActivationToken       = "deleteActivationToken"
	QueryDeleteExpiredActivationTokens = "deleteExpiredActivationTokens"
)

// buildKey creates a query key by combining query name and dialect
func buildKey(queryName string, dialect Dialect) string {
	return fmt.Sprintf("%s.%s", queryName, dialect)
}

// getQueries returns the queries map with dynamically built keys
func getQueries() map[string]string {
	return map[string]string{
		// User queries
		//language=SQLite
		buildKey(QueryInsertUser, DialectSQLite): `
		INSERT INTO users (email, password_hash, language, is_active)
		VALUES (?, ?, ?, ?)
		RETURNING id;
	`,
		// language=PostgreSQL
		buildKey(QueryInsertUser, DialectPostgres): `
		INSERT INTO users (email, password_hash, language, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id;
	`,

		buildKey(QueryGetUserByEmail, DialectSQLite): `
		SELECT id, email, password_hash, is_active, first_name, last_name, age, height, language, created_at, updated_at
		FROM users
		WHERE email = ?
	`,
		buildKey(QueryGetUserByEmail, DialectPostgres): `
		SELECT id, email, password_hash, is_active, first_name, last_name, age, height, language, created_at, updated_at
		FROM users
		WHERE email = $1
	`,

		buildKey(QueryGetUserByID, DialectSQLite): `
		SELECT id, email, password_hash, is_active, first_name, last_name, age, height, language, created_at, updated_at
		FROM users
		WHERE id = ?
	`,
		buildKey(QueryGetUserByID, DialectPostgres): `
		SELECT id, email, password_hash, is_active, first_name, last_name, age, height, language, created_at, updated_at
		FROM users
		WHERE id = $1
	`,

		buildKey(QueryCopyGlobalIngredients, DialectSQLite): `
		INSERT INTO user_ingredients (user_id, name, kcal_per_100g, fats, carbs, proteins, global_ingredient_id)
		SELECT ?, gin.name, gi.kcal_per_100g, gi.fats, gi.carbs, gi.proteins, gi.id
		FROM global_ingredients gi
		JOIN global_ingredient_names gin ON gi.id = gin.ingredient_id
		WHERE gin.language_code = ?
		GROUP BY gin.name
	`,
		buildKey(QueryCopyGlobalIngredients, DialectPostgres): `
		INSERT INTO user_ingredients (user_id, name, kcal_per_100g, fats, carbs, proteins, global_ingredient_id)
		SELECT DISTINCT ON (gin.name) $1, gin.name, gi.kcal_per_100g, gi.fats, gi.carbs, gi.proteins, gi.id
		FROM global_ingredients gi
		JOIN global_ingredient_names gin ON gi.id = gin.ingredient_id
		WHERE gin.language_code = $2
	`,

		buildKey(QueryUpdateUserProfile, DialectSQLite): `
		UPDATE users
		SET first_name = ?, last_name = ?, email = ?, age = ?, height = ?, language = ?, updated_at = datetime('now')
		WHERE id = ?
	`,

		buildKey(QueryUpdateUserProfile, DialectPostgres): `UPDATE users
		SET first_name = $1, last_name = $2, email = $3, age = $4, height = $5,
		    language = $6, updated_at = NOW()
		WHERE id = $7
	`,

		buildKey(QueryAddWeightEntry, DialectSQLite): `
		INSERT INTO weight_history (user_id, weight)
		VALUES (?, ?)
	`,
		buildKey(QueryAddWeightEntry, DialectPostgres): `
		INSERT INTO weight_history (user_id, weight)
		VALUES ($1, $2)
	`,

		buildKey(QueryActivateUser, DialectSQLite): `
		UPDATE users
		SET is_active = 1, updated_at = datetime('now')
		WHERE id = ?
	`,
		buildKey(QueryActivateUser, DialectPostgres): `
		UPDATE users
		SET is_active = true, updated_at = NOW()
		WHERE id = $1
	`,

		buildKey(QueryDeleteUser, DialectSQLite): `
		DELETE FROM users WHERE id = ?
	`,
		buildKey(QueryDeleteUser, DialectPostgres): `
		DELETE FROM users WHERE id = $1
	`,

		// Weight History queries
		buildKey(QueryGetWeightHistory, DialectSQLite): `
		SELECT id, user_id, weight, recorded_at, created_at
		FROM weight_history
		WHERE user_id = ?
		ORDER BY recorded_at DESC
	`,
		buildKey(QueryGetWeightHistory, DialectPostgres): `
		SELECT id, user_id, weight, recorded_at, created_at
		FROM weight_history
		WHERE user_id = $1
		ORDER BY recorded_at DESC
	`,

		buildKey(QueryGetWeightHistoryByDateRange, DialectSQLite): `
		SELECT id, user_id, weight, recorded_at, created_at
		FROM weight_history
		WHERE user_id = ? AND substr(recorded_at, 1, 10) >= ? AND substr(recorded_at, 1, 10) <= ?
		ORDER BY recorded_at ASC
	`,
		buildKey(QueryGetWeightHistoryByDateRange, DialectPostgres): `
		SELECT id, user_id, weight, recorded_at, created_at
		FROM weight_history
		WHERE user_id = $1 AND DATE(recorded_at) >= $2 AND DATE(recorded_at) <= $3
		ORDER BY recorded_at ASC
	`,

		buildKey(QueryGetLatestWeightHistory, DialectSQLite): `
		SELECT id, user_id, weight, recorded_at, created_at
		FROM weight_history
		WHERE user_id = ?
		ORDER BY recorded_at DESC
		LIMIT 1
	`,
		buildKey(QueryGetLatestWeightHistory, DialectPostgres): `
		SELECT id, user_id, weight, recorded_at, created_at
		FROM weight_history
		WHERE user_id = $1
		ORDER BY recorded_at DESC
		LIMIT 1
	`,

		buildKey(QueryCreateWeightHistory, DialectSQLite): `
		INSERT INTO weight_history (user_id, weight, recorded_at)
		VALUES (?, ?, ?)
	`,
		buildKey(QueryCreateWeightHistory, DialectPostgres): `
		INSERT INTO weight_history (user_id, weight, recorded_at)
		VALUES ($1, $2, $3)
	`,

		buildKey(QueryUpdateWeightHistory, DialectSQLite): `
		UPDATE weight_history
		SET weight = ?, recorded_at = ?
		WHERE id = ? AND user_id = ?
	`,
		buildKey(QueryUpdateWeightHistory, DialectPostgres): `
		UPDATE weight_history
		SET weight = $1, recorded_at = $2
		WHERE id = $3 AND user_id = $4
	`,

		buildKey(QueryDeleteWeightHistory, DialectSQLite): `
		DELETE FROM weight_history
		WHERE id = ? AND user_id = ?
	`,
		buildKey(QueryDeleteWeightHistory, DialectPostgres): `
		DELETE FROM weight_history
		WHERE id = $1 AND user_id = $2
	`,

		// CalorieEntry queries
		buildKey(QueryInsertCalorieEntry, DialectSQLite): `
		INSERT INTO calorie_entries (user_id, food, calories, weight, kcal_per_100g, fats, carbs, proteins, meal_datetime, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		buildKey(QueryInsertCalorieEntry, DialectPostgres): `
		INSERT INTO calorie_entries (user_id, food, calories, weight, kcal_per_100g, fats, carbs, proteins, meal_datetime, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`,

		buildKey(QueryGetCalorieEntryByID, DialectSQLite): `
		SELECT id, user_id, food, calories, weight, kcal_per_100g, fats, carbs, proteins, meal_datetime, updated_at, created_at
		FROM calorie_entries
		WHERE id = ?
	`,
		buildKey(QueryGetCalorieEntryByID, DialectPostgres): `
		SELECT id, user_id, food, calories, weight, kcal_per_100g, fats, carbs, proteins, meal_datetime, updated_at, created_at
		FROM calorie_entries
		WHERE id = $1
	`,

		buildKey(QueryGetCalorieEntriesByUserID, DialectSQLite): `
		SELECT id, user_id, food, calories, weight, kcal_per_100g, fats, carbs, proteins, meal_datetime, updated_at, created_at
		FROM calorie_entries
		WHERE user_id = ?
		ORDER BY meal_datetime DESC, created_at DESC
	`,
		buildKey(QueryGetCalorieEntriesByUserID, DialectPostgres): `
		SELECT id, user_id, food, calories, weight, kcal_per_100g, fats, carbs, proteins, meal_datetime, updated_at, created_at
		FROM calorie_entries
		WHERE user_id = $1
		ORDER BY meal_datetime DESC, created_at DESC
	`,

		buildKey(QueryGetCalorieEntriesByDateRange, DialectSQLite): `
		SELECT id, user_id, food, calories, weight, kcal_per_100g, fats, carbs, proteins, meal_datetime, updated_at, created_at
		FROM calorie_entries
		WHERE user_id = ? AND (
			strftime('%Y-%m-%d', meal_datetime) BETWEEN ? AND ?
			OR strftime('%Y-%m-%d', substr(meal_datetime, 1, 19)) BETWEEN ? AND ?
		)
		ORDER BY meal_datetime DESC
	`,
		buildKey(QueryGetCalorieEntriesByDateRange, DialectPostgres): `
		SELECT id, user_id, food, calories, weight, kcal_per_100g, fats, carbs, proteins, meal_datetime, updated_at, created_at
		FROM calorie_entries
		WHERE user_id = $1 AND DATE(meal_datetime) BETWEEN $2 AND $3
		ORDER BY meal_datetime DESC
	`,

		buildKey(QueryUpdateCalorieEntry, DialectSQLite): `
		UPDATE calorie_entries
		SET food = ?, calories = ?, weight = ?, kcal_per_100g = ?, fats = ?, carbs = ?, proteins = ?, meal_datetime = ?, updated_at = ?
		WHERE id = ? AND user_id = ?
	`,
		buildKey(QueryUpdateCalorieEntry, DialectPostgres): `
		UPDATE calorie_entries
		SET food = $1, calories = $2, weight = $3, kcal_per_100g = $4, fats = $5, carbs = $6, proteins = $7, meal_datetime = $8, updated_at = $9
		WHERE id = $10 AND user_id = $11
	`,

		buildKey(QueryDeleteCalorieEntry, DialectSQLite): `
		DELETE FROM calorie_entries
		WHERE id = ? AND user_id = ?
	`,
		buildKey(QueryDeleteCalorieEntry, DialectPostgres): `
		DELETE FROM calorie_entries
		WHERE id = $1 AND user_id = $2
	`,

		// Ingredient queries
		buildKey(QueryGetAllUserIngredients, DialectSQLite): `
		SELECT id, user_id, name, kcal_per_100g, fats, carbs, proteins, global_ingredient_id, created_at, updated_at
		FROM user_ingredients
		WHERE user_id = ?
		ORDER BY name
	`,
		//noinspection SqlDialectInspection,SqlResolve
		buildKey(QueryGetAllUserIngredients, DialectPostgres): `
		SELECT id, user_id, name, kcal_per_100g, fats, carbs, proteins, global_ingredient_id, created_at, updated_at
		FROM user_ingredients
		WHERE user_id = $1
		ORDER BY name
	`,

		buildKey(QueryGetUserIngredientByName, DialectSQLite): `
		SELECT id, user_id, name, kcal_per_100g, fats, carbs, proteins, global_ingredient_id, created_at, updated_at
		FROM user_ingredients
		WHERE user_id = ? AND name = ?
	`,
		buildKey(QueryGetUserIngredientByName, DialectPostgres): `
		SELECT id, user_id, name, kcal_per_100g, fats, carbs, proteins, global_ingredient_id, created_at, updated_at
		FROM user_ingredients
		WHERE user_id = $1 AND name = $2
	`,

		buildKey(QueryGetUserIngredientByID, DialectSQLite): `
		SELECT id, user_id, name, kcal_per_100g, fats, carbs, proteins, global_ingredient_id, created_at, updated_at
		FROM user_ingredients
		WHERE user_id = ? AND id = ?
	`,
		buildKey(QueryGetUserIngredientByID, DialectPostgres): `
		SELECT id, user_id, name, kcal_per_100g, fats, carbs, proteins, global_ingredient_id, created_at, updated_at
		FROM user_ingredients
		WHERE user_id = $1 AND id = $2
	`,

		buildKey(QueryUpdateUserIngredient, DialectSQLite): `
		UPDATE user_ingredients
		SET name = ?, kcal_per_100g = ?, fats = ?, carbs = ?, proteins = ?, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = ? AND id = ?
	`,
		buildKey(QueryUpdateUserIngredient, DialectPostgres): `
		UPDATE user_ingredients
		SET name = $1, kcal_per_100g = $2, fats = $3, carbs = $4, proteins = $5, updated_at = CURRENT_TIMESTAMP
		WHERE user_id = $6 AND id = $7
	`,

		buildKey(QueryInsertUserIngredient, DialectSQLite): `
		INSERT INTO user_ingredients (user_id, name, kcal_per_100g, fats, carbs, proteins)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		buildKey(QueryInsertUserIngredient, DialectPostgres): `
		INSERT INTO user_ingredients (user_id, name, kcal_per_100g, fats, carbs, proteins)
		VALUES ($1, $2, $3, $4, $5, $6)
	`,

		buildKey(QueryDeleteUserIngredient, DialectSQLite): `
		DELETE FROM user_ingredients
		WHERE user_id = ? AND id = ?
	`,
		buildKey(QueryDeleteUserIngredient, DialectPostgres): `
		DELETE FROM user_ingredients
		WHERE user_id = $1 AND id = $2
	`,

		buildKey(QueryCopyGlobalIngredientsToUser, DialectSQLite): `
		INSERT INTO user_ingredients (user_id, name, kcal_per_100g, fats, carbs, proteins, global_ingredient_id)
		SELECT ?, gin.name, gi.kcal_per_100g, gi.fats, gi.carbs, gi.proteins, gi.id
		FROM global_ingredients gi
		INNER JOIN global_ingredient_names gin ON gi.id = gin.ingredient_id
		WHERE gin.language_code = ?
		AND NOT EXISTS (
			SELECT 1 FROM user_ingredients ui
			WHERE ui.user_id = ? AND ui.name = gin.name
		)
	`,
		buildKey(QueryCopyGlobalIngredientsToUser, DialectPostgres): `
		INSERT INTO user_ingredients (user_id, name, kcal_per_100g, fats, carbs, proteins, global_ingredient_id)
		SELECT $1, gin.name, gi.kcal_per_100g, gi.fats, gi.carbs, gi.proteins, gi.id
		FROM global_ingredients gi
		INNER JOIN global_ingredient_names gin ON gi.id = gin.ingredient_id
		WHERE gin.language_code = $2
		AND NOT EXISTS (
			SELECT 1 FROM user_ingredients ui
			WHERE ui.user_id = $1 AND ui.name = gin.name
		)
	`,

		buildKey(QueryCountUserIngredients, DialectSQLite): `
		SELECT COUNT(*) FROM user_ingredients WHERE user_id = ?
	`,
		buildKey(QueryCountUserIngredients, DialectPostgres): `
		SELECT COUNT(*) FROM user_ingredients WHERE user_id = $1
	`,

		buildKey(QueryInsertGlobalIngredient, DialectSQLite): `
		INSERT INTO global_ingredients (kcal_per_100g, fats, carbs, proteins, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`,
		buildKey(QueryInsertGlobalIngredient, DialectPostgres): `
		INSERT INTO global_ingredients (kcal_per_100g, fats, carbs, proteins, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`,

		buildKey(QueryInsertGlobalIngredientName, DialectSQLite): `
		INSERT INTO global_ingredient_names (ingredient_id, language_code, name)
		VALUES (?, ?, ?)
	`,
		buildKey(QueryInsertGlobalIngredientName, DialectPostgres): `
		INSERT INTO global_ingredient_names (ingredient_id, language_code, name)
		VALUES ($1, $2, $3)
	`,

		buildKey(QueryGetGlobalIngredientByID, DialectSQLite): `
		SELECT id, kcal_per_100g, fats, carbs, proteins, created_at, updated_at
		FROM global_ingredients
		WHERE id = ?
	`,
		buildKey(QueryGetGlobalIngredientByID, DialectPostgres): `
		SELECT id, kcal_per_100g, fats, carbs, proteins, created_at, updated_at
		FROM global_ingredients
		WHERE id = $1
	`,

		buildKey(QueryGetGlobalIngredientNames, DialectSQLite): `
		SELECT language_code, name
		FROM global_ingredient_names
		WHERE ingredient_id = ?
	`,
		buildKey(QueryGetGlobalIngredientNames, DialectPostgres): `
		SELECT language_code, name
		FROM global_ingredient_names
		WHERE ingredient_id = $1
	`,

		// Activation Token queries
		buildKey(QueryCreateActivationToken, DialectSQLite): `
		INSERT INTO activation_tokens (user_id, token, expires_at)
		VALUES (?, ?, ?)
	`,
		buildKey(QueryCreateActivationToken, DialectPostgres): `
		INSERT INTO activation_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
		RETURNING id
	`,

		buildKey(QueryGetActivationTokenByToken, DialectSQLite): `
		SELECT id, user_id, token, created_at, expires_at
		FROM activation_tokens
		WHERE token = ?
	`,
		buildKey(QueryGetActivationTokenByToken, DialectPostgres): `
		SELECT id, user_id, token, created_at, expires_at
		FROM activation_tokens
		WHERE token = $1
	`,

		buildKey(QueryDeleteActivationToken, DialectSQLite): `
		DELETE FROM activation_tokens WHERE token = ?
	`,
		buildKey(QueryDeleteActivationToken, DialectPostgres): `
		DELETE FROM activation_tokens WHERE token = $1
	`,

		buildKey(QueryDeleteExpiredActivationTokens, DialectSQLite): `
		DELETE FROM activation_tokens WHERE expires_at < ?
	`,
		buildKey(QueryDeleteExpiredActivationTokens, DialectPostgres): `
		DELETE FROM activation_tokens WHERE expires_at < $1
	`,
	}
}
