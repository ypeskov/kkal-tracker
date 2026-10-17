package preferences

// Servicer defines the preferences service contract used by handlers.
type Servicer interface {
	// Get returns the value of the key for the user: the stored one, with defaults for what is missing or invalid
	Get(userID int, key string) (any, error)
	// Save validates a JSON request body, stores its normalized value in place of the previous one and returns it
	Save(userID int, key string, body []byte) (any, error)
}
