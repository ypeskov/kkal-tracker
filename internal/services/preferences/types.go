package preferences

// KeyReports is the key of the Reports page settings
const KeyReports = "reports"

// DateToToday is the value of ReportsPreferences.DateTo that follows the current day of the client
const DateToToday = "today"

// ReportsPreferences are the chart settings of the Reports page. A nil range means the client default (last 30 days).
type ReportsPreferences struct {
	DateFrom           *string `json:"date_from"`
	DateTo             *string `json:"date_to"`
	ShowWeight         bool    `json:"show_weight"`
	ShowCalories       bool    `json:"show_calories"`
	ShowGoalProjection bool    `json:"show_goal_projection"`
	Period             string  `json:"period"`
	StepInterval       int     `json:"step_interval"`
}
