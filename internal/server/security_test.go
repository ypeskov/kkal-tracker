package server

import (
	"embed"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ypeskov/kkal-tracker/internal/config"

	"github.com/labstack/echo/v4"
)

func discardLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func mustParseCIDR(t *testing.T, cidr string) *net.IPNet {
	t.Helper()
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		t.Fatalf("invalid CIDR %q: %v", cidr, err)
	}
	return ipNet
}

func TestNewIPExtractor(t *testing.T) {
	clusterOnly := []*net.IPNet{mustParseCIDR(t, "10.42.0.0/16")}

	tests := []struct {
		name           string
		trustedProxies []*net.IPNet
		remoteAddr     string
		forwardedFor   string
		want           string
	}{
		{"no header returns direct address", clusterOnly, "203.0.113.7:1234", "", "203.0.113.7"},
		{"header from untrusted client is ignored", clusterOnly, "203.0.113.7:1234", "198.51.100.1", "203.0.113.7"},
		{"header from trusted proxy is used", clusterOnly, "10.42.0.223:1234", "198.51.100.1", "198.51.100.1"},
		{"spoofed entries left of the proxy-appended one are ignored", clusterOnly, "10.42.0.223:1234", "1.2.3.4, 198.51.100.1", "198.51.100.1"},
		{"private network outside the configured range is not trusted", clusterOnly, "192.168.1.5:1234", "198.51.100.1", "192.168.1.5"},
		{"default trusts private networks", nil, "192.168.1.5:1234", "198.51.100.1", "198.51.100.1"},
		{"default trusts loopback", nil, "127.0.0.1:1234", "198.51.100.1", "198.51.100.1"},
		{"default ignores header from public client", nil, "203.0.113.7:1234", "198.51.100.1", "203.0.113.7"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.forwardedFor != "" {
				req.Header.Set(echo.HeaderXForwardedFor, tt.forwardedFor)
			}

			if got := newIPExtractor(tt.trustedProxies)(req); got != tt.want {
				t.Errorf("extracted IP = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestNewRateLimiter(t *testing.T) {
	e := echo.New()
	e.IPExtractor = echo.ExtractIPDirect()
	e.GET("/", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	}, newRateLimiter(discardLogger(), "test", 2, time.Hour))

	request := func(remoteAddr string) int {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = remoteAddr
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		return rec.Code
	}

	for i := 1; i <= 2; i++ {
		if code := request("203.0.113.7:1234"); code != http.StatusOK {
			t.Fatalf("request %d within burst: status = %d, want %d", i, code, http.StatusOK)
		}
	}
	if code := request("203.0.113.7:1234"); code != http.StatusTooManyRequests {
		t.Errorf("request over burst: status = %d, want %d", code, http.StatusTooManyRequests)
	}
	if code := request("203.0.113.8:1234"); code != http.StatusOK {
		t.Errorf("request from another IP: status = %d, want %d", code, http.StatusOK)
	}
}

// newTestServer builds the fully wired HTTP server without a database.
// Only requests that are rejected before reaching a repository may be sent to it.
func newTestServer(t *testing.T) *http.Server {
	t.Helper()

	cfg := &config.Config{
		Port:      "0",
		JWTSecret: strings.Repeat("s", 32),
		AppURL:    "http://localhost:8080",
	}
	return New(cfg, discardLogger(), nil, embed.FS{}).Start()
}

func serve(srv *http.Server, method, path, body string, headers map[string]string) int {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.RemoteAddr = "203.0.113.7:1234"
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, req)
	return rec.Code
}

func TestServerTimeouts(t *testing.T) {
	srv := newTestServer(t)

	if srv.ReadHeaderTimeout != readHeaderTimeout || srv.ReadTimeout != readTimeout ||
		srv.WriteTimeout != writeTimeout || srv.IdleTimeout != idleTimeout {
		t.Errorf("timeouts = %v/%v/%v/%v, want %v/%v/%v/%v",
			srv.ReadHeaderTimeout, srv.ReadTimeout, srv.WriteTimeout, srv.IdleTimeout,
			readHeaderTimeout, readTimeout, writeTimeout, idleTimeout)
	}
}

func TestServerBodyLimit(t *testing.T) {
	srv := newTestServer(t)

	body := `{"email":"` + strings.Repeat("a", 2<<20) + `"}`
	if code := serve(srv, http.MethodPost, "/api/auth/login", body, nil); code != http.StatusRequestEntityTooLarge {
		t.Errorf("oversized body: status = %d, want %d", code, http.StatusRequestEntityTooLarge)
	}
}

func TestServerLoginHasNoStrictRateLimit(t *testing.T) {
	srv := newTestServer(t)

	// Stays below the group-wide limit of 5 requests per second
	for i := 1; i <= 5; i++ {
		if code := serve(srv, http.MethodPost, "/api/auth/login", `{}`, nil); code != http.StatusBadRequest {
			t.Fatalf("attempt %d: status = %d, want %d", i, code, http.StatusBadRequest)
		}
	}
}

func TestServerRegisterRateLimit(t *testing.T) {
	srv := newTestServer(t)

	// Invalid payloads are rejected by validation, before the auth service is reached
	for i := 1; i <= registerRateBurst; i++ {
		if code := serve(srv, http.MethodPost, "/api/auth/register", `{}`, nil); code != http.StatusBadRequest {
			t.Fatalf("attempt %d within burst: status = %d, want %d", i, code, http.StatusBadRequest)
		}
	}
	if code := serve(srv, http.MethodPost, "/api/auth/register", `{}`, nil); code != http.StatusTooManyRequests {
		t.Errorf("attempt over burst: status = %d, want %d", code, http.StatusTooManyRequests)
	}
}

func TestServerRateLimitIgnoresSpoofedForwardedFor(t *testing.T) {
	srv := newTestServer(t)

	// The client is not a trusted proxy, so rotating X-Forwarded-For must not give it a fresh bucket
	codes := make([]int, 0, registerRateBurst+1)
	for i := 0; i <= registerRateBurst; i++ {
		headers := map[string]string{echo.HeaderXForwardedFor: "198.51.100." + string(rune('1'+i))}
		codes = append(codes, serve(srv, http.MethodPost, "/api/auth/register", `{}`, headers))
	}
	if last := codes[len(codes)-1]; last != http.StatusTooManyRequests {
		t.Errorf("statuses = %v, want the last one to be %d", codes, http.StatusTooManyRequests)
	}
}

func TestServerAPIv1RateLimitRunsBeforeKeyCheck(t *testing.T) {
	srv := newTestServer(t)

	// No API key: the first request is rejected by the key check, the next one by the limiter
	if code := serve(srv, http.MethodGet, "/api/v1/data", "", nil); code != http.StatusUnauthorized {
		t.Fatalf("first request: status = %d, want %d", code, http.StatusUnauthorized)
	}
	if code := serve(srv, http.MethodGet, "/api/v1/data", "", nil); code != http.StatusTooManyRequests {
		t.Errorf("second request: status = %d, want %d", code, http.StatusTooManyRequests)
	}
}
