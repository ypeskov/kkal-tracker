package reports

type ReportDataResponse struct {
	WeightHistory  []WeightDataPoint  `json:"weight_history"`
	CalorieHistory []CalorieDataPoint `json:"calorie_history"`
}

type WeightDataPoint struct {
	Date   string  `json:"date"`
	Weight float64 `json:"weight"`
}

type CalorieDataPoint struct {
	Date     string `json:"date"`
	Calories int    `json:"calories"`
}