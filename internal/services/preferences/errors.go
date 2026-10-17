package preferences

import "errors"

var (
	// ErrUnknownKey means the key is not in the allow-list of preferences
	ErrUnknownKey = errors.New("unknown preference")
	// ErrInvalidBody means the request body is not a JSON object
	ErrInvalidBody = errors.New("invalid request body")
)

// ValidationError is an invalid or missing field of a preference value. The message names the field and is safe to
// show to the client.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}
