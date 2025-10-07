package router

import (
	"go-base-project/internal/bootstrap"
	"go-base-project/internal/config"
	"go-base-project/internal/constant"
	customMiddleware "go-base-project/internal/middleware"
	"net/http"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
)

// SetupRoutes mengkonfigurasi semua rute untuk aplikasi.
func SetupRoutes(e *echo.Echo, cfg config.Config, handlers *bootstrap.Handlers, m *customMiddleware.Middleware, redisClient *redis.Client) {
	// Middleware global - urutan sangat penting!
	e.Use(otelecho.Middleware("go-base-project"))
	e.Use(m.Prometheus())
	e.Use(m.RequestID()) // 1. Tambahkan Request ID ke konteks
	e.Use(m.Logger())    // 2. Gunakan logger kustom kita yang membaca Request ID
	e.Use(middleware.Recover())

	// ADD THIS: Security headers middleware
	if cfg.EnableSecurityHeaders {
		e.Use(m.SecurityHeaders())
	}

	// Rate limiting middleware - configurable storage
	if cfg.Env == "production" || cfg.RateLimitStorage != "" {
		rateLimitConfig := customMiddleware.RateLimitConfig{
			RequestsPerSecond: cfg.RateLimitRPS,
			BurstSize:         cfg.RateLimitBurst,
			Storage:           cfg.RateLimitStorage,
		}

		// Add Redis client if using Redis storage
		if cfg.RateLimitStorage == "redis" && redisClient != nil {
			rateLimitConfig.RedisClient = redisClient
		}

		e.Use(m.RateLimit(rateLimitConfig))
	}

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
		authRoutes.GET("/me", handlers.Auth.GetCurrentUser, m.JWT)
		authRoutes.POST("/switch-organization", handlers.Auth.SwitchOrganization, m.JWT)
	}

	// General role-related routes (accessible by authenticated users)
	roleRoutes := api.Group("/roles", m.JWT)
	{
		// DISABLED: Role approval request functionality
		// roleRoutes.POST("/approval-requests", handlers.Role.CreateRoleApprovalRequest)
		roleRoutes.GET("/predefined-options", handlers.Role.GetPredefinedRoleOptions)
	}

	// Organization routes (accessible by authenticated users)
	orgRoutes := api.Group("/organizations", m.JWT)
	{
		orgRoutes.GET("", handlers.Organization.ListOrganizations)
		orgRoutes.GET("/statistics", handlers.Organization.GetOrganizationStatistics)
		orgRoutes.GET("/:id", handlers.Organization.GetOrganization)
		orgRoutes.GET("/code/:code", handlers.Organization.GetOrganizationByCode)
		orgRoutes.POST("/join", handlers.Organization.JoinOrganization)
		orgRoutes.DELETE("/:id/leave", handlers.Organization.LeaveOrganization)
	}

	// User-specific organization routes
	userOrgRoutes := api.Group("/users", m.JWT)
	{
		userOrgRoutes.GET("/me/organizations", handlers.Organization.GetUserOrganizations)
	}

	// Organization-scoped routes - require organization context
	orgContextRoutes := api.Group("/organizations/:orgId", m.JWT, m.OrganizationContext())
	{
		// Example: Organization-specific data endpoints
		orgContextRoutes.GET("/dashboard", func(c echo.Context) error {
			orgID := c.Get(constant.OrganizationIDKey).(uuid.UUID)
			return c.JSON(http.StatusOK, map[string]interface{}{
				"message":         "Organization dashboard",
				"organization_id": orgID,
			})
		}, m.RequirePermission("dashboard:view"))

		// Multi-tenant role isolation demonstrations
		orgContextRoutes.GET("/users", handlers.User.ListUsersByOrganization, m.RequirePermission("users:read"))
		orgContextRoutes.POST("/users/:userId/assign-role", handlers.User.AssignRoleToUserInOrganization, m.RequirePermission("roles:assign"))
		orgContextRoutes.GET("/users/:userId/role", handlers.User.GetUserRoleInOrganization, m.RequirePermission("users:read"))

		orgContextRoutes.GET("/reports", func(c echo.Context) error {
			orgID := c.Get(constant.OrganizationIDKey).(uuid.UUID)
			return c.JSON(http.StatusOK, map[string]interface{}{
				"message":         "Organization reports",
				"organization_id": orgID,
			})
		}, m.RequirePermission("reports:view"))
	}

	adminRoutes := api.Group("/admin", m.JWT) // This middleware protects the group
	{
		adminRoutes.GET("/dashboard", func(c echo.Context) error {
			return c.JSON(http.StatusOK, map[string]string{"message": constant.MsgWelcomeAdmin})
		}, m.RequirePermission("dashboard:view"))

		roleRoutes := adminRoutes.Group("/roles")
		{
			roleRoutes.GET("", handlers.Role.ListRoles, m.RequirePermission("users:read"))
			roleRoutes.POST("", handlers.Role.CreateRole, m.RequirePermission("roles:create"))
			roleRoutes.PUT("/:id", handlers.Role.UpdateRole, m.RequirePermission("roles:update"))
			roleRoutes.PUT("/:id/permissions", handlers.Role.UpdateRolePermissions, m.RequirePermission("roles:assign"))

			// Organization-specific role routes
			roleRoutes.GET("/organization-types", handlers.Role.GetRolesForOrganizationType, m.RequirePermission("roles:read"))

			// DISABLED: Role approval management routes
			// roleRoutes.GET("/approval-requests", handlers.Role.ListRoleApprovalRequests, m.RequirePermission("roles:approve"))
			// roleRoutes.PUT("/approval-requests/:id/decision", handlers.Role.ApproveRejectRoleRequest, m.RequirePermission("roles:approve"))
		}

		// Permission management routes
		permissionRoutes := adminRoutes.Group("/permissions")
		{
			permissionRoutes.GET("", handlers.Role.ListPermissions, m.RequirePermission("permissions:read"))
			permissionRoutes.POST("", handlers.Role.CreatePermission, m.RequirePermission("permissions:create"))
			permissionRoutes.PUT("/:id", handlers.Role.UpdatePermission, m.RequirePermission("permissions:update"))
			permissionRoutes.DELETE("/:id", handlers.Role.DeletePermission, m.RequirePermission("permissions:delete"))
		}

		userRoutes := adminRoutes.Group("/users")
		{
			// Basic user management
			userRoutes.POST("", handlers.User.CreateUser, m.RequirePermission("users:create"))
			userRoutes.GET("", handlers.User.ListUsers, m.RequirePermission("users:read"))
			userRoutes.GET("/:id", handlers.User.GetUserByID, m.RequirePermission("users:read"))
			userRoutes.PUT("/:id", handlers.User.UpdateUser, m.RequirePermission("users:update"))
			userRoutes.DELETE("/:id", handlers.User.DeleteUser, m.RequirePermission("users:delete"))

			// User-Organization Management
			userRoutes.POST("/assign-organization", handlers.User.AssignUserToOrganization, m.RequirePermission("users:assign-organization"))
			userRoutes.POST("/bulk-assign-organization", handlers.User.BulkAssignUsersToOrganization, m.RequirePermission("users:bulk-assign-organization"))
			userRoutes.POST("/log-organization-action", handlers.User.LogUserOrganizationAction, m.RequirePermission("users:log-actions"))
			userRoutes.GET("/:userId/organizations", handlers.User.GetUserOrganizations, m.RequirePermission("users:read"))
			userRoutes.GET("/:userId/organization-history", handlers.User.GetUserOrganizationHistory, m.RequirePermission("users:read-history"))
			userRoutes.PUT("/:userId/organizations/:organizationId", handlers.User.UpdateUserOrganizationRole, m.RequirePermission("users:update-organization-role"))
			userRoutes.DELETE("/:userId/organizations/:organizationId", handlers.User.RemoveUserFromOrganization, m.RequirePermission("users:remove-organization"))
		}

		// Admin organization management routes
		organizationRoutes := adminRoutes.Group("/organizations")
		{
			organizationRoutes.POST("", handlers.Organization.CreateOrganization, m.RequirePermission("organizations:create"))
			organizationRoutes.PUT("/:id", handlers.Organization.UpdateOrganization, m.RequirePermission("organizations:update"))
			organizationRoutes.DELETE("/:id", handlers.Organization.DeleteOrganization, m.RequirePermission("organizations:delete"))
			organizationRoutes.GET("/:id/members", handlers.Organization.GetOrganizationMembers, m.RequirePermission("organizations:read"))
			organizationRoutes.GET("/:organizationId/members", handlers.User.GetOrganizationMembers, m.RequirePermission("organizations:read-members"))
			organizationRoutes.GET("/:organizationId/user-history", handlers.User.GetOrganizationUserHistory, m.RequirePermission("organizations:read-history"))
			organizationRoutes.POST("/complete-structure", handlers.Organization.CreateCompleteOrganizationStructure, m.RequirePermission("organizations:create"))
		}
	}
}
