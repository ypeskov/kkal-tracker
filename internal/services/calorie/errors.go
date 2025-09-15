package calorie

import "errors"

var (
	ErrInvalidDate   = errors.New("invalid date format")
	ErrEntryNotFound = errors.New("calorie entry not found")
)