package handler

import (
	"beresin-backend/internal/constant"
	"context"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// HealthHandler menangani request health check dengan validasi dependensi.
type HealthHandler struct {
	db    *gorm.DB
	redis *redis.Client
}

// NewHealthHandler membuat instance baru dari HealthHandler dengan dependensi database dan redis.
func NewHealthHandler(db *gorm.DB, redis *redis.Client) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
	}
}

// PublicHealthCheck
// @Summary Public Health Check
// @Description Basic health check endpoint that doesn't require authentication
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]interface{} "{"status": "ok", "timestamp": "2025-08-08T10:30:00Z"}"
// @Router /health/public [get]
func (h *HealthHandler) PublicHealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":    constant.MsgStatusOK,
		"timestamp": time.Now().UTC(),
		"service":   "beresin-backend",
	})
}

// PrivateHealthCheck
// @Summary Private Health Check
// @Description Comprehensive health check that validates database and redis connectivity
// @Security BearerAuth
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]interface{} "Detailed health status"
// @Failure 401 {object} map[string]string "{"error": "unauthorized"}"
// @Failure 503 {object} map[string]interface{} "Service unavailable due to dependency failure"
// @Router /health/private [get]
func (h *HealthHandler) PrivateHealthCheck(c echo.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	healthStatus := map[string]interface{}{
		"status":    constant.MsgStatusOK,
		"timestamp": time.Now().UTC(),
		"service":   "beresin-backend",
		"checks": map[string]interface{}{
			"database": h.checkDatabase(ctx),
			"redis":    h.checkRedis(ctx),
		},
	}

	// Check if any dependency failed
	checks := healthStatus["checks"].(map[string]interface{})
	for _, check := range checks {
		if checkResult, ok := check.(map[string]interface{}); ok {
			if status, exists := checkResult["status"]; exists && status != "healthy" {
				healthStatus["status"] = "degraded"
				return c.JSON(http.StatusServiceUnavailable, healthStatus)
			}
		}
	}

	return c.JSON(http.StatusOK, healthStatus)
}

func (h *HealthHandler) checkDatabase(ctx context.Context) map[string]interface{} {
	if h.db == nil {
		return map[string]interface{}{
			"status": "unhealthy",
			"error":  "database connection not initialized",
		}
	}

	sqlDB, err := h.db.DB()
	if err != nil {
		return map[string]interface{}{
			"status": "unhealthy",
			"error":  "failed to get underlying sql.DB",
		}
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
	}

	return map[string]interface{}{
		"status": "healthy",
		"stats": map[string]interface{}{
			"open_connections": sqlDB.Stats().OpenConnections,
			"in_use":           sqlDB.Stats().InUse,
			"idle":             sqlDB.Stats().Idle,
		},
	}
}

func (h *HealthHandler) checkRedis(ctx context.Context) map[string]interface{} {
	if h.redis == nil {
		return map[string]interface{}{
			"status": "unhealthy",
			"error":  "redis connection not initialized",
		}
	}

	_, err := h.redis.Ping(ctx).Result()
	if err != nil {
		return map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		}
	}

	return map[string]interface{}{
		"status": "healthy",
	}
}
