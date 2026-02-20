package weight

import (
	"time"

	"ypeskov/kkal-tracker/internal/models"
)

// Servicer defines the weight service contract used by handlers and other services.
type Servicer interface {
	GetWeightHistory(userID int) ([]*models.WeightHistory, error)
	GetWeightHistoryByDateRange(userID int, dateFrom, dateTo string) ([]*models.WeightHistory, error)
	CreateWeightEntry(userID int, weight float64, recordedAt *time.Time) (*models.WeightHistory, error)
	UpdateWeightEntry(id, userID int, weight float64, recordedAt *time.Time) (*models.WeightHistory, error)
	DeleteWeightEntry(id, userID int) error
}
