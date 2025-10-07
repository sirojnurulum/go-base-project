package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

// Config holds all configuration for the application.
type Config struct {
	Port                 string
	DatabaseURL          string
	JWTSecret            string
	Env                  string
	RedisURL             string
	AdminDefaultPassword string
	GoogleClientID       string
	GoogleClientSecret   string
	FrontendURL          string

	// Database Connection Pool Settings
	DBMaxOpenConns    int
	DBMaxIdleConns    int
	DBConnMaxLifetime time.Duration
	DBConnMaxIdleTime time.Duration

	// Rate Limiting Settings
	RateLimitRPS     float64
	RateLimitBurst   int
	RateLimitStorage string // "memory" or "redis"

	// Security Settings
	// Security and Debugging
	EnableSecurityHeaders bool
	EnableDetailedTracing bool
	CookieSecure          bool
	CookieSameSite        string
}

// Load loads environment variables from a .env file or from the system environment.
func Load(path ...string) (Config, error) {
	// Load .env file. It's okay if it doesn't exist, as prod envs will be set in the system.
	err := godotenv.Load(path...)
	if err != nil && len(path) > 0 { // Hanya error jika path spesifik diberikan dan gagal
		return Config{}, fmt.Errorf("error loading .env file from path %s: %w", path[0], err)
	} else if err != nil {
		log.Info().Msg("No .env file found, loading from system environment variables")
	}

	dbMaxOpenConns, err := strconv.Atoi(getEnv("DB_MAX_OPEN_CONNS", "25"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid DB_MAX_OPEN_CONNS value: %w", err)
	}
	dbMaxIdleConns, err := strconv.Atoi(getEnv("DB_MAX_IDLE_CONNS", "10"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid DB_MAX_IDLE_CONNS value: %w", err)
	}
	dbConnMaxLifetime, err := time.ParseDuration(getEnv("DB_CONN_MAX_LIFETIME", "5m"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid DB_CONN_MAX_LIFETIME value: %w", err)
	}
	dbConnMaxIdleTime, err := time.ParseDuration(getEnv("DB_CONN_MAX_IDLE_TIME", "1m"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid DB_CONN_MAX_IDLE_TIME value: %w", err)
	}

	// Rate limiting configuration
	rateLimitRPS, err := strconv.ParseFloat(getEnv("RATE_LIMIT_RPS", "10"), 64)
	if err != nil {
		return Config{}, fmt.Errorf("invalid RATE_LIMIT_RPS value: %w", err)
	}

	rateLimitBurst, err := strconv.Atoi(getEnv("RATE_LIMIT_BURST", "20"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid RATE_LIMIT_BURST value: %w", err)
	}

	env := getEnv("ENV", "development")
	cfg := Config{
		Port:                  getEnv("PORT", "8080"),
		DatabaseURL:           getEnv("DATABASE_URL", ""),
		JWTSecret:             getEnv("JWT_SECRET", ""),
		Env:                   env,
		RedisURL:              getEnv("REDIS_URL", "redis://localhost:6379/0"),
		AdminDefaultPassword:  getEnv("ADMIN_DEFAULT_PASSWORD", "password"),
		GoogleClientID:        getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:    getEnv("GOOGLE_CLIENT_SECRET", ""),
		FrontendURL:           getEnv("FRONTEND_URL", "http://localhost:5173"),
		DBMaxOpenConns:        dbMaxOpenConns,
		DBMaxIdleConns:        dbMaxIdleConns,
		DBConnMaxLifetime:     dbConnMaxLifetime,
		DBConnMaxIdleTime:     dbConnMaxIdleTime,
		EnableSecurityHeaders: env == "production",
		EnableDetailedTracing: getEnvBool("ENABLE_DETAILED_TRACING", env == "development"),
		CookieSecure:          env == "production",
		CookieSameSite:        getEnv("COOKIE_SAME_SITE", "lax"),
		RateLimitRPS:          rateLimitRPS,
		RateLimitBurst:        rateLimitBurst,
		RateLimitStorage:      getEnv("RATE_LIMIT_STORAGE", "memory"), // default: memory
	}

	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("FATAL: JWT_SECRET environment variable is not set")
	}

	// In a production environment, Google variables must also be present.
	if cfg.Env == "production" {
		if cfg.GoogleClientID == "" || cfg.GoogleClientSecret == "" {
			return Config{}, fmt.Errorf("FATAL: GOOGLE_CLIENT_ID and GOOGLE_CLIENT_SECRET must be set in production")
		}
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return strings.TrimSpace(value)
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if value, ok := os.LookupEnv(key); ok {
		parsed, err := strconv.ParseBool(strings.TrimSpace(value))
		if err != nil {
			return fallback
		}
		return parsed
	}
	return fallback
}
