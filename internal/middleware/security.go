package middleware

import (
	"beresin-backend/internal/config"
	"fmt"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

// SecurityHeaders adds comprehensive security headers to responses
func (m *Middleware) SecurityHeaders(cfg config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Prevent XSS attacks
			c.Response().Header().Set("X-Content-Type-Options", "nosniff")
			c.Response().Header().Set("X-Frame-Options", "DENY")
			c.Response().Header().Set("X-XSS-Protection", "1; mode=block")

			// Remove server information
			c.Response().Header().Set("Server", "")
			c.Response().Header().Set("X-Powered-By", "")

			// Content Security Policy
			csp := fmt.Sprintf("default-src 'self'; connect-src 'self' %s; img-src 'self' data: https:; font-src 'self' data:; style-src 'self' 'unsafe-inline'; script-src 'self'", cfg.FrontendURL)
			c.Response().Header().Set("Content-Security-Policy", csp)

			// Referrer Policy
			c.Response().Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

			// Permissions Policy
			c.Response().Header().Set("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=()")

			// HSTS for production
			if cfg.Env == "production" || c.Request().Header.Get("X-Forwarded-Proto") == "https" {
				c.Response().Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
				c.Response().Header().Set("Content-Security-Policy", csp+"; upgrade-insecure-requests")
			}

			return next(c)
		}
	}
}

// RateLimitConfig contains configuration for rate limiting
type RateLimitConfig struct {
	RequestsPerSecond float64
	BurstSize         int
	CleanupInterval   time.Duration
}

// RateLimiter handles rate limiting using in-memory store
type RateLimiter struct {
	visitors map[string]*rate.Limiter
	mu       sync.RWMutex
	rate     rate.Limit
	burst    int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerSecond float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*rate.Limiter),
		rate:     rate.Limit(requestsPerSecond),
		burst:    burst,
	}

	// Clean up old entries every 3 minutes
	go rl.cleanupVisitors()

	return rl
}

// Allow checks if a request from the given IP should be allowed
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.visitors[ip]
	if !exists {
		limiter = rate.NewLimiter(rl.rate, rl.burst)
		rl.visitors[ip] = limiter
	}

	return limiter.Allow()
}

// cleanupVisitors removes old entries from the visitors map
func (rl *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(time.Minute * 3)

		rl.mu.Lock()
		for ip, limiter := range rl.visitors {
			// Remove visitors that haven't made requests recently
			if limiter.TokensAt(time.Now()) == float64(rl.burst) {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimit creates a rate limiting middleware
func (m *Middleware) RateLimit(requestsPerSecond float64, burst int) echo.MiddlewareFunc {
	limiter := NewRateLimiter(requestsPerSecond, burst)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get client IP (consider X-Forwarded-For for reverse proxy setups)
			ip := c.RealIP()
			if ip == "" {
				ip = c.Request().RemoteAddr
			}

			if !limiter.Allow(ip) {
				return echo.NewHTTPError(429, map[string]string{
					"error":   "Too Many Requests",
					"message": "Rate limit exceeded. Please try again later.",
				})
			}

			return next(c)
		}
	}
}

// AuthRateLimit creates a stricter rate limit for authentication endpoints
func (m *Middleware) AuthRateLimit() echo.MiddlewareFunc {
	// More restrictive for auth endpoints: 5 requests per minute
	return m.RateLimit(5.0/60.0, 5)
}

// APIRateLimit creates a general rate limit for API endpoints
func (m *Middleware) APIRateLimit() echo.MiddlewareFunc {
	// General API: 100 requests per minute
	return m.RateLimit(100.0/60.0, 20)
}
