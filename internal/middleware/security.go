package middleware

import (
	"github.com/labstack/echo/v4"
)

// SecurityHeaders adds security headers to responses
func (m *Middleware) SecurityHeaders() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Prevent XSS attacks
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")
			c.Response().Header().Set("X-Frame-Options", "DENY")
			c.Response().Header().Set("X-XSS-Protection", "1; mode=block")

			// Prevent information disclosure
			c.Response().Header().Set("Server", "")
			c.Response().Header().Set("X-Powered-By", "")

			// Content Security Policy (basic)
			c.Response().Header().Set("Content-Security-Policy", "default-src 'self'")

			// HSTS in production
			if c.Request().Header.Get("X-Forwarded-Proto") == "https" {
				c.Response().Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}

			return next(c)
		}
	}
}

// RateLimit adds basic rate limiting (simple in-memory implementation)
func (m *Middleware) RateLimit() echo.MiddlewareFunc {
	// Note: For production, use Redis-based rate limiting
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Basic rate limiting logic would go here
			// For now, just pass through - implement with Redis in production
			return next(c)
		}
	}
}
