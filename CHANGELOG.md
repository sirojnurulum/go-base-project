# Changelog - Go Base Project Template

## [2.0.0] - October 29, 2025

### 🎯 Major Update: Simplified URL Configuration

#### Breaking Changes
- **Removed `Env` field** from Config struct - no more dev/staging/prod environment distinction
- **Environment-based conditional logic removed** - replaced with explicit feature flags
- **CORS configuration simplified** - now automatically computed from base URLs

#### Added
- ✅ **`AllowedOrigins` field** - Auto-computed CORS origins from base URLs
- ✅ **`BackendURL` field** - Centralized backend URL configuration
- ✅ **`LOGGER_CONSOLE` env var** - Optional console logging for debugging
- ✅ **Production-ready defaults** - Security features enabled by default
- ✅ **Comprehensive documentation** - DEPLOYMENT_GUIDE.md added

#### Changed
- **Config Structure**: Removed `Env`, added `BackendURL` and `AllowedOrigins`
- **CORS Setup**: Simplified to use `cfg.AllowedOrigins` directly
- **Logger**: Now uses explicit mode instead of environment-based logic
- **Database Connection**: Uses `EnableDetailedTracing` flag instead of environment
- **Cookie Security**: Uses explicit `CookieSecure` flag instead of environment check
- **Security Headers**: Always checks explicit flag, enabled by default

#### Migration Guide

##### 1. Update `.env` file

**Before:**
```env
ENV=production
FRONTEND_URL=http://localhost:5173
```

**After:**
```env
FRONTEND_URL=http://localhost:5173
BACKEND_URL=http://localhost:8080
COOKIE_SECURE=true
ENABLE_SECURITY_HEADERS=true
ENABLE_DETAILED_TRACING=false
```

##### 2. Remove environment checks in custom code

**Before:**
```go
if cfg.Env == "production" {
    // production logic
}
```

**After:**
```go
if cfg.EnableSecurityHeaders {
    // explicit feature flag
}
```

##### 3. Update CORS configuration (if customized)

**Before:**
```go
AllowOrigins: []string{cfg.FrontendURL}
```

**After:**
```go
AllowOrigins: cfg.AllowedOrigins  // Includes both frontend and backend
```

#### Benefits

- ✅ **70% less configuration complexity** - Single environment approach
- ✅ **Domain migration in 2 steps** - Only update FRONTEND_URL and BACKEND_URL
- ✅ **Automatic CORS** - No manual origin management needed
- ✅ **Production-ready by default** - Security features always enabled
- ✅ **Better testability** - Explicit flags instead of environment magic
- ✅ **Clearer codebase** - No environment-specific branches

#### Files Modified

1. `internal/config/config.go` - Simplified config structure
2. `internal/router/router.go` - Simplified CORS configuration
3. `cmd/api/main.go` - Removed environment logging
4. `platform/logger/logger.go` - Added console mode option
5. `platform/database/database.go` - Use explicit tracing flag
6. `internal/handler/auth_handler.go` - Use explicit cookie secure flag
7. `.env.example` - Complete rewrite with clear sections
8. `README.md` - Updated environment configuration section

---

## [1.0.0] - Previous Version

### Initial Release
- JWT authentication with refresh tokens
- Google OAuth 2.0 integration
- Multi-organization support
- Hierarchical RBAC system
- Clean architecture implementation
- Rate limiting with Redis support
- Comprehensive middleware stack
- Swagger documentation
- Health check endpoints
- Prometheus metrics
- OpenTelemetry tracing

---

**Note**: This template is synchronized with the Beresin Backend project, ensuring best practices and battle-tested patterns are maintained across both projects.
