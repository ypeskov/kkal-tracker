package server

import (
	"bufio"
	"bytes"
	"embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"net/http/httptest"
	"net/mail"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"

	"ypeskov/kkal-tracker/internal/config"
	"ypeskov/kkal-tracker/internal/database"

	"github.com/pressly/goose/v3"
)

// fakeSMTPServer accepts messages over SMTP on a local port and keeps them in memory
type fakeSMTPServer struct {
	listener net.Listener
	mu       sync.Mutex
	messages [][]byte
}

func newFakeSMTPServer(t *testing.T) *fakeSMTPServer {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start fake SMTP server: %v", err)
	}
	s := &fakeSMTPServer{listener: listener}
	t.Cleanup(func() { listener.Close() })

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go s.handle(conn)
		}
	}()

	return s
}

func (s *fakeSMTPServer) port() int {
	return s.listener.Addr().(*net.TCPAddr).Port
}

func (s *fakeSMTPServer) handle(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	reply := func(line string) { fmt.Fprintf(conn, "%s\r\n", line) }

	reply("220 localhost ESMTP")
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		command := strings.ToUpper(strings.TrimSpace(line))

		switch {
		case strings.HasPrefix(command, "EHLO"), strings.HasPrefix(command, "HELO"):
			reply("250-localhost")
			reply("250 AUTH PLAIN")
		case strings.HasPrefix(command, "AUTH"):
			reply("235 Authentication successful")
		case strings.HasPrefix(command, "DATA"):
			reply("354 End data with <CR><LF>.<CR><LF>")
			var data bytes.Buffer
			for {
				dataLine, err := reader.ReadString('\n')
				if err != nil {
					return
				}
				if dataLine == ".\r\n" {
					break
				}
				// Undo SMTP dot-stuffing
				data.WriteString(strings.TrimPrefix(dataLine, "."))
			}
			s.mu.Lock()
			s.messages = append(s.messages, data.Bytes())
			s.mu.Unlock()
			reply("250 OK")
		case strings.HasPrefix(command, "QUIT"):
			reply("221 Bye")
			return
		default:
			reply("250 OK")
		}
	}
}

// lastMessage returns the most recently received message
func (s *fakeSMTPServer) lastMessage(t *testing.T) []byte {
	t.Helper()

	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.messages) == 0 {
		t.Fatal("no email was sent")
	}
	return s.messages[len(s.messages)-1]
}

func (s *fakeSMTPServer) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.messages)
}

// smokeClient sends JSON requests to the fully wired server
type smokeClient struct {
	t       *testing.T
	handler http.Handler
	token   string
	apiKey  string
}

func (c *smokeClient) do(method, path string, body any) *httptest.ResponseRecorder {
	c.t.Helper()

	var payload io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			c.t.Fatalf("failed to encode request body: %v", err)
		}
		payload = bytes.NewReader(encoded)
	}

	req := httptest.NewRequest(method, path, payload)
	req.RemoteAddr = "203.0.113.7:1234"
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}

	rec := httptest.NewRecorder()
	c.handler.ServeHTTP(rec, req)
	return rec
}

// expect sends a request, checks the status code and decodes the JSON response into out (if not nil)
func (c *smokeClient) expect(wantStatus int, method, path string, body, out any) *httptest.ResponseRecorder {
	c.t.Helper()

	rec := c.do(method, path, body)
	if rec.Code != wantStatus {
		c.t.Fatalf("%s %s: status = %d, want %d; body: %s", method, path, rec.Code, wantStatus, rec.Body.String())
	}
	if out != nil {
		if err := json.Unmarshal(rec.Body.Bytes(), out); err != nil {
			c.t.Fatalf("%s %s: response is not valid JSON: %v; body: %s", method, path, err, rec.Body.String())
		}
	}
	return rec
}

