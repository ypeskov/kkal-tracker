package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"ypeskov/kkal-tracker/internal/middleware"
	"ypeskov/kkal-tracker/internal/models"
	authservice "ypeskov/kkal-tracker/internal/services/auth"

	"github.com/labstack/echo/v5"
)

// fakeAuthService records whether the handler reached the service layer
type fakeAuthService struct {
	loginCalls    int
	registerCalls int
	loginErr      error
}

func (f *fakeAuthService) Login(email, password string) (*models.User, string, error) {
	f.loginCalls++
	if f.loginErr != nil {
		return nil, "", f.loginErr
	}
	return &models.User{Email: email}, "token", nil
}

func (f *fakeAuthService) Register(email, password, languageCode string, skipActivation bool) (*models.User, string, error) {
	f.registerCalls++
	return &models.User{Email: email}, "", nil
}

func (f *fakeAuthService) GetCurrentUser(userID int) (*models.User, error) {
	return &models.User{}, nil
}

func (f *fakeAuthService) ActivateUser(token string) error {
	return nil
}

func performRequest(t *testing.T, handle func(h *Handler) echo.HandlerFunc, body string) (int, *fakeAuthService) {
	t.Helper()

	e := echo.New()
	e.Validator = middleware.NewValidator()

	service := &fakeAuthService{}
	h := NewHandler(service, slog.New(slog.NewTextHandler(io.Discard, nil)))

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	err := handle(h)(e.NewContext(req, rec))
	if err == nil {
		return rec.Code, service
	}
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("unexpected error type %T: %v", err, err)
	}
	return httpErr.Code, service
}

func TestLoginValidation(t *testing.T) {
	login := func(h *Handler) echo.HandlerFunc { return h.Login }

	tests := []struct {
		name      string
		body      string
		wantCode  int
		wantCalls int
	}{
		{"valid request", `{"email":"user@example.com","password":"secret"}`, http.StatusOK, 1},
		{"missing email", `{"password":"secret"}`, http.StatusBadRequest, 0},
		{"malformed email", `{"email":"not-an-email","password":"secret"}`, http.StatusBadRequest, 0},
		{"missing password", `{"email":"user@example.com"}`, http.StatusBadRequest, 0},
		{"empty body", `{}`, http.StatusBadRequest, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, service := performRequest(t, login, tt.body)
			if code != tt.wantCode {
				t.Errorf("status = %d, want %d", code, tt.wantCode)
			}
			if service.loginCalls != tt.wantCalls {
				t.Errorf("service.Login calls = %d, want %d", service.loginCalls, tt.wantCalls)
			}
		})
	}
}

func TestRegisterValidation(t *testing.T) {
	register := func(h *Handler) echo.HandlerFunc { return h.Register }

	tests := []struct {
		name      string
		body      string
		wantCode  int
		wantCalls int
	}{
		{"valid request", `{"email":"user@example.com","password":"secret1","language_code":"en_US"}`, http.StatusCreated, 1},
		{"malformed email", `{"email":"not-an-email","password":"secret1","language_code":"en_US"}`, http.StatusBadRequest, 0},
		{"empty password", `{"email":"user@example.com","password":"","language_code":"en_US"}`, http.StatusBadRequest, 0},
		{"short password", `{"email":"user@example.com","password":"12345","language_code":"en_US"}`, http.StatusBadRequest, 0},
		{"password over bcrypt limit", `{"email":"user@example.com","password":"` + strings.Repeat("p", 73) + `","language_code":"en_US"}`, http.StatusBadRequest, 0},
		{"missing language code", `{"email":"user@example.com","password":"secret1"}`, http.StatusBadRequest, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, service := performRequest(t, register, tt.body)
			if code != tt.wantCode {
				t.Errorf("status = %d, want %d", code, tt.wantCode)
			}
			if service.registerCalls != tt.wantCalls {
				t.Errorf("service.Register calls = %d, want %d", service.registerCalls, tt.wantCalls)
			}
		})
	}
}

func TestLoginFailureIsLoggedAtWarnLevel(t *testing.T) {
	tests := []struct {
		name     string
		loginErr error
		wantCode int
		wantLogs int
	}{
		{"invalid credentials", authservice.ErrInvalidCredentials, http.StatusUnauthorized, 1},
		{"account not activated", authservice.ErrUserNotActivated, http.StatusForbidden, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logs bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&logs, &slog.HandlerOptions{Level: slog.LevelWarn}))

			e := echo.New()
			e.Validator = middleware.NewValidator()
			h := NewHandler(&fakeAuthService{loginErr: tt.loginErr}, logger)

			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"email":"user@example.com","password":"wrong"}`))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req.RemoteAddr = "203.0.113.7:1234"

			err := h.Login(e.NewContext(req, httptest.NewRecorder()))

			var httpErr *echo.HTTPError
			if !errors.As(err, &httpErr) || httpErr.Code != tt.wantCode {
				t.Fatalf("Login() error = %v, want HTTP %d", err, tt.wantCode)
			}

			lines := strings.Split(strings.TrimSpace(logs.String()), "\n")
			if len(lines) != tt.wantLogs || lines[0] == "" {
				t.Fatalf("log lines = %q, want %d line(s)", logs.String(), tt.wantLogs)
			}
			var entry map[string]any
			if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
				t.Fatalf("log entry is not JSON: %v", err)
			}
			if entry["level"] != "WARN" || entry["email"] != "user@example.com" || entry["remote_ip"] != "203.0.113.7" {
				t.Errorf("log entry = %v, want WARN with email and remote_ip", entry)
			}
			if strings.Contains(logs.String(), "wrong") {
				t.Errorf("log must not contain the password: %s", logs.String())
			}
		})
	}
}
