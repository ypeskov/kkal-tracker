package calorie

import "ypeskov/kkal-tracker/internal/models"

// Servicer defines the calorie service contract used by handlers and other services.
type Servicer interface {
	CreateEntry(req *CreateEntryRequest) (*CreateEntryResult, error)
	UpdateEntry(req *UpdateEntryRequest) (*models.CalorieEntry, error)
	DeleteEntry(entryID, userID int) error
	GetEntriesByDateRange(userID int, dateFrom, dateTo string) ([]*models.CalorieEntry, error)
	GetTotalCaloriesForDate(userID int, date string) (int, error)
	GetWeeklyStats(userID int, startDate string) (map[string]int, error)
}
