package middleware

import (
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
)

func Logger(logger *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			req := c.Request()
			res := c.Response()

			logger.Info("Request",
				slog.String("method", req.Method),
				slog.String("uri", req.RequestURI),
				slog.Int("status", res.Status),
				slog.Duration("latency", time.Since(start)),
				slog.String("remote_ip", c.RealIP()),
				slog.String("user_agent", req.UserAgent()),
			)

			return err
		}
	}
}