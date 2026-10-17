package preferences

import (
	"bytes"
	"encoding/json"
	"slices"
	"time"
)

// Accepted calendar dates: anything else is a typo or an attack, not a report range
const (
	minDate = "1900-01-01"
	maxDate = "2100-12-31"
)

// fields is a decoded JSON object: field name → raw value. Names match exactly; fields nobody reads are ignored.
type fields map[string]json.RawMessage

// fieldError is a field that is missing or invalid; message names the field and is safe to show to the client
type fieldError struct {
	field   string
	message string
}

// decodeObject decodes a JSON object; anything else (malformed JSON, null, an array, a scalar) is an error
func decodeObject(data []byte) (fields, error) {
	var object fields
	if err := json.Unmarshal(data, &object); err != nil || object == nil {
		return nil, ErrInvalidBody
	}
	return object, nil
}

// isNull reports whether a raw value is the JSON null
func isNull(raw json.RawMessage) bool {
	return bytes.Equal(bytes.TrimSpace(raw), []byte("null"))
}

// bool reads a required boolean
func (f fields) bool(name string) (bool, bool) {
	raw, present := f[name]
	if !present || isNull(raw) {
		return false, false
	}
	var value bool
	if err := json.Unmarshal(raw, &value); err != nil {
		return false, false
	}
	return value, true
}

// oneOfString reads a required string that must be one of allowed
func (f fields) oneOfString(name string, allowed []string) (string, bool) {
	raw, present := f[name]
	if !present || isNull(raw) {
		return "", false
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil || !slices.Contains(allowed, value) {
		return "", false
	}
	return value, true
}

// oneOfInt reads a required integer that must be one of allowed; strings and fractions are rejected
func (f fields) oneOfInt(name string, allowed []int) (int, bool) {
	raw, present := f[name]
	if !present || isNull(raw) {
		return 0, false
	}
	var value int
	if err := json.Unmarshal(raw, &value); err != nil || !slices.Contains(allowed, value) {
		return 0, false
	}
	return value, true
}

// optionalDate reads a date (YYYY-MM-DD) within minDate..maxDate, or null; a missing field counts as null.
// With allowToday, the literal DateToToday is accepted too.
func (f fields) optionalDate(name string, allowToday bool) (*string, bool) {
	raw, present := f[name]
	if !present || isNull(raw) {
		return nil, true
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return nil, false
	}
	if allowToday && value == DateToToday {
		return &value, true
	}
	if !isCalendarDate(value) {
		return nil, false
	}
	return &value, true
}

// isCalendarDate reports whether value is an existing calendar day in YYYY-MM-DD form within minDate..maxDate
func isCalendarDate(value string) bool {
	if _, err := time.Parse(time.DateOnly, value); err != nil {
		return false
	}
	// Both sides are YYYY-MM-DD here, so they compare as strings
	return value >= minDate && value <= maxDate
}
