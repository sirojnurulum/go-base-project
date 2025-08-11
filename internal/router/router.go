package router

import (
	"beresin-backend/internal/bootstrap"
	"beresin-backend/internal/config"
	"beresin-backend/internal/constant"
	customMiddleware "beresin-backend/internal/middleware"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// SetupRoutes mengkonfigurasi semua rute untuk aplikasi dengan security dan rate limiting.
func SetupRoutes(e *echo.Echo, cfg config.Config, handlers *bootstrap.Handlers, m *customMiddleware.Middleware, redisClient *redis.Client) {
	// Middleware global - urutan sangat penting!
	e.Use(m.Prometheus())
	e.Use(m.RequestID()) // 1. Tambahkan Request ID ke konteks
	e.Use(m.Logger())    // 2. Gunakan logger kustom kita yang membaca Request ID
	e.Use(middleware.Recover())

	// Security Headers - Apply to all routes
	e.Use(m.SecurityHeaders(cfg))

	// General API Rate Limiting
	if cfg.Env == "production" && redisClient != nil {
		// Use Redis-based rate limiting in production
		e.Use(m.RedisAPIRateLimit(redisClient))
	} else {
		// Use in-memory rate limiting for development
		e.Use(m.APIRateLimit())
	}

	// CORS Configuration
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{cfg.FrontendURL},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, echo.HeaderXRequestID},
		AllowMethods:     []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
		AllowCredentials: true,
	}))

	// Metrics endpoint (no rate limiting for monitoring)
	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))

	// Swagger documentation (no rate limiting for docs)
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	api := e.Group("/api")

	// Health check routes (lighter rate limiting)
	healthRoutes := api.Group("/health")
	{
		healthRoutes.GET("/public", handlers.Health.PublicHealthCheck)
		healthRoutes.GET("/private", handlers.Health.PrivateHealthCheck, m.JWT)
	}

	// Auth routes with stricter rate limiting
	authRoutes := api.Group("/auth")

	// Apply auth-specific rate limiting
	if cfg.Env == "production" && redisClient != nil {
		authRoutes.Use(m.RedisAuthRateLimit(redisClient))
	} else {
		authRoutes.Use(m.AuthRateLimit())
	}

	{
		authRoutes.POST("/login", handlers.Auth.Login)
		authRoutes.POST("/refresh", handlers.Auth.RefreshToken)
		authRoutes.POST("/logout", handlers.Auth.Logout)
		authRoutes.GET("/google/login", handlers.Auth.GoogleLogin)
		authRoutes.GET("/google/callback", handlers.Auth.GoogleCallback)
	}

	// Admin routes with JWT protection
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

		organizationRoutes := adminRoutes.Group("/organizations")
		{
			organizationRoutes.POST("", handlers.Organization.CreateOrganization, m.RequirePermission("organizations:create"))
			organizationRoutes.GET("", handlers.Organization.GetAllOrganizations, m.RequirePermission("organizations:read"))
			organizationRoutes.GET("/:id", handlers.Organization.GetOrganizationByID, m.RequirePermission("organizations:read"))
			organizationRoutes.PUT("/:id", handlers.Organization.UpdateOrganization, m.RequirePermission("organizations:update"))
			organizationRoutes.DELETE("/:id", handlers.Organization.DeleteOrganization, m.RequirePermission("organizations:delete"))
			organizationRoutes.POST("/:id/join", handlers.Organization.JoinOrganization, m.RequirePermission("organizations:join"))
			organizationRoutes.DELETE("/:id/leave", handlers.Organization.LeaveOrganization, m.RequirePermission("organizations:leave"))
			organizationRoutes.GET("/:id/members", handlers.Organization.GetOrganizationMembers, m.RequirePermission("organizations:read"))
		}
	}
}
