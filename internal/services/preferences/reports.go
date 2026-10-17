package preferences

var (
	reportPeriods       = []string{"daily", "weekly", "monthly", "yearly"}
	reportStepIntervals = []int{1, 5, 10, 15}
)

// Messages of the invalid fields of the reports preferences, in the order the fields are checked
const (
	msgDateFrom     = "date_from must be a date (YYYY-MM-DD) from 1900-01-01 to 2100-12-31, or null"
	msgDateTo       = `date_to must be a date (YYYY-MM-DD) from 1900-01-01 to 2100-12-31, "today", or null`
	msgDateRange    = "date_from and date_to must both be set or both be null"
	msgPeriod       = "period must be one of daily, weekly, monthly, yearly"
	msgStepInterval = "step_interval must be one of 1, 5, 10, 15"
)

// reportsDefinition is the reports key: the chart settings of the Reports page
var reportsDefinition = definition{
	version:  1,
	defaults: func() any { return defaultReports() },
	read:     func(f fields) (any, []fieldError) { return readReports(f) },
}

// defaultReports are the settings of a user who has saved nothing: the client's rolling range, every chart on, daily
func defaultReports() ReportsPreferences {
	return ReportsPreferences{
		ShowWeight:         true,
		ShowCalories:       true,
		ShowGoalProjection: true,
		Period:             "daily",
		StepInterval:       1,
	}
}

// readReports reads the reports settings from a request body or a stored value. Every invalid field keeps its
// default and is reported; the range is valid or reset as a whole.
func readReports(f fields) (ReportsPreferences, []fieldError) {
	result := defaultReports()
	var errs []fieldError

	dateFrom, fromOK := f.optionalDate("date_from", false)
	if !fromOK {
		errs = append(errs, fieldError{"date_from", msgDateFrom})
	}
	dateTo, toOK := f.optionalDate("date_to", true)
	if !toOK {
		errs = append(errs, fieldError{"date_to", msgDateTo})
	}
	if fromOK && toOK {
		if (dateFrom == nil) != (dateTo == nil) {
			errs = append(errs, fieldError{"date_range", msgDateRange})
		} else {
			result.DateFrom, result.DateTo = dateFrom, dateTo
		}
	}

	for _, flag := range []struct {
		name  string
		value *bool
	}{
		{"show_weight", &result.ShowWeight},
		{"show_calories", &result.ShowCalories},
		{"show_goal_projection", &result.ShowGoalProjection},
	} {
		if value, ok := f.bool(flag.name); ok {
			*flag.value = value
		} else {
			errs = append(errs, fieldError{flag.name, flag.name + " must be true or false"})
		}
	}

	if period, ok := f.oneOfString("period", reportPeriods); ok {
		result.Period = period
	} else {
		errs = append(errs, fieldError{"period", msgPeriod})
	}

	if step, ok := f.oneOfInt("step_interval", reportStepIntervals); ok {
		result.StepInterval = step
	} else {
		errs = append(errs, fieldError{"step_interval", msgStepInterval})
	}

	return result, errs
}
