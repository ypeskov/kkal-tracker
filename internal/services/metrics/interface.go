package metrics

// Servicer defines the metrics service contract used by handlers.
type Servicer interface {
	GetHealthMetrics(userID int) (*HealthMetrics, error)
	CalculateTDEE(userID int, activityLevel ActivityLevel) (*float64, error)
}