// newSmokeEnvironment starts the server on a fresh migrated SQLite database with a fake SMTP server
func newSmokeEnvironment(t *testing.T) (*smokeClient, *fakeSMTPServer) {
	t.Helper()

	logger := discardLogger()

	db, err := database.New(filepath.Join(t.TempDir(), "app.db"), logger)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	goose.SetLogger(goose.NopLogger())
	if err := goose.SetDialect("sqlite3"); err != nil {
		t.Fatalf("failed to set migration dialect: %v", err)
	}
	if err := goose.Up(db, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	smtpServer := newFakeSMTPServer(t)

	cfg := &config.Config{
		Port:         "0",
		JWTSecret:    strings.Repeat("s", 32),
		AppURL:       "http://localhost:8080",
		SMTPHost:     "127.0.0.1",
		SMTPPort:     smtpServer.port(),
		SMTPUser:     "user",
		SMTPPassword: "password",
		SMTPFrom:     "noreply@example.com",
	}
	srv := New(cfg, logger, db, embed.FS{}).Start()

	return &smokeClient{t: t, handler: srv.Handler}, smtpServer
}

func TestSmokeUserJourney(t *testing.T) {
	client, smtpServer := newSmokeEnvironment(t)

	const email = "smoke@example.com"
	const password = "smoke-password"

	// --- Auth: register, activate by the emailed link, log in
	client.expect(http.StatusUnauthorized, http.MethodGet, "/api/calories?dateFrom=2026-01-01&dateTo=2026-01-31", nil, nil)

	client.expect(http.StatusCreated, http.MethodPost, "/api/auth/register",
		map[string]string{"email": email, "password": password, "language_code": "en_US"}, nil)

	client.expect(http.StatusForbidden, http.MethodPost, "/api/auth/login",
		map[string]string{"email": email, "password": password}, nil)

	activationEmail := string(smtpServer.lastMessage(t))
	if !strings.Contains(activationEmail, "To: "+email) {
		t.Fatalf("activation email is not addressed to %s", email)
	}
	match := regexp.MustCompile(`/activate/([A-Za-z0-9_-]+)`).FindStringSubmatch(activationEmail)
	if match == nil {
		t.Fatal("activation email does not contain an activation link")
	}
	client.expect(http.StatusOK, http.MethodGet, "/api/auth/activate/"+match[1], nil, nil)

	client.expect(http.StatusUnauthorized, http.MethodPost, "/api/auth/login",
		map[string]string{"email": email, "password": "wrong-password"}, nil)

	var login struct {
		Token string `json:"token"`
	}
	client.expect(http.StatusOK, http.MethodPost, "/api/auth/login",
		map[string]string{"email": email, "password": password}, &login)
	if login.Token == "" {
		t.Fatal("login did not return a token")
	}
	client.token = login.Token

	var me struct {
		Email string `json:"email"`
	}
	client.expect(http.StatusOK, http.MethodGet, "/api/auth/me", nil, &me)
	if me.Email != email {
		t.Errorf("/api/auth/me email = %q, want %q", me.Email, email)
	}

	// --- Calories: create, list, update, delete
	type calorieEntry struct {
		ID       int    `json:"id"`
		Food     string `json:"food"`
		Calories int    `json:"calories"`
	}
	entryRequest := map[string]any{
		"food": "Oatmeal", "calories": 150, "weight": 100.0, "kcalPer100g": 150.0,
		"meal_datetime": "2026-01-15T08:30:00Z",
	}
	var createResult struct {
		Entry calorieEntry `json:"entry"`
	}
	client.expect(http.StatusCreated, http.MethodPost, "/api/calories", entryRequest, &createResult)
	created := createResult.Entry
	if created.ID == 0 || created.Food != "Oatmeal" {
		t.Fatalf("created calorie entry = %+v", created)
	}

	const caloriesRange = "/api/calories?dateFrom=2026-01-01&dateTo=2026-01-31"
	var entries []calorieEntry
	client.expect(http.StatusOK, http.MethodGet, caloriesRange, nil, &entries)
	if len(entries) != 1 || entries[0].ID != created.ID {
		t.Fatalf("calorie entries = %+v, want the created entry", entries)
	}

	entryRequest["food"] = "Oatmeal with honey"
	entryRequest["calories"] = 210
	var updated calorieEntry
	client.expect(http.StatusOK, http.MethodPut, fmt.Sprintf("/api/calories/%d", created.ID), entryRequest, &updated)
	if updated.Food != "Oatmeal with honey" || updated.Calories != 210 {
		t.Errorf("updated calorie entry = %+v", updated)
	}

	client.expect(http.StatusBadRequest, http.MethodPost, "/api/calories", map[string]any{"food": "No calories"}, nil)

	// A handler-level 404 must reach the client as JSON, not as the SPA index page
	notFound := client.expect(http.StatusNotFound, http.MethodGet, "/api/ingredients/999999", nil, nil)
	if contentType := notFound.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "application/json") {
		t.Errorf("API 404 Content-Type = %q, want application/json", contentType)
	}

	// --- Weight: create, list, update
	type weightEntry struct {
		ID     int     `json:"id"`
		Weight float64 `json:"weight"`
	}
	var weight weightEntry
	client.expect(http.StatusCreated, http.MethodPost, "/api/weight",
		map[string]any{"weight": 80.5, "recorded_at": "2026-01-15"}, &weight)
	if weight.ID == 0 || weight.Weight != 80.5 {
		t.Fatalf("created weight entry = %+v", weight)
	}

	const weightRange = "/api/weight?from=2026-01-01&to=2026-01-31"
	var weights []weightEntry
	client.expect(http.StatusOK, http.MethodGet, weightRange, nil, &weights)
	if len(weights) != 1 || weights[0].ID != weight.ID {
		t.Fatalf("weight entries = %+v, want the created entry", weights)
	}

	var updatedWeight weightEntry
	client.expect(http.StatusOK, http.MethodPut, fmt.Sprintf("/api/weight/%d", weight.ID),
		map[string]any{"weight": 79.9, "recorded_at": "2026-01-16"}, &updatedWeight)
	if updatedWeight.Weight != 79.9 {
		t.Errorf("updated weight = %v, want 79.9", updatedWeight.Weight)
	}

	// --- Export: download and email delivery
	exportRequest := map[string]string{
		"date_from": "2026-01-01", "date_to": "2026-01-31", "data_type": "both", "delivery_type": "download",
	}
	download := client.expect(http.StatusOK, http.MethodPost, "/api/export", exportRequest, nil)
	if !bytes.HasPrefix(download.Body.Bytes(), []byte("PK")) {
		t.Error("downloaded export is not an XLSX (zip) file")
	}
	if contentType := download.Header().Get("Content-Type"); !strings.Contains(contentType, "spreadsheetml") {
		t.Errorf("download Content-Type = %q", contentType)
	}

	emailsBefore := smtpServer.count()
	exportRequest["delivery_type"] = "email"
	client.expect(http.StatusOK, http.MethodPost, "/api/export", exportRequest, nil)
	if smtpServer.count() != emailsBefore+1 {
		t.Fatalf("export by email sent %d emails, want 1", smtpServer.count()-emailsBefore)
	}
	attachment := exportAttachment(t, smtpServer.lastMessage(t))
	if !bytes.HasPrefix(attachment, []byte("PK")) {
		t.Error("emailed export attachment is not an XLSX (zip) file")
	}

	// --- Cleanup through the API: delete what was created
	client.expect(http.StatusNoContent, http.MethodDelete, fmt.Sprintf("/api/calories/%d", created.ID), nil, nil)
	client.expect(http.StatusOK, http.MethodGet, caloriesRange, nil, &entries)
	if len(entries) != 0 {
		t.Errorf("calorie entries after delete = %+v, want none", entries)
	}

	client.expect(http.StatusNoContent, http.MethodDelete, fmt.Sprintf("/api/weight/%d", weight.ID), nil, nil)
	client.expect(http.StatusOK, http.MethodGet, weightRange, nil, &weights)
	if len(weights) != 0 {
		t.Errorf("weight entries after delete = %+v, want none", weights)
	}
}

// exportAttachment parses an export email and returns the decoded attachment
func exportAttachment(t *testing.T, message []byte) []byte {
	t.Helper()

	parsed, err := mail.ReadMessage(bytes.NewReader(message))
	if err != nil {
		t.Fatalf("export email is not parseable: %v", err)
	}
	_, params, err := mime.ParseMediaType(parsed.Header.Get("Content-Type"))
	if err != nil {
		t.Fatalf("export email has an invalid Content-Type: %v", err)
	}

	reader := multipart.NewReader(parsed.Body, params["boundary"])
	for {
		part, err := reader.NextPart()
		if err != nil {
			t.Fatalf("export email has no attachment: %v", err)
		}
		if part.FileName() == "" {
			continue
		}
		encoded, err := io.ReadAll(part)
		if err != nil {
			t.Fatalf("failed to read the attachment: %v", err)
		}
		decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(string(encoded), "\r\n", ""))
		if err != nil {
			t.Fatalf("attachment is not valid base64: %v", err)
		}
		return decoded
	}
}
