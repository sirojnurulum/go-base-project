package middleware

import (
	"beresin-backend/internal/config"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthorizationService for testing
type MockAuthorizationService struct {
	mock.Mock
}

func (m *MockAuthorizationService) CheckPermission(roleID uuid.UUID, permission string) (bool, error) {
	args := m.Called(roleID, permission)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuthorizationService) GetAndCachePermissionsForRole(roleID uuid.UUID) ([]string, error) {
	args := m.Called(roleID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockAuthorizationService) InvalidateRolePermissionsCache(roleID uuid.UUID) error {
	args := m.Called(roleID)
	return args.Error(0)
}

func TestRateLimit(t *testing.T) {
	// Create middleware instance
	mockAuthService := &MockAuthorizationService{}
	middleware := NewMiddleware(mockAuthService, "test-secret")

	// Create echo instance
	e := echo.New()

	// Add rate limiting middleware (5 requests per minute, burst 2)
	rateLimitMiddleware := middleware.RateLimit(5.0/60.0, 2)

	// Test handler
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	t.Run("Allow requests within limit", func(t *testing.T) {
		// First request should pass
		req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req1.RemoteAddr = "192.168.1.1:12345"
		rec1 := httptest.NewRecorder()
		c1 := e.NewContext(req1, rec1)

		err := rateLimitMiddleware(handler)(c1)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec1.Code)

		// Second request should also pass (within burst)
		req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req2.RemoteAddr = "192.168.1.1:12345"
		rec2 := httptest.NewRecorder()
		c2 := e.NewContext(req2, rec2)

		err = rateLimitMiddleware(handler)(c2)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec2.Code)
	})

	t.Run("Block requests exceeding limit", func(t *testing.T) {
		// Create a fresh rate limiter for this test
		middleware := NewMiddleware(mockAuthService, "test-secret")
		rateLimitMiddleware := middleware.RateLimit(1.0/60.0, 1) // Very strict: 1 request per minute, burst 1

		// First request should pass
		req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req1.RemoteAddr = "192.168.1.2:12345"
		rec1 := httptest.NewRecorder()
		c1 := e.NewContext(req1, rec1)

		err := rateLimitMiddleware(handler)(c1)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec1.Code)

		// Second request should be blocked
		req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req2.RemoteAddr = "192.168.1.2:12345"
		rec2 := httptest.NewRecorder()
		c2 := e.NewContext(req2, rec2)

		err = rateLimitMiddleware(handler)(c2)

		// Should return an HTTP error
		if httpError, ok := err.(*echo.HTTPError); ok {
			assert.Equal(t, http.StatusTooManyRequests, httpError.Code)
		} else {
			t.Fatal("Expected HTTP error for rate limit exceeded")
		}
	})

	t.Run("Different IPs have separate limits", func(t *testing.T) {
		// Create a fresh rate limiter
		middleware := NewMiddleware(mockAuthService, "test-secret")
		rateLimitMiddleware := middleware.RateLimit(1.0/60.0, 1)

		// First IP - first request should pass
		req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req1.RemoteAddr = "192.168.1.3:12345"
		rec1 := httptest.NewRecorder()
		c1 := e.NewContext(req1, rec1)

		err := rateLimitMiddleware(handler)(c1)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec1.Code)

		// Different IP - first request should also pass
		req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
		req2.RemoteAddr = "192.168.1.4:12345"
		rec2 := httptest.NewRecorder()
		c2 := e.NewContext(req2, rec2)

		err = rateLimitMiddleware(handler)(c2)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec2.Code)
	})
}

func TestSecurityHeaders(t *testing.T) {
	// Create test config
	cfg := config.Config{
		Env:         "development",
		FrontendURL: "http://localhost:3000",
	}

	// Create middleware instance
	mockAuthService := &MockAuthorizationService{}
	middleware := NewMiddleware(mockAuthService, "test-secret")

	// Create echo instance
	e := echo.New()

	// Add security headers middleware
	securityMiddleware := middleware.SecurityHeaders(cfg)

	// Test handler
	handler := func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}

	t.Run("Security headers applied in development", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := securityMiddleware(handler)(c)
		assert.NoError(t, err)

		// Check security headers
		headers := rec.Header()
		assert.Equal(t, "nosniff", headers.Get("X-Content-Type-Options"))
		assert.Equal(t, "DENY", headers.Get("X-Frame-Options"))
		assert.Equal(t, "1; mode=block", headers.Get("X-XSS-Protection"))
		assert.Equal(t, "", headers.Get("Server"))
		assert.Equal(t, "", headers.Get("X-Powered-By"))
		assert.Contains(t, headers.Get("Content-Security-Policy"), "default-src 'self'")
		assert.Equal(t, "strict-origin-when-cross-origin", headers.Get("Referrer-Policy"))
		assert.Contains(t, headers.Get("Permissions-Policy"), "geolocation=()")

		// HSTS should not be set in development
		assert.Equal(t, "", headers.Get("Strict-Transport-Security"))
	})

	t.Run("HSTS headers applied in production", func(t *testing.T) {
		// Create production config
		prodCfg := config.Config{
			Env:         "production",
			FrontendURL: "https://myapp.com",
		}

		prodSecurityMiddleware := middleware.SecurityHeaders(prodCfg)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := prodSecurityMiddleware(handler)(c)
		assert.NoError(t, err)

		// Check HSTS header is set in production
		headers := rec.Header()
		assert.Equal(t, "max-age=31536000; includeSubDomains; preload", headers.Get("Strict-Transport-Security"))
		assert.Contains(t, headers.Get("Content-Security-Policy"), "upgrade-insecure-requests")
	})

	t.Run("HSTS headers applied with X-Forwarded-Proto", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-Forwarded-Proto", "https")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := securityMiddleware(handler)(c)
		assert.NoError(t, err)

		// Check HSTS header is set when behind HTTPS proxy
		headers := rec.Header()
		assert.Equal(t, "max-age=31536000; includeSubDomains; preload", headers.Get("Strict-Transport-Security"))
	})
}
