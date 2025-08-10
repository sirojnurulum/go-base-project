package middleware

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// RedisRateLimiter handles rate limiting using Redis
type RedisRateLimiter struct {
	client *redis.Client
}

// NewRedisRateLimiter creates a new Redis-based rate limiter
func NewRedisRateLimiter(client *redis.Client) *RedisRateLimiter {
	return &RedisRateLimiter{
		client: client,
	}
}

// Allow checks if a request should be allowed using sliding window algorithm
func (rrl *RedisRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	if rrl.client == nil {
		// Fallback to allow if Redis is not available
		log.Warn().Msg("Redis client not available for rate limiting, allowing request")
		return true, nil
	}

	pipeline := rrl.client.Pipeline()
	now := time.Now().Unix()
	windowStart := now - int64(window.Seconds())

	// Remove old entries outside the time window
	pipeline.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart, 10))

	// Count current entries in the window
	countCmd := pipeline.ZCard(ctx, key)

	// Add current request
	pipeline.ZAdd(ctx, key, &redis.Z{
		Score:  float64(now),
		Member: fmt.Sprintf("%d-%d", now, time.Now().Nanosecond()), // Unique member
	})

	// Set expiration for the key
	pipeline.Expire(ctx, key, window+time.Minute) // Extra minute for cleanup

	// Execute pipeline
	_, err := pipeline.Exec(ctx)
	if err != nil {
		log.Error().Err(err).Str("key", key).Msg("Redis rate limiting pipeline failed")
		// Allow request if Redis fails
		return true, err
	}

	// Check if limit is exceeded
	count := countCmd.Val()
	return count < int64(limit), nil
}

// RedisRateLimit creates a Redis-based rate limiting middleware
func (m *Middleware) RedisRateLimit(client *redis.Client, requestsPerWindow int, window time.Duration, keyPrefix string) echo.MiddlewareFunc {
	limiter := NewRedisRateLimiter(client)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ctx := context.Background()

			// Create rate limit key based on IP and prefix
			ip := c.RealIP()
			if ip == "" {
				ip = c.Request().RemoteAddr
			}

			key := fmt.Sprintf("rate_limit:%s:%s", keyPrefix, ip)

			allowed, err := limiter.Allow(ctx, key, requestsPerWindow, window)
			if err != nil {
				// Log error but continue (fail-open approach)
				log.Error().Err(err).Str("ip", ip).Msg("Rate limiting check failed")
			}

			if !allowed {
				return echo.NewHTTPError(429, map[string]interface{}{
					"error":       "Too Many Requests",
					"message":     "Rate limit exceeded. Please try again later.",
					"retry_after": int(window.Seconds()),
				})
			}

			return next(c)
		}
	}
}

// RedisAuthRateLimit creates stricter rate limiting for auth endpoints
func (m *Middleware) RedisAuthRateLimit(client *redis.Client) echo.MiddlewareFunc {
	// 10 requests per 5 minutes for auth endpoints
	return m.RedisRateLimit(client, 10, 5*time.Minute, "auth")
}

// RedisAPIRateLimit creates general rate limiting for API endpoints
func (m *Middleware) RedisAPIRateLimit(client *redis.Client) echo.MiddlewareFunc {
	// 1000 requests per 15 minutes for general API
	return m.RedisRateLimit(client, 1000, 15*time.Minute, "api")
}
