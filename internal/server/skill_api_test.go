package server

import (
	"archive/zip"
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"ypeskov/kkal-tracker/internal/auth"
	"ypeskov/kkal-tracker/internal/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v5"
)

const skillDownloadPath = "/api/skills/food-log/download"

// syncBuffer collects log output written from several goroutines
type syncBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

type apiKeyListItem struct {
	ID        int        `json:"id"`
	Name      string     `json:"name"`
	KeyPrefix string     `json:"key_prefix"`
	ExpiresAt *time.Time `json:"expires_at"`
	IsRevoked bool       `json:"is_revoked"`
}

// listAPIKeys lists the keys of the user whose JWT the client holds
func listAPIKeys(t *testing.T, client *smokeClient) []apiKeyListItem {
	t.Helper()
	var keys []apiKeyListItem
	client.expect(http.StatusOK, http.MethodGet, "/api/api-keys", nil, &keys)
	return keys
}

// downloadSkill downloads the skill with the client's JWT and returns the response and the zip files by name
func downloadSkill(t *testing.T, client *smokeClient) (*httptest.ResponseRecorder, map[string]string) {
	t.Helper()

	rec := client.expect(http.StatusOK, http.MethodPost, skillDownloadPath, nil, nil)
	content := rec.Body.Bytes()
	reader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		t.Fatalf("the downloaded skill is not a zip: %v", err)
	}

	files := make(map[string]string, len(reader.File))
	for _, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			t.Fatalf("failed to open %s: %v", file.Name, err)
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Fatalf("failed to read %s: %v", file.Name, err)
		}
		files[file.Name] = string(data)
	}
	return rec, files
}

// skillKey checks the zip layout of the skill and returns the API key from its api_key file
func skillKey(t *testing.T, files map[string]string, skillName string) string {
	t.Helper()

	want := []string{skillName + "/SKILL.md", skillName + "/api_key", skillName + "/scripts/kkal_client.py"}
	var got []string
	for name := range files {
		got = append(got, name)
	}
	slices.Sort(got)
	if !slices.Equal(got, want) {
		t.Fatalf("zip entries = %q, want exactly %q", got, want)
	}

	content := files[skillName+"/api_key"]
	key := strings.TrimSuffix(content, "\n")
	if !regexp.MustCompile(`^[0-9a-f]{64}$`).MatchString(key) || content != key+"\n" {
		t.Fatal("api_key is not a 64-character hex key followed by a newline")
	}
	return key
}

// withKey switches the client to API key auth and returns a function that restores the JWT
func withKey(client *smokeClient, key string) (restore func()) {
	token, apiKey := client.token, client.apiKey
	client.token, client.apiKey = "", key
	return func() { client.token, client.apiKey = token, apiKey }
}

func TestSkillDownloadRequiresAuth(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "skill-auth@example.com")
	token := client.token
	keysBefore := len(listAPIKeys(t, client))

	expired, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &auth.Claims{
		UserID: 1, Email: "skill-auth@example.com",
		RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour))},
	}).SignedString([]byte(strings.Repeat("s", 32)))
	if err != nil {
		t.Fatal(err)
	}

	client.apiKey = ""
	for _, badToken := range []string{"", "invalid", expired} {
		client.token = badToken
		client.expect(http.StatusUnauthorized, http.MethodPost, skillDownloadPath, nil, nil)
	}

	// GET does not reach the handler: no key, no zip
	client.token = token
	if rec := client.do(http.MethodGet, skillDownloadPath, nil); rec.Code < 400 {
		t.Errorf("GET %s: status = %d, want an error", skillDownloadPath, rec.Code)
	}

	if keysAfter := len(listAPIKeys(t, client)); keysAfter != keysBefore {
		t.Errorf("keys after rejected downloads = %d, want %d", keysAfter, keysBefore)
	}
}

