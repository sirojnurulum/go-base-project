package handler

import (
	"go-base-project/internal/apperror"
	"go-base-project/internal/constant"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// CustomHTTPErrorHandler adalah handler terpusat untuk semua error di aplikasi.
func CustomHTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	requestID, _ := c.Get(constant.RequestIDKey).(string)
	logger := log.With().Str("request_id", requestID).Logger()

	// Coba cast ke AppError kustom kita
	if appErr, ok := err.(*apperror.AppError); ok {
		// Jika ada error asli, log error tersebut untuk debugging
		if appErr.Err != nil {
			logger.Error().Err(appErr.Err).Str("message", appErr.Message).Msg("Application error with underlying cause")
		} else {
			logger.Warn().Int("code", appErr.Code).Msg(appErr.Message)
		}
		c.JSON(appErr.Code, map[string]string{"error": appErr.Message})
		return
	}

	// Coba cast ke HTTPError bawaan Echo (misalnya dari validator)
	if httpErr, ok := err.(*echo.HTTPError); ok {
		logger.Warn().Err(httpErr).Msg("HTTP error")
		c.JSON(httpErr.Code, map[string]interface{}{"error": httpErr.Message})
		return
	}

	// Untuk semua error lain yang tidak terduga, kembalikan 500
	logger.Error().Err(err).Msg("Unhandled internal server error")
	c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
}
