package middleware

import (
	"go-base-project/internal/constant"
	"go-base-project/internal/service"
	"go-base-project/internal/util"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
)

// Middleware provides a container for all application middleware that require dependencies.
type Middleware struct {
	authorizationService service.AuthorizationServiceInterface
	jwtSecret            string
}

// NewMiddleware creates a new instance of the Middleware provider.
// Note that we only inject the JWTSecret, not the entire config struct.
func NewMiddleware(authorizationService service.AuthorizationServiceInterface, jwtSecret string) *Middleware {
	return &Middleware{
		authorizationService: authorizationService,
		jwtSecret:            jwtSecret,
	}
}

// RequestID adds a request ID to the context of each request.
func (m *Middleware) RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get request ID from header or generate a new one
			reqID := c.Request().Header.Get(echo.HeaderXRequestID)
			if reqID == "" {
				reqID = uuid.New().String()
			}
			c.Set(constant.RequestIDKey, reqID)
			c.Response().Header().Set(echo.HeaderXRequestID, reqID)
			return next(c)
		}
	}
}

// Prometheus returns a middleware that collects Prometheus metrics.
func (m *Middleware) Prometheus() echo.MiddlewareFunc {
	p := prometheus.NewPrometheus("echo", nil)
	return p.HandlerFunc
}

// JWT is a middleware for validating JWTs.
// This middleware is also responsible for placing user info into the context.
func (m *Middleware) JWT(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		claims, err := util.VerifyAndGetClaims(c, m.jwtSecret)
		if err != nil {
			// Use HTTPError to be handled by the centralized error handler
			return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
		}

		// Set user ID and role ID in the context for subsequent handlers.
		// Use constants for keys to maintain consistency.
		c.Set(constant.UserIDKey, claims.UserID)
		c.Set(constant.RoleIDKey, claims.RoleID)

		// Set organization ID if present in claims
		if claims.OrganizationID != nil {
			c.Set(constant.OrganizationIDKey, *claims.OrganizationID)
		}

		return next(c)
	}
}

// OrganizationContext creates a middleware to ensure organization context is properly set.
// This middleware extracts organization context from JWT claims or request headers and validates user access.
// Super admin bypasses organization access validation.
func (m *Middleware) OrganizationContext() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get user ID from context (set by JWT middleware)
			userIDValue := c.Get(constant.UserIDKey)
			userID, ok := userIDValue.(uuid.UUID)
			if !ok {
				return echo.NewHTTPError(http.StatusUnauthorized, constant.ErrMsgUserNotFoundInContext)
			}

			// Check if user is super admin - if yes, bypass organization access validation
			roleIDValue := c.Get(constant.RoleIDKey)
			roleID, ok := roleIDValue.(uuid.UUID)
			if ok {
				isSuperAdmin, err := m.authorizationService.IsRoleSuperAdmin(c.Request().Context(), roleID)
				if err == nil && isSuperAdmin {
					// For super admin, still extract organization ID for context but skip access validation
					organizationID, hasOrganizationContext := m.extractOrganizationID(c)

					// For super admin, if organization context is not found, we still return error
					// because organization-scoped endpoints need organization context
					if !hasOrganizationContext {
						return echo.NewHTTPError(http.StatusBadRequest, constant.ErrMsgOrganizationContextRequired)
					}

					// Set organization ID in context and proceed (skip access validation for super admin)
					c.Set(constant.OrganizationIDKey, organizationID)
					return next(c)
				}
			}

			// Extract organization ID from JWT claims or headers
			organizationID, hasOrganizationContext := m.extractOrganizationID(c)

			// If organization context is required but not found, return error
			if !hasOrganizationContext {
				return echo.NewHTTPError(http.StatusBadRequest, constant.ErrMsgOrganizationContextRequired)
			}

			// Validate that the user has access to this organization (for non-super admin users)
			hasAccess, err := m.authorizationService.CheckUserOrganizationAccess(c.Request().Context(), userID, organizationID)
			if err != nil {
				c.Logger().Errorf("organization access check failed: %v", err)
				return echo.NewHTTPError(http.StatusForbidden, constant.ErrMsgInsufficientPermissions)
			}
			if !hasAccess {
				return echo.NewHTTPError(http.StatusForbidden, constant.ErrMsgOrganizationAccessDenied)
			}

			// Set organization ID in context for subsequent handlers
			c.Set(constant.OrganizationIDKey, organizationID)

			return next(c)
		}
	}
}

