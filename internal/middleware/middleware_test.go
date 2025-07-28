package middleware_test

import (
	"beresin-backend/internal/constant"
	"beresin-backend/internal/middleware"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockAuthorizationService is a mock for the AuthorizationService interface.
type MockAuthorizationService struct {
	mock.Mock
}

func (m *MockAuthorizationService) CheckPermission(roleID uuid.UUID, requiredPermission string) (bool, error) {
	args := m.Called(roleID, requiredPermission)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuthorizationService) InvalidateRolePermissionsCache(roleID uuid.UUID) error {
	args := m.Called(roleID)
	return args.Error(0)
}

func (m *MockAuthorizationService) GetAndCachePermissionsForRole(roleID uuid.UUID) ([]string, error) {
	panic("not implemented") // We don't need this for the middleware test
}

// Helper to generate a test JWT
// This helper is now simplified as we only need a placeholder token.
// The actual claims are set directly in the context for middleware unit tests.
func generateTestToken(secret string) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   uuid.NewString(),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 1)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func TestRequirePermission(t *testing.T) {
	// Setup
	e := echo.New()

	// A simple handler that should be called on success
	successHandler := func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	t.Run("Success - Has Permission", func(t *testing.T) {
		mockAuthzService := new(MockAuthorizationService)
		testSecret := "test-secret"
		mw := middleware.NewMiddleware(mockAuthzService, testSecret)

		roleID := uuid.New()
		permission := "users:read"

		mockAuthzService.On("CheckPermission", roleID, permission).Return(true, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		// Simulate that the JWT middleware has run and set the role ID in the context.
		c.Set(constant.RoleIDKey, roleID)

		handler := mw.RequirePermission(permission)(successHandler)
		err := handler(c)

		// The handler should execute without error and the successHandler should be called.
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		mockAuthzService.AssertExpectations(t)
	})

	t.Run("Forbidden - Insufficient Permission", func(t *testing.T) {
		mockAuthzService := new(MockAuthorizationService)
		testSecret := "test-secret"
		mw := middleware.NewMiddleware(mockAuthzService, testSecret)

		roleID := uuid.New()
		permission := "users:delete"

		mockAuthzService.On("CheckPermission", roleID, permission).Return(false, nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		// Simulate that the JWT middleware has run and set the role ID in the context.
		c.Set(constant.RoleIDKey, roleID)

		handler := mw.RequirePermission(permission)(successHandler)
		err := handler(c)

		// The handler should return an error, not write to the response recorder.
		require.Error(t, err)
		httpErr, ok := err.(*echo.HTTPError)
		require.True(t, ok, "error should be an *echo.HTTPError")

		assert.Equal(t, http.StatusForbidden, httpErr.Code)
		assert.Equal(t, constant.ErrMsgInsufficientPermissions, httpErr.Message)
		mockAuthzService.AssertExpectations(t)
	})

	t.Run("Forbidden - No Role In Context", func(t *testing.T) {
		// This simulates a scenario where the JWT middleware ran but failed to set a role ID,
		// or the RequirePermission middleware is called without the JWT middleware.
		mockAuthzService := new(MockAuthorizationService)
		testSecret := "test-secret"
		mw := middleware.NewMiddleware(mockAuthzService, testSecret)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Note: We do NOT set the RoleIDKey in the context for this test.

		handler := mw.RequirePermission("any:permission")(successHandler)
		err := handler(c)

		// The handler should return an error.
		require.Error(t, err)
		httpErr, ok := err.(*echo.HTTPError)
		require.True(t, ok, "error should be an *echo.HTTPError")

		assert.Equal(t, http.StatusForbidden, httpErr.Code)
		assert.Equal(t, constant.ErrMsgRoleNotFoundInToken, httpErr.Message)

		// The authorization service should not be called if the role ID is missing.
		mockAuthzService.AssertNotCalled(t, "CheckPermission", mock.Anything, mock.Anything)
	})
}
