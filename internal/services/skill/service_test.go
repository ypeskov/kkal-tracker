package skill

import (
	"archive/zip"
	"bytes"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"ypeskov/kkal-tracker/internal/config"
	"ypeskov/kkal-tracker/internal/models"
	apikeyservice "ypeskov/kkal-tracker/internal/services/apikey"
	skillfiles "ypeskov/kkal-tracker/internal/skill"
)

const testRawKey = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

var testNow = time.Date(2026, 10, 12, 15, 4, 59, 0, time.FixedZone("EEST", 3*60*60))

// fakeAPIKeyService keeps keys in memory. Methods the service does not use come from the embedded nil interface.
type fakeAPIKeyService struct {
	apikeyservice.Servicer

	keys      map[int]*models.APIKey
	nextID    int
	createErr error
	deleteErr error
}

func newFakeAPIKeyService() *fakeAPIKeyService {
	return &fakeAPIKeyService{keys: make(map[int]*models.APIKey), nextID: 1}
}

func (f *fakeAPIKeyService) CreateKey(userID int, name string, expiryDays *int) (*models.APIKey, string, error) {
	if f.createErr != nil {
		return nil, "", f.createErr
	}
	if expiryDays != nil {
		return nil, "", errors.New("skill keys must not expire")
	}
	key := &models.APIKey{ID: f.nextID, UserID: userID, Name: name, KeyPrefix: testRawKey[:8]}
	f.keys[key.ID] = key
	f.nextID++
	return key, testRawKey, nil
}

func (f *fakeAPIKeyService) DeleteKey(id, userID int) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	if key, ok := f.keys[id]; ok && key.UserID == userID {
		delete(f.keys, id)
	}
	return nil
}

func newTestService(cfg *config.Config, keys *fakeAPIKeyService, logs io.Writer) *Service {
	logger := slog.New(slog.NewTextHandler(logs, &slog.HandlerOptions{Level: slog.LevelDebug}))
	s := New(cfg, keys, logger)
	s.now = func() time.Time { return testNow }
	return s
}

func unzip(t *testing.T, content []byte) map[string]*zip.File {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		t.Fatalf("download is not a zip: %v", err)
	}
	files := make(map[string]*zip.File)
	for _, file := range reader.File {
		files[file.Name] = file
	}
	return files
}

func readZipFile(t *testing.T, file *zip.File) string {
	t.Helper()
	rc, err := file.Open()
	if err != nil {
		t.Fatalf("failed to open %s: %v", file.Name, err)
	}
	defer rc.Close()
	content, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("failed to read %s: %v", file.Name, err)
	}
	return string(content)
}

func TestDownloadFoodLog(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		appURL      string
		wantName    string
	}{
		{"production", "production", "https://kcal.example.test", skillfiles.ProductionName},
		{"development", "development", "https://dev.example.test/", skillfiles.DevelopmentName},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys := newFakeAPIKeyService()
			var logs bytes.Buffer
			s := newTestService(&config.Config{Environment: tt.environment, AppURL: tt.appURL}, keys, &logs)

			download, err := s.DownloadFoodLog(7)
			if err != nil {
				t.Fatalf("DownloadFoodLog: %v", err)
			}
			if download.FileName != tt.wantName+".zip" {
				t.Errorf("file name = %q, want %q", download.FileName, tt.wantName+".zip")
			}

			files := unzip(t, download.Content)
			if len(files) != 3 {
				t.Errorf("zip entries = %d, want 3", len(files))
			}
			want, err := skillfiles.Render(skillfiles.Embedded(), tt.environment == "development", tt.appURL)
			if err != nil {
				t.Fatal(err)
			}
			for _, file := range want.Files {
				entry := files[tt.wantName+"/"+file.Path]
				if entry == nil {
					t.Fatalf("zip has no %s/%s", tt.wantName, file.Path)
				}
				if readZipFile(t, entry) != string(file.Content) {
					t.Errorf("%s differs from the rendered skill", entry.Name)
				}
				if entry.Mode().Perm() != 0o644 {
					t.Errorf("%s mode = %v, want 0644", entry.Name, entry.Mode())
				}
			}
			keyEntry := files[tt.wantName+"/api_key"]
			if keyEntry == nil {
				t.Fatalf("zip has no %s/api_key", tt.wantName)
			}
			if readZipFile(t, keyEntry) != testRawKey+"\n" {
				t.Error("api_key does not hold the new key plus a newline")
			}
			if keyEntry.Mode().Perm() != 0o600 {
				t.Errorf("api_key mode = %v, want 0600", keyEntry.Mode())
			}

			if len(keys.keys) != 1 || keys.keys[1].UserID != 7 || keys.keys[1].Name != "Claude skill 2026-10-12 12:04 UTC" {
				t.Errorf("created keys = %+v, want one key of user 7 named by the UTC time", keys.keys)
			}
			if strings.Contains(logs.String(), testRawKey) || strings.Contains(logs.String(), testRawKey[8:]) {
				t.Errorf("the raw key was logged: %s", logs.String())
			}
		})
	}
}

