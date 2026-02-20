package reports

// Servicer defines the reports service contract used by handlers.
type Servicer interface {
	GetAggregatedMetrics(userID int, dateFrom, dateTo string) (*ReportDataResponse, error)
}
