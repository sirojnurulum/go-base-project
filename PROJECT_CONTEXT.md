# Beresin Backend - Project Context

## Project Overview
- **Name**: Beresin Backend
- **Type**: Go REST API Backend with RBAC
- **Status**: Production-ready template
- **Primary Use**: Enterprise-grade backend foundation

## Architecture Summary
- **Pattern**: Clean Architecture (Handler -> Service -> Repository)
- **Auth**: JWT with refresh token rotation
- **Database**: PostgreSQL with GORM
- **Cache**: Redis for sessions and permissions
- **Framework**: Echo v4
- **Migration**: Goose

## Key Features Implemented
- [x] Complete RBAC system with granular permissions
- [x] JWT authentication with HttpOnly cookie refresh tokens  
- [x] UUIDv7 primary keys for time-ordered performance
- [x] Structured logging with Zerolog + request tracing
- [x] Prometheus metrics integration
- [x] Comprehensive testing with mocks
- [x] Hot-reload development with Air
- [x] Swagger API documentation
- [x] Database connection pooling
- [x] Graceful shutdown handling
- [x] **Advanced rate limiting** (in-memory + Redis-based)
- [x] **Comprehensive security headers** (XSS, CSP, HSTS, etc.)
- [x] **Environment-aware security** (dev vs prod configurations)

## Technical Decisions Made
- **UUID Strategy**: UUIDv7 for primary keys (time-ordered)
- **Token Strategy**: Short-lived access (15min) + long-lived refresh (7d) with rotation
- **Caching Strategy**: Redis for refresh tokens and permission caching
- **Logging Strategy**: Zerolog with context enrichment (request_id, user_id)
- **Testing Strategy**: Service layer focus with repository mocking
- **Rate Limiting Strategy**: Dual approach - in-memory for dev, Redis sliding window for production
- **Security Headers Strategy**: Environment-aware with progressive enhancement (stricter in prod)
- **IP Detection Strategy**: Real IP with proxy header support for rate limiting

## Current TODO Status
- **Foundation**: ✅ Complete
- **Database**: ✅ Complete  
- **Security**: ✅ Complete (with advanced rate limiting & security headers)
- **Observability**: 🔄 Partial (missing OpenTelemetry)
- **Testing**: ✅ Complete
- **Deployment**: ✅ Complete (Docker + docker-compose added)
- **Health Checks**: ✅ Enhanced with dependency validation
- **Rate Limiting**: ✅ Complete (dual strategy: in-memory + Redis)
- **Security Headers**: ✅ Complete (environment-aware, comprehensive)

## Next Steps
1. ✅ Complete Docker multi-stage builds
2. ✅ Add docker-compose for full development stack
3. Add OpenTelemetry for distributed tracing
4. Implement Redis-based rate limiting
5. Add integration tests
6. Consider adding API versioning strategy

## Context for AI Assistant
This is a **mature, production-ready Go backend template** demonstrating enterprise best practices. When working on this project, focus on:
- Maintaining the established clean architecture patterns
- Following the existing security implementations  
- Preserving the comprehensive testing approach
- Keeping consistent with the structured logging patterns

**Do not suggest major architectural changes** - this is a stable, well-designed foundation.