func TestSkillDownloadInstallsAWorkingSkill(t *testing.T) {
	var logs syncBuffer
	logger := slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug}))
	client, smtpServer := newSmokeEnvironmentWith(t, logger, nil)

	// User B has data of its own that A's skill key must not reach
	signUp(t, client, smtpServer, "skill-b@example.com")
	addIngredient(t, client, "Bread of B", 250)
	tokenB := client.token

	signUp(t, client, smtpServer, "skill-a@example.com")
	addIngredient(t, client, "Tomato of A", 20)
	agentKey := client.apiKey

	before := time.Now().UTC().Truncate(time.Minute)
	rec, files := downloadSkill(t, client)
	after := time.Now().UTC()

	for header, want := range map[string]string{
		echo.HeaderContentType:        "application/zip",
		echo.HeaderContentDisposition: `attachment; filename="kkal-tracker-food-log.zip"`,
		echo.HeaderCacheControl:       "no-store",
	} {
		if got := rec.Header().Get(header); got != want {
			t.Errorf("%s = %q, want %q", header, got, want)
		}
	}

	const name = "kkal-tracker-food-log"
	key := skillKey(t, files, name)

	// The skill points at APP_URL, the client keeps the placeholder
	if !strings.Contains(files[name+"/SKILL.md"], "- Base URL: `http://localhost:8080`\n") {
		t.Error("SKILL.md does not point at APP_URL")
	}
	pyClient := files[name+"/scripts/kkal_client.py"]
	if !strings.Contains(pyClient, `DEFAULT_BASE_URL = "http://localhost:8080"`+"\n") ||
		!strings.Contains(pyClient, `API_KEY = "XXXXXXXXX"`+"\n") {
		t.Error("kkal_client.py does not point at APP_URL or lost the key placeholder")
	}

	// The key works on the external API and reaches only A's data
	restore := withKey(client, key)
	ingredients := listIngredients(t, client)
	if _, ok := ingredients["Tomato of A"]; !ok {
		t.Errorf("ingredients with the skill key = %v, want A's ingredient", ingredients)
	}
	if _, ok := ingredients["Bread of B"]; ok {
		t.Error("the skill key reaches B's ingredients")
	}
	restore()

	// The key is listed for A under the download name, without expiry
	var skillKeys []apiKeyListItem
	for _, item := range listAPIKeys(t, client) {
		if item.KeyPrefix == key[:8] {
			skillKeys = append(skillKeys, item)
		}
	}
	if len(skillKeys) != 1 {
		t.Fatalf("A's keys with the prefix of the skill key = %+v, want one", skillKeys)
	}
	created, err := time.Parse("2006-01-02 15:04", strings.TrimSuffix(strings.TrimPrefix(skillKeys[0].Name, "Claude skill "), " UTC"))
	if err != nil || !strings.HasSuffix(skillKeys[0].Name, " UTC") || created.Before(before) || created.After(after) {
		t.Errorf("key name = %q, want \"Claude skill YYYY-MM-DD HH:MM UTC\" with the current UTC time", skillKeys[0].Name)
	}
	if skillKeys[0].ExpiresAt != nil || skillKeys[0].IsRevoked {
		t.Errorf("skill key = %+v, want an active key without expiry", skillKeys[0])
	}

	// B does not see A's key
	tokenA := client.token
	client.token = tokenB
	for _, item := range listAPIKeys(t, client) {
		if item.KeyPrefix == key[:8] {
			t.Error("B's key list shows A's skill key")
		}
	}
	client.token = tokenA

	// A second download creates another key; both and the earlier key stay active
	_, secondFiles := downloadSkill(t, client)
	secondKey := skillKey(t, secondFiles, name)
	if secondKey == key {
		t.Fatal("the second download returned the same key")
	}
	for _, k := range []string{key, secondKey, agentKey} {
		restore := withKey(client, k)
		client.expect(http.StatusOK, http.MethodGet, "/api/v1/ingredients", nil, nil)
		restore()
	}

	// The raw key never reaches the logs or the response headers
	for _, k := range []string{key, secondKey} {
		if strings.Contains(logs.String(), k) || strings.Contains(logs.String(), k[8:]) {
			t.Error("a skill key was logged")
		}
		for header, values := range rec.Header() {
			if strings.Contains(strings.Join(values, " "), k) {
				t.Errorf("header %s contains the key", header)
			}
		}
	}
}

// With the production and the development configuration of the real servers the skill files in the zip are
// the files of the repository byte for byte
func TestSkillDownloadMatchesTheRepository(t *testing.T) {
	tests := []struct {
		environment, appURL, skillName string
	}{
		{"production", "https://kcal.peskov.info", "kkal-tracker-food-log"},
		// A trailing slash is covered by the unit tests of internal/skill: the CORS setup rejects such an APP_URL
		{"development", "https://dev-kcal.peskov.info", "kkal-tracker-food-log-dev"},
	}
	for _, tt := range tests {
		t.Run(tt.environment, func(t *testing.T) {
			client, smtpServer := newSmokeEnvironmentWith(t, discardLogger(), func(cfg *config.Config) {
				cfg.Environment, cfg.AppURL = tt.environment, tt.appURL
			})
			signUp(t, client, smtpServer, "skill-"+tt.environment+"@example.com")

			rec, files := downloadSkill(t, client)
			if got, want := rec.Header().Get(echo.HeaderContentDisposition), `attachment; filename="`+tt.skillName+`.zip"`; got != want {
				t.Errorf("Content-Disposition = %q, want %q", got, want)
			}
			skillKey(t, files, tt.skillName)

			for _, path := range []string{"SKILL.md", "scripts/kkal_client.py"} {
				committed, err := os.ReadFile(filepath.Join("..", "..", "skills", tt.skillName, filepath.FromSlash(path)))
				if err != nil {
					t.Fatalf("failed to read the repository copy: %v", err)
				}
				if files[tt.skillName+"/"+path] != string(committed) {
					t.Errorf("%s/%s in the zip differs from the repository", tt.skillName, path)
				}
			}
		})
	}
}

func TestSkillDownloadRateLimitPerUser(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)
	signUp(t, client, smtpServer, "skill-limit-b@example.com")
	tokenB := client.token
	signUp(t, client, smtpServer, "skill-limit-a@example.com")
	keysBefore := len(listAPIKeys(t, client))

	for i := 1; i <= skillDownloadRateBurst; i++ {
		client.expect(http.StatusOK, http.MethodPost, skillDownloadPath, nil, nil)
	}
	rec := client.expect(http.StatusTooManyRequests, http.MethodPost, skillDownloadPath, nil, nil)
	if !strings.Contains(rec.Body.String(), "Too many requests, please try again later") {
		t.Errorf("429 body = %s, want the standard message", rec.Body.String())
	}
	if keys := len(listAPIKeys(t, client)); keys != keysBefore+skillDownloadRateBurst {
		t.Errorf("keys = %d, want %d: a limited request must not create a key", keys, keysBefore+skillDownloadRateBurst)
	}

	// Same client IP, another user: not affected
	client.token = tokenB
	client.expect(http.StatusOK, http.MethodPost, skillDownloadPath, nil, nil)
}
