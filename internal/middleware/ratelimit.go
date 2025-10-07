package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"golang.org/x/time/rate"
)

// RateLimitConfig holds the configuration for rate limiting
type RateLimitConfig struct {
	RequestsPerSecond float64
	BurstSize         int
	WindowSize        time.Duration
	Storage           string // "memory" or "redis"
	RedisClient       *redis.Client
}

// RateLimiterInterface defines the interface for rate limiting implementations
type RateLimiterInterface interface {
	Allow(ip string) bool
	Cleanup() // Optional cleanup for implementations that need it
}

// IPLimiterEntry holds rate limiter with last used timestamp
type IPLimiterEntry struct {
	limiter  *rate.Limiter
	lastUsed time.Time
}

// IPRateLimiter manages rate limiters for different IPs with TTL cleanup
type IPRateLimiter struct {
	ips        map[string]*IPLimiterEntry
	mu         sync.RWMutex
	r          rate.Limit
	b          int
	cleanupTTL time.Duration
}

// NewIPRateLimiter creates a new IP-based rate limiter with TTL cleanup
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	limiter := &IPRateLimiter{
		ips:        make(map[string]*IPLimiterEntry),
		r:          r,
		b:          b,
		cleanupTTL: 1 * time.Hour, // Default TTL: 1 hour
	}

	// Start periodic cleanup goroutine
	go limiter.startPeriodicCleanup()

	return limiter
}

// RedisRateLimiter implements Redis-based rate limiting using token bucket algorithm
type RedisRateLimiter struct {
	client    *redis.Client
	keyPrefix string
	rps       float64
	burst     int
	window    time.Duration
}

// NewRedisRateLimiter creates a new Redis-based rate limiter
func NewRedisRateLimiter(client *redis.Client, rps float64, burst int) *RedisRateLimiter {
	return &RedisRateLimiter{
		client:    client,
		keyPrefix: "rate_limit:",
		rps:       rps,
		burst:     burst,
		window:    time.Second,
	}
}

// Allow implements token bucket algorithm using Redis
func (r *RedisRateLimiter) Allow(ip string) bool {
	ctx := context.Background()
	key := r.keyPrefix + ip
	now := time.Now().Unix()

	// Use Redis pipeline for atomic operations
	pipe := r.client.Pipeline()

	// Get current bucket state
	bucketKey := key + ":bucket"
	lastRefillKey := key + ":last_refill"

	// Get current tokens and last refill time
	tokensCmd := pipe.Get(ctx, bucketKey)
	lastRefillCmd := pipe.Get(ctx, lastRefillKey)

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		// On Redis error, allow request (fail open)
		return true
	}

	// Parse current state
	currentTokens := float64(r.burst) // Default to full bucket
	lastRefill := now                 // Default to now

	if tokensResult := tokensCmd.Val(); tokensResult != "" {
		if tokens, err := strconv.ParseFloat(tokensResult, 64); err == nil {
			currentTokens = tokens
		}
	}

	if lastRefillResult := lastRefillCmd.Val(); lastRefillResult != "" {
		if timestamp, err := strconv.ParseInt(lastRefillResult, 10, 64); err == nil {
			lastRefill = timestamp
		}
	}

	// Calculate token refill
	timePassed := float64(now - lastRefill)
	tokensToAdd := timePassed * r.rps
	newTokens := currentTokens + tokensToAdd

	// Cap at burst size
	if newTokens > float64(r.burst) {
		newTokens = float64(r.burst)
	}

	// Check if request can be allowed
	if newTokens < 1 {
		// Not enough tokens - set TTL and reject
		r.client.SetEX(ctx, bucketKey, fmt.Sprintf("%.2f", newTokens), time.Hour)
		r.client.SetEX(ctx, lastRefillKey, strconv.FormatInt(now, 10), time.Hour)
		return false
	}

	// Allow request - consume one token
	newTokens--

	// Update Redis state with TTL
	pipe = r.client.Pipeline()
	pipe.SetEX(ctx, bucketKey, fmt.Sprintf("%.2f", newTokens), time.Hour)
	pipe.SetEX(ctx, lastRefillKey, strconv.FormatInt(now, 10), time.Hour)
	pipe.Exec(ctx)

	return true
}

// Cleanup for RedisRateLimiter (no-op since Redis handles TTL)
func (r *RedisRateLimiter) Cleanup() {
	// Redis handles cleanup via TTL, nothing to do
}

// startPeriodicCleanup runs cleanup every 30 minutes in background
func (i *IPRateLimiter) startPeriodicCleanup() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		i.mu.Lock()
		i.cleanupExpiredEntries()
		i.mu.Unlock()
	}
}

// AddIP creates a new rate limiter and adds it to the map with TTL cleanup
func (i *IPRateLimiter) AddIP(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	// Create new limiter entry with current timestamp
	limiter := rate.NewLimiter(i.r, i.b)
	entry := &IPLimiterEntry{
		limiter:  limiter,
		lastUsed: time.Now(),
	}
	i.ips[ip] = entry

	// Perform TTL-based cleanup when map grows large
	if len(i.ips) > 1000 {
		i.cleanupExpiredEntries()
	}

	return limiter
}

// cleanupExpiredEntries removes entries that haven't been used within TTL
// Must be called with mutex locked
func (i *IPRateLimiter) cleanupExpiredEntries() {
	now := time.Now()
	for ip, entry := range i.ips {
		if now.Sub(entry.lastUsed) > i.cleanupTTL {
			delete(i.ips, ip)
		}
	}
}

// GetLimiter returns the rate limiter for the provided IP and updates lastUsed
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.RLock()
	entry, exists := i.ips[ip]
	if exists {
		// Update last used timestamp
		entry.lastUsed = time.Now()
		i.mu.RUnlock()
		return entry.limiter
	}
	i.mu.RUnlock()

	// Create new limiter if doesn't exist
	return i.AddIP(ip)
}

// Allow implements the RateLimiter interface for in-memory limiting
func (i *IPRateLimiter) Allow(ip string) bool {
	limiter := i.GetLimiter(ip)
	return limiter.Allow()
}

// Cleanup implements the RateLimiter interface for in-memory limiting
func (i *IPRateLimiter) Cleanup() {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.cleanupExpiredEntries()
}

// CreateRateLimiter creates appropriate rate limiter based on configuration
func CreateRateLimiter(cfg RateLimitConfig) RateLimiterInterface {
	switch cfg.Storage {
	case "redis":
		if cfg.RedisClient == nil {
			// Fallback to memory if Redis client not provided
			return NewIPRateLimiter(rate.Limit(cfg.RequestsPerSecond), cfg.BurstSize)
		}
		return NewRedisRateLimiter(cfg.RedisClient, cfg.RequestsPerSecond, cfg.BurstSize)
	default: // "memory" or any other value defaults to memory
		return NewIPRateLimiter(rate.Limit(cfg.RequestsPerSecond), cfg.BurstSize)
	}
}

// RateLimit returns a middleware that implements rate limiting per IP
func (m *Middleware) RateLimit(cfg RateLimitConfig) echo.MiddlewareFunc {
	limiter := CreateRateLimiter(cfg)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get client IP
			ip := c.RealIP()

			// Check if request is allowed using the configured limiter
			if !limiter.Allow(ip) {
				return echo.NewHTTPError(http.StatusTooManyRequests, "Rate limit exceeded")
			}

			return next(c)
		}
	}
}
