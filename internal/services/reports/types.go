package reports

// WeightDataPoint represents daily weight data
type WeightDataPoint struct {
	Date   string  `json:"date"`
	Weight float64 `json:"weight"`
}

// CalorieDataPoint represents daily calorie data
type CalorieDataPoint struct {
	Date     string `json:"date"`
	Calories int    `json:"calories"`
}

// ReportDataResponse contains aggregated metrics
type ReportDataResponse struct {
	WeightHistory  []WeightDataPoint  `json:"weight_history"`
	CalorieHistory []CalorieDataPoint `json:"calorie_history"`
}