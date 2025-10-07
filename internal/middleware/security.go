package middleware

import (
	"github.com/labstack/echo/v4"
)

// SecurityHeaders adds security-related HTTP headers
func (m *Middleware) SecurityHeaders() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Prevent XSS attacks
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")
			c.Response().Header().Set("X-Frame-Options", "DENY")
			c.Response().Header().Set("X-XSS-Protection", "1; mode=block")

			// HSTS for HTTPS (only set if using HTTPS)
			if c.Request().TLS != nil || c.Request().Header.Get("X-Forwarded-Proto") == "https" {
				c.Response().Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			}

			// Content Security Policy (basic policy)
			c.Response().Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'")

			// Referrer Policy
			c.Response().Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Permissions Policy (formerly Feature Policy)
			c.Response().Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

			return next(c)
		}
	}
}
