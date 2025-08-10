package config

import (
	"fmt"
	"os"
	"strconv"
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
	EnableRateLimit       bool
	APIRequestsPerMinute  int
	AuthRequestsPerMinute int
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

	// Parse rate limiting settings
	enableRateLimit := getEnv("ENABLE_RATE_LIMIT", "true") == "true"
	apiRequestsPerMinute, err := strconv.Atoi(getEnv("API_REQUESTS_PER_MINUTE", "1000"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid API_REQUESTS_PER_MINUTE value: %w", err)
	}
	authRequestsPerMinute, err := strconv.Atoi(getEnv("AUTH_REQUESTS_PER_MINUTE", "10"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid AUTH_REQUESTS_PER_MINUTE value: %w", err)
	}

	cfg := Config{
		Port:                  getEnv("PORT", "8080"),
		DatabaseURL:           getEnv("DATABASE_URL", ""),
		JWTSecret:             getEnv("JWT_SECRET", ""),
		Env:                   getEnv("ENV", "development"),
		RedisURL:              getEnv("REDIS_URL", "redis://localhost:6379/0"),
		AdminDefaultPassword:  getEnv("ADMIN_DEFAULT_PASSWORD", "password"),
		GoogleClientID:        getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:    getEnv("GOOGLE_CLIENT_SECRET", ""),
		FrontendURL:           getEnv("FRONTEND_URL", "http://localhost:5173"),
		DBMaxOpenConns:        dbMaxOpenConns,
		DBMaxIdleConns:        dbMaxIdleConns,
		DBConnMaxLifetime:     dbConnMaxLifetime,
		DBConnMaxIdleTime:     dbConnMaxIdleTime,
		EnableRateLimit:       enableRateLimit,
		APIRequestsPerMinute:  apiRequestsPerMinute,
		AuthRequestsPerMinute: authRequestsPerMinute,
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
		return value
	}
	return fallback
}
