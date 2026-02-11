package apikey

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"time"

	"ypeskov/kkal-tracker/internal/models"
	"ypeskov/kkal-tracker/internal/repositories"
)

var (
	ErrKeyExpired = errors.New("API key has expired")
	ErrKeyRevoked = errors.New("API key has been revoked")
	ErrKeyInvalid = errors.New("invalid API key")
)

type Service struct {
	apiKeyRepo repositories.APIKeyRepository
	logger     *slog.Logger
}

func New(apiKeyRepo repositories.APIKeyRepository, logger *slog.Logger) *Service {
	return &Service{
		apiKeyRepo: apiKeyRepo,
		logger:     logger.With("service", "apikey"),
	}
}

// CreateKey generates a new API key for the user.
// Returns the model and the raw key string (shown only once).
func (s *Service) CreateKey(userID int, name string, expiryDays *int) (*models.APIKey, string, error) {
	s.logger.Debug("Creating API key", "user_id", userID, "name", name)

	rawKey, err := generateRawKey()
	if err != nil {
		s.logger.Error("Failed to generate API key", "error", err)
		return nil, "", err
	}

	keyHash := hashKey(rawKey)
	keyPrefix := rawKey[:8]

	var expiresAt *time.Time
	if expiryDays != nil {
		t := time.Now().AddDate(0, 0, *expiryDays)
		expiresAt = &t
	}

	apiKey, err := s.apiKeyRepo.Create(userID, name, keyHash, keyPrefix, expiresAt)
	if err != nil {
		s.logger.Error("Failed to create API key", "error", err, "user_id", userID)
		return nil, "", err
	}

	s.logger.Info("API key created", "user_id", userID, "prefix", keyPrefix+"...")
	return apiKey, rawKey, nil
}

// ValidateKey checks a raw API key and returns the associated model if valid
func (s *Service) ValidateKey(rawKey string) (*models.APIKey, error) {
	keyHash := hashKey(rawKey)

	apiKey, err := s.apiKeyRepo.GetByKeyHash(keyHash)
	if err != nil {
		if errors.Is(err, repositories.ErrNotFound) {
			return nil, ErrKeyInvalid
		}
		s.logger.Error("Failed to look up API key", "error", err)
		return nil, err
	}

	if apiKey.IsRevoked {
		return nil, ErrKeyRevoked
	}

	if apiKey.IsExpired() {
		return nil, ErrKeyExpired
	}

	return apiKey, nil
}

// GetUserKeys returns all API keys for a user
func (s *Service) GetUserKeys(userID int) ([]*models.APIKey, error) {
	return s.apiKeyRepo.GetByUserID(userID)
}

// RevokeKey revokes an API key
func (s *Service) RevokeKey(id, userID int) error {
	return s.apiKeyRepo.Revoke(id, userID)
}

// DeleteKey deletes an API key
func (s *Service) DeleteKey(id, userID int) error {
	return s.apiKeyRepo.Delete(id, userID)
}

// generateRawKey creates a cryptographically secure random key (64 hex chars)
func generateRawKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// hashKey computes SHA-256 hash of a raw key
func hashKey(rawKey string) string {
	h := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(h[:])
}