func TestDownloadFoodLogDeletesTheKeyWhenZippingFails(t *testing.T) {
	keys := newFakeAPIKeyService()
	var logs bytes.Buffer
	s := newTestService(&config.Config{AppURL: "https://kcal.example.test"}, keys, &logs)
	s.writeZip = func(io.Writer, []zipEntry, time.Time) error { return errors.New("disk full") }

	download, err := s.DownloadFoodLog(7)
	if err == nil || download != nil {
		t.Fatalf("DownloadFoodLog = %v, %v; want an error", download, err)
	}
	if strings.Contains(err.Error(), testRawKey) {
		t.Error("the error contains the raw key")
	}
	if len(keys.keys) != 0 {
		t.Errorf("keys left after a failed download: %+v", keys.keys)
	}
	if strings.Contains(logs.String(), testRawKey) {
		t.Errorf("the raw key was logged: %s", logs.String())
	}
}

func TestDownloadFoodLogLogsAFailedKeyCleanup(t *testing.T) {
	keys := newFakeAPIKeyService()
	keys.deleteErr = errors.New("database is locked")
	var logs bytes.Buffer
	s := newTestService(&config.Config{AppURL: "https://kcal.example.test"}, keys, &logs)
	s.writeZip = func(io.Writer, []zipEntry, time.Time) error { return errors.New("disk full") }

	if _, err := s.DownloadFoodLog(7); err == nil || !strings.Contains(err.Error(), "disk full") {
		t.Fatalf("error = %v, want the zip error", err)
	}
	for _, part := range []string{"level=ERROR", "database is locked", "user_id=7", "api_key_id=1"} {
		if !strings.Contains(logs.String(), part) {
			t.Errorf("log %q does not contain %q", logs.String(), part)
		}
	}
}

func TestDownloadFoodLogFailsWithoutCreatingAKey(t *testing.T) {
	tests := []struct {
		name   string
		appURL string
		source skillfiles.Source
	}{
		{"invalid base URL", "localhost:8080", skillfiles.Embedded()},
		{"broken source", "https://kcal.example.test", skillfiles.Source{Skill: "# no frontmatter\n"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			keys := newFakeAPIKeyService()
			s := newTestService(&config.Config{AppURL: tt.appURL}, keys, io.Discard)
			s.source = tt.source

			if download, err := s.DownloadFoodLog(7); err == nil || download != nil {
				t.Fatalf("DownloadFoodLog = %v, %v; want an error", download, err)
			}
			if len(keys.keys) != 0 {
				t.Errorf("a key was created: %+v", keys.keys)
			}
		})
	}
}

func TestDownloadFoodLogKeyCreationFailure(t *testing.T) {
	keys := newFakeAPIKeyService()
	keys.createErr = errors.New("database is locked")
	s := newTestService(&config.Config{AppURL: "https://kcal.example.test"}, keys, io.Discard)

	download, err := s.DownloadFoodLog(7)
	if !errors.Is(err, keys.createErr) || download != nil {
		t.Fatalf("DownloadFoodLog = %v, %v; want the key creation error", download, err)
	}
}

func TestKeyName(t *testing.T) {
	if got := KeyName(testNow); got != "Claude skill 2026-10-12 12:04 UTC" {
		t.Errorf("KeyName = %q", got)
	}
}
