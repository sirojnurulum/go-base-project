package router

import (
	"beresin-backend/internal/bootstrap"
	"beresin-backend/internal/config"
	"beresin-backend/internal/constant"
	customMiddleware "beresin-backend/internal/middleware"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// SetupRoutes mengkonfigurasi semua rute untuk aplikasi.
func SetupRoutes(e *echo.Echo, cfg config.Config, handlers *bootstrap.Handlers, m *customMiddleware.Middleware) {
	// Middleware global - urutan sangat penting!
	e.Use(m.Prometheus())
	e.Use(m.RequestID()) // 1. Tambahkan Request ID ke konteks
	e.Use(m.Logger())    // 2. Gunakan logger kustom kita yang membaca Request ID
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{cfg.FrontendURL},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, echo.HeaderXRequestID},
		AllowMethods:     []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
		AllowCredentials: true,
	}))

	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	api := e.Group("/api")

	healthRoutes := api.Group("/health")
	{
		healthRoutes.GET("/public", handlers.Health.PublicHealthCheck)
		healthRoutes.GET("/private", handlers.Health.PrivateHealthCheck, m.JWT)
	}

	authRoutes := api.Group("/auth")
	{
		authRoutes.POST("/login", handlers.Auth.Login)
		authRoutes.POST("/refresh", handlers.Auth.RefreshToken)
		authRoutes.POST("/logout", handlers.Auth.Logout)
		authRoutes.GET("/google/login", handlers.Auth.GoogleLogin)
		authRoutes.GET("/google/callback", handlers.Auth.GoogleCallback)
	}

	adminRoutes := api.Group("/admin", m.JWT) // This middleware protects the group
	{
		adminRoutes.GET("/dashboard", func(c echo.Context) error {
			return c.JSON(http.StatusOK, map[string]string{"message": constant.MsgWelcomeAdmin})
		}, m.RequirePermission("dashboard:view"))

		roleRoutes := adminRoutes.Group("/roles")
		{
			roleRoutes.POST("", handlers.Role.CreateRole, m.RequirePermission("roles:create"))
			roleRoutes.PUT("/:id/permissions", handlers.Role.UpdateRolePermissions, m.RequirePermission("roles:assign"))
		}

		userRoutes := adminRoutes.Group("/users")
		{
			userRoutes.POST("", handlers.User.CreateUser, m.RequirePermission("users:create"))
			userRoutes.GET("", handlers.User.ListUsers, m.RequirePermission("users:read"))
			userRoutes.GET("/:id", handlers.User.GetUserByID, m.RequirePermission("users:read"))
			userRoutes.PUT("/:id", handlers.User.UpdateUser, m.RequirePermission("users:update"))
			userRoutes.DELETE("/:id", handlers.User.DeleteUser, m.RequirePermission("users:delete"))
		}
	}
}
