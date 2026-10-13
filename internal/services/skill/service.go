package skill

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"path"
	"time"

	"ypeskov/kkal-tracker/internal/config"
	apikeyservice "ypeskov/kkal-tracker/internal/services/apikey"
	skillfiles "ypeskov/kkal-tracker/internal/skill"
)

// zipEntry is a file of the archive; name is the full path inside the archive
type zipEntry struct {
	name    string
	content []byte
	mode    fs.FileMode
}

type Service struct {
	apiKeyService apikeyservice.Servicer
	development   bool
	baseURL       string
	source        skillfiles.Source
	now           func() time.Time
	writeZip      func(w io.Writer, entries []zipEntry, modified time.Time) error
	logger        *slog.Logger
}

func New(cfg *config.Config, apiKeyService apikeyservice.Servicer, logger *slog.Logger) *Service {
	return &Service{
		apiKeyService: apiKeyService,
		development:   cfg.IsDevelopment(),
		baseURL:       cfg.AppURL,
		source:        skillfiles.Embedded(),
		now:           time.Now,
		writeZip:      writeZip,
		logger:        logger.With("service", "skill"),
	}
}

// DownloadFoodLog creates a new API key for the user and returns the food logging skill of this server with the key
// in its client script. The development server gives the "-dev" variant. The skill is rendered before the key exists,
// so a broken template creates no key; if anything fails after the key was created, the key is deleted.
// The raw key appears only in the returned content, never in errors or logs.
func (s *Service) DownloadFoodLog(userID int) (*Download, error) {
	rendered, err := skillfiles.Render(s.source, s.development, s.baseURL)
	if err != nil {
		return nil, fmt.Errorf("render skill: %w", err)
	}

	now := s.now().UTC()
	apiKey, rawKey, err := s.apiKeyService.CreateKey(userID, KeyName(now), nil)
	if err != nil {
		return nil, fmt.Errorf("create API key: %w", err)
	}

	withKey, err := rendered.WithAPIKey(rawKey)
	if err != nil {
		s.deleteKey(apiKey.ID, userID)
		return nil, fmt.Errorf("insert API key: %w", err)
	}

	entries := make([]zipEntry, 0, len(withKey.Files))
	for _, file := range withKey.Files {
		mode := fs.FileMode(0o644)
		if file.Path == skillfiles.ClientFile {
			// The client holds the key: readable by the owner only once unpacked
			mode = 0o600
		}
		entries = append(entries, zipEntry{name: path.Join(withKey.Name, file.Path), content: file.Content, mode: mode})
	}

	var archive bytes.Buffer
	if err := s.writeZip(&archive, entries, now); err != nil {
		s.deleteKey(apiKey.ID, userID)
		return nil, fmt.Errorf("build zip: %w", err)
	}

	s.logger.Info("Skill prepared for download", "user_id", userID, "skill", withKey.Name, "api_key_id", apiKey.ID)
	return &Download{FileName: withKey.Name + ".zip", Content: archive.Bytes()}, nil
}

// KeyName is the name of the API key created for a skill download
func KeyName(t time.Time) string {
	return "Claude skill " + t.UTC().Format("2006-01-02 15:04") + " UTC"
}

// deleteKey removes a key that never reached the user; a failure is only logged
func (s *Service) deleteKey(keyID, userID int) {
	if err := s.apiKeyService.DeleteKey(keyID, userID); err != nil {
		s.logger.Error("Failed to delete the API key of a failed skill download",
			"error", err, "user_id", userID, "api_key_id", keyID)
		return
	}
	s.logger.Warn("Deleted the API key of a failed skill download", "user_id", userID, "api_key_id", keyID)
}

// writeZip packs the entries in the given order, without directory entries
func writeZip(w io.Writer, entries []zipEntry, modified time.Time) error {
	zw := zip.NewWriter(w)
	for _, entry := range entries {
		header := &zip.FileHeader{Name: entry.name, Method: zip.Deflate, Modified: modified}
		header.SetMode(entry.mode)

		fw, err := zw.CreateHeader(header)
		if err != nil {
			return err
		}
		if _, err := fw.Write(entry.content); err != nil {
			return err
		}
	}
	return zw.Close()
}
