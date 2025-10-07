package handler

import (
	"go-base-project/internal/constant"
	"net/http"

	"github.com/labstack/echo/v4"
)

// HealthHandler menangani request health check.
type HealthHandler struct{}

// NewHealthHandler membuat instance baru dari HealthHandler.
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// PublicHealthCheck
// @Summary Public Health Check
// @Description Endpoint ini tidak memerlukan otentikasi.
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]string "{"status": "ok"}"
// @Router /health/public [get]
func (h *HealthHandler) PublicHealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": constant.MsgStatusOK})
}

// PrivateHealthCheck
// @Summary Private Health Check
// @Description Endpoint ini memerlukan otentikasi.
// @Security BearerAuth
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]string "{"status": "ok", "message": "authenticated"}"
// @Failure 401 {object} map[string]string "{"error": "unauthorized"}"
// @Router /health/private [get]
func (h *HealthHandler) PrivateHealthCheck(c echo.Context) error {
	// Middleware otentikasi seharusnya sudah memverifikasi token
	// Jika sampai di sini, berarti user sudah login
	return c.JSON(http.StatusOK, map[string]string{"status": constant.MsgStatusOK, "message": constant.MsgAuthenticated})
}