// RequirePermission creates a middleware to check if a user has a given permission.
// This middleware now reads the roleID from the context, no longer parsing the token itself.
// If organization context is present, it uses organization-scoped permission checking.
// Super admin with level 100 bypasses all permission checking.
func (m *Middleware) RequirePermission(permissionName string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// First, check if user has super admin role (bypasses all permission checks)
			roleIDValue := c.Get(constant.RoleIDKey)
			roleID, ok := roleIDValue.(uuid.UUID)
			if !ok {
				return echo.NewHTTPError(http.StatusForbidden, constant.ErrMsgRoleNotFoundInToken)
			}

			// Check if this role is super admin - if yes, bypass all permission checking
			isSuperAdmin, err := m.authorizationService.IsRoleSuperAdmin(c.Request().Context(), roleID)
			if err != nil {
				c.Logger().Errorf("failed to check super admin status: %v", err)
				return echo.NewHTTPError(http.StatusForbidden, constant.ErrMsgInsufficientPermissions)
			}
			if isSuperAdmin {
				// Super admin bypasses all permission checks including organization-scoped checks
				return next(c)
			}

			// Check if organization context is present
			orgIDValue := c.Get(constant.OrganizationIDKey)
			userIDValue := c.Get(constant.UserIDKey)

			// If both organization and user context are present, use organization-scoped checking
			if orgIDValue != nil && userIDValue != nil {
				organizationID, orgOk := orgIDValue.(uuid.UUID)
				userID, userOk := userIDValue.(uuid.UUID)

				if orgOk && userOk {
					hasPermission, err := m.authorizationService.CheckPermissionInOrganization(c.Request().Context(), userID, organizationID, permissionName)
					if err != nil {
						c.Logger().Errorf("organization-scoped permission check failed: %v", err)
						return echo.NewHTTPError(http.StatusForbidden, constant.ErrMsgInsufficientPermissions)
					}
					if !hasPermission {
						return echo.NewHTTPError(http.StatusForbidden, constant.ErrMsgInsufficientPermissions)
					}
					return next(c)
				}
			}

			// Fallback to traditional role-based permission checking
			// roleID already declared above, no need to redeclare

			// Use the AuthorizationService to check for permission.
			hasPermission, err := m.authorizationService.CheckPermission(c.Request().Context(), roleID, permissionName)
			if err != nil {
				// Log the internal error and return a generic message
				c.Logger().Errorf("permission check failed: %v", err)
				return echo.NewHTTPError(http.StatusForbidden, constant.ErrMsgInsufficientPermissions)
			}
			if !hasPermission {
				return echo.NewHTTPError(http.StatusForbidden, constant.ErrMsgInsufficientPermissions)
			}

			// If all checks pass, proceed to the next handler
			return next(c)
		}
	}
}

// extractOrganizationID is a helper method to extract organization ID from JWT claims or headers
func (m *Middleware) extractOrganizationID(c echo.Context) (uuid.UUID, bool) {
	// Try to get organization ID from JWT claims first
	if orgIDValue := c.Get(constant.OrganizationIDKey); orgIDValue != nil {
		if orgID, ok := orgIDValue.(uuid.UUID); ok {
			return orgID, true
		}
	}

	// If not in JWT claims, try to get from X-Organization-ID header
	if orgIDHeader := c.Request().Header.Get("X-Organization-ID"); orgIDHeader != "" {
		if orgID, err := uuid.Parse(orgIDHeader); err == nil {
			return orgID, true
		}
	}

	return uuid.UUID{}, false
}
