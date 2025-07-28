package middleware

import (
	"beresin-backend/internal/constant"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
)

// Logger returns a custom logging middleware using zerolog.
// It enriches logs with request_id and user_id from the context.
func (m *Middleware) Logger() echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:      true,
		LogStatus:   true,
		LogLatency:  true,
		LogMethod:   true,
		LogRemoteIP: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			// Get context from previous middleware
			requestID, _ := c.Get(constant.RequestIDKey).(string)
			userID, _ := c.Get(constant.UserIDKey).(uuid.UUID)

			// Build logger with context
			logger := log.Info()
			if requestID != "" {
				logger = logger.Str("request_id", requestID)
			}
			if userID != uuid.Nil {
				logger = logger.Str("user_id", userID.String())
			}

			logger.Str("URI", v.URI).Int("status", v.Status).Dur("latency", v.Latency).Str("method", v.Method).Str("remote_ip", v.RemoteIP).Msg("request")
			return nil
		},
	})
}
