package preferences

import (
	"encoding/json"
	"errors"
	"log/slog"
	"strconv"

	"ypeskov/kkal-tracker/internal/repositories"
)

// versionField holds the schema version in a stored value; it is not part of the API object
const versionField = "v"

// definition describes an allowed preference key: its defaults and how its fields are read and validated
type definition struct {
	// version is the schema version written with every stored value; a stored value of another version is reset
	version int
	// defaults returns the value of a user who has stored nothing
	defaults func() any
	// read reads the fields of a request body or a stored value. Invalid fields keep their defaults and are reported.
	read func(f fields) (any, []fieldError)
}

// definitions is the allow-list of preference keys
var definitions = map[string]definition{
	KeyReports: reportsDefinition,
}

type Service struct {
	repo   repositories.UserPreferenceRepository
	logger *slog.Logger
}

func New(repo repositories.UserPreferenceRepository, logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger.With("service", "preferences"),
	}
}

// Get returns the value of the key for the user. Nothing stored gives the defaults; a stored value that cannot be
// decoded gives the defaults, and its invalid fields give their defaults, with a warning (never with the value).
func (s *Service) Get(userID int, key string) (any, error) {
	def, ok := definitions[key]
	if !ok {
		return nil, ErrUnknownKey
	}

	stored, err := s.repo.Get(userID, key)
	if errors.Is(err, repositories.ErrNotFound) {
		return def.defaults(), nil
	}
	if err != nil {
		return nil, err
	}

	object, err := decodeObject([]byte(stored.Value))
	if err != nil || !hasVersion(object, def.version) {
		s.logger.Warn("Stored preference cannot be decoded, defaults are used", "user_id", userID, "key", key)
		return def.defaults(), nil
	}

	value, errs := def.read(object)
	if len(errs) > 0 {
		names := make([]string, len(errs))
		for i, e := range errs {
			names[i] = e.field
		}
		s.logger.Warn("Stored preference has invalid fields, their defaults are used",
			"user_id", userID, "key", key, "fields", names)
	}
	return value, nil
}

// Save validates the request body of the key and stores the server's own encoding of the normalized value, replacing
// the previous one. Unknown fields are ignored.
func (s *Service) Save(userID int, key string, body []byte) (any, error) {
	def, ok := definitions[key]
	if !ok {
		return nil, ErrUnknownKey
	}

	object, err := decodeObject(body)
	if err != nil {
		return nil, err
	}
	value, errs := def.read(object)
	if len(errs) > 0 {
		return nil, &ValidationError{Message: errs[0].message}
	}

	encoded, err := encodeStored(def.version, value)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Upsert(userID, key, encoded); err != nil {
		return nil, err
	}

	s.logger.Debug("Preference saved", "user_id", userID, "key", key)
	return value, nil
}

// encodeStored encodes a value with its schema version: {"v": <version>, ...fields of the value}
func encodeStored(version int, value any) (string, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	object, err := decodeObject(data)
	if err != nil {
		return "", err
	}
	object[versionField] = json.RawMessage(strconv.Itoa(version))

	encoded, err := json.Marshal(object)
	if err != nil {
		return "", err
	}
	return string(encoded), nil
}

// hasVersion reports whether a stored object carries the given schema version
func hasVersion(object fields, version int) bool {
	var stored int
	raw, present := object[versionField]
	return present && json.Unmarshal(raw, &stored) == nil && stored == version
}
