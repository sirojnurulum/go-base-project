package middleware

import (
	"beresin-backend/internal/constant"
	"beresin-backend/internal/service"
	"beresin-backend/internal/util"
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

		// Set user ID, role ID, and organization ID in the context for subsequent handlers.
		// Use constants for keys to maintain consistency.
		c.Set(constant.UserIDKey, claims.UserID)
		c.Set(constant.RoleIDKey, claims.RoleID)
		if claims.OrganizationID != nil {
			c.Set(constant.OrganizationIDKey, *claims.OrganizationID)
		}

		return next(c)
	}
}

// RequirePermission creates a middleware to check if a user has a given permission.
// This middleware now reads the roleID from the context, no longer parsing the token itself.
func (m *Middleware) RequirePermission(permissionName string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get the role ID from the context set by the JWT middleware
			roleIDValue := c.Get(constant.RoleIDKey)
			roleID, ok := roleIDValue.(uuid.UUID)
			if !ok {
				// This should not happen if the JWT middleware ran before
				return echo.NewHTTPError(http.StatusForbidden, constant.ErrMsgRoleNotFoundInToken)
			}

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
