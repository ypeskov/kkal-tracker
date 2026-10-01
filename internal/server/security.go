package server

import (
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

const (
	readHeaderTimeout = 10 * time.Second
	readTimeout       = 30 * time.Second
	// Must stay above the longest handler: AI analysis may run for up to 30 seconds
	writeTimeout = 60 * time.Second
	idleTimeout  = 120 * time.Second

	// No endpoint accepts uploads, JSON payloads are small
	maxRequestBodySize = "1M"

	// Register sends an email: 3 attempts at once, then one attempt every 20 minutes per IP
	registerRateBurst    = 3
	registerRateInterval = 20 * time.Minute

	minRateLimiterExpiry = 3 * time.Minute

	// The frontend is a single bundle served from this origin and talks only to /api.
	// Scripts: no inline code and no eval. Styles: inline style attributes are allowed
	// (sanitized AI markup, chart canvases). Images: data: and blob: for generated content.
	contentSecurityPolicy = "default-src 'self'; " +
		"script-src 'self'; " +
		"style-src 'self' 'unsafe-inline'; " +
		"img-src 'self' data: blob:; " +
		"font-src 'self'; " +
		"connect-src 'self'; " +
		"object-src 'none'; " +
		"base-uri 'self'; " +
		"form-action 'self'; " +
		"frame-ancestors 'none'"
)

// newIPExtractor returns the extractor used by c.RealIP(), which keys the rate limiters.
// X-Forwarded-For is honored only when the request comes from a trusted proxy, so a client
// cannot rotate the header to get a fresh rate limit bucket. With no explicit ranges
// the Echo defaults apply: loopback, link-local and private networks are trusted.
func newIPExtractor(trustedProxies []*net.IPNet) echo.IPExtractor {
	if len(trustedProxies) == 0 {
		return echo.ExtractIPFromXFFHeader()
	}

	options := []echo.TrustOption{
		echo.TrustLoopback(false),
		echo.TrustLinkLocal(false),
		echo.TrustPrivateNet(false),
	}
	for _, ipRange := range trustedProxies {
		options = append(options, echo.TrustIPRange(ipRange))
	}
	return echo.ExtractIPFromXFFHeader(options...)
}

// newRateLimiter returns a per-IP rate limiter that lets burst requests through at once
// and then one more request per interval. Denied requests are logged at warn level.
// NOTE: The in-memory store is per-instance. If scaling to multiple replicas,
// switch to a distributed store (e.g., Redis) for effective rate limiting.
func newRateLimiter(logger *slog.Logger, name string, burst int, interval time.Duration) echo.MiddlewareFunc {
	// Keep a visitor at least until its bucket is full again, otherwise cleanup would reset the limit early
	expiresIn := max(time.Duration(burst)*interval, minRateLimiterExpiry)

	return echomiddleware.RateLimiterWithConfig(echomiddleware.RateLimiterConfig{
		Store: echomiddleware.NewRateLimiterMemoryStoreWithConfig(echomiddleware.RateLimiterMemoryStoreConfig{
			Rate:      rate.Every(interval),
			Burst:     burst,
			ExpiresIn: expiresIn,
		}),
		DenyHandler: func(c echo.Context, identifier string, err error) error {
			logger.Warn("Rate limit exceeded", "limiter", name, "remote_ip", identifier, "path", c.Path())
			return echo.NewHTTPError(http.StatusTooManyRequests, "Too many requests, please try again later")
		},
	})
}
