package repositories

import (
	"time"

	"ypeskov/kkal-tracker/internal/models"
)

// UserRepository defines the contract for user data access
type UserRepository interface {
	Create(email, passwordHash string) (*models.User, error)
	CreateWithLanguage(email, passwordHash, languageCode string) (*models.User, error)
	GetByID(id int) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	UpdateProfile(userID int, profile *models.ProfileUpdateRequest) error
	GetLatestWeight(userID int) (*float64, error)
	AddWeightEntry(userID int, weight float64) error
}

// CalorieEntryRepository defines the contract for calorie entry data access
type CalorieEntryRepository interface {
	Create(userID int, food string, calories int, weight float64, kcalPer100g float64,
		fats, carbs, proteins *float64, mealDatetime time.Time) (*models.CalorieEntry, error)
	GetByID(id int) (*models.CalorieEntry, error)
	GetByUserID(userID int) ([]*models.CalorieEntry, error)
	GetByUserIDAndDateRange(userID int, dateFrom, dateTo string) ([]*models.CalorieEntry, error)
	Update(id, userID int, food string, calories int, weight float64, kcalPer100g float64,
		fats, carbs, proteins *float64, mealDatetime time.Time) (*models.CalorieEntry, error)
	Delete(id, userID int) error
}

// IngredientRepository defines the contract for ingredient data access
type IngredientRepository interface {
	// SearchUserIngredients User ingredients
	SearchUserIngredients(userID int, query string, limit int) ([]*models.UserIngredient, error)
	GetAllUserIngredients(userID int) ([]*models.UserIngredient, error)
	GetUserIngredientByName(userID int, name string) (*models.UserIngredient, error)
	GetUserIngredientByID(userID int, ingredientID int) (*models.UserIngredient, error)
	CreateOrUpdateUserIngredient(userID int, name string, kcalPer100g float64,
		fats, carbs, proteins *float64) (*models.UserIngredient, error)
	CreateUserIngredient(userID int, name string, kcalPer100g float64,
		fats, carbs, proteins *float64) (*models.UserIngredient, error)
	UpdateUserIngredient(userID int, ingredientID int, name string, kcalPer100g float64,
		fats, carbs, proteins *float64) (*models.UserIngredient, error)
	DeleteUserIngredient(userID int, ingredientID int) error
	CopyGlobalIngredientsToUser(userID int, languageCode string) error

	// CreateGlobalIngredient Global ingredients (admin)
	CreateGlobalIngredient(kcalPer100g float64, fats, carbs, proteins *float64,
		names map[string]string) (*models.GlobalIngredient, error)
	GetGlobalIngredientByID(id int) (*models.GlobalIngredient, error)
}
