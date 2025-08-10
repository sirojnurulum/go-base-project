# Beresin Backend - Project Context

## Project Overview
- **Name**: Beresin Backend
- **Type**: Go REST API Backend with RBAC + Advanced Security
- **Status**: Production-ready enterprise template
- **Primary Use**: Secure, scalable backend foundation with comprehensive rate limiting
- **Last Major Update**: August 10, 2025 - Advanced security implementation

## Architecture Summary
- **Pattern**: Clean Architecture (Handler -> Service -> Repository)
- **Auth**: JWT with refresh token rotation + comprehensive rate limiting
- **Database**: PostgreSQL with GORM + UUIDv7 optimization
- **Cache**: Redis for sessions, permissions + distributed rate limiting
- **Framework**: Echo v4 with advanced security middleware
- **Migration**: Goose with automatic execution
- **Security**: Multi-layer protection (rate limiting + headers + RBAC)

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
- [x] **Advanced dual-strategy rate limiting** (in-memory + Redis sliding window)
- [x] **Comprehensive security headers** (XSS, CSP, HSTS, privacy policies)
- [x] **Environment-aware security** (dev vs prod configurations)
- [x] **Enhanced health checks** with dependency validation
- [x] **Full containerization** (Docker + docker-compose)
- [x] **Production security documentation** (SECURITY.md)

## Technical Decisions Made
- **UUID Strategy**: UUIDv7 for primary keys (time-ordered performance)
- **Token Strategy**: Short-lived access (15min) + long-lived refresh (7d) with rotation
- **Caching Strategy**: Redis for refresh tokens and permission caching
- **Logging Strategy**: Zerolog with context enrichment (request_id, user_id)
- **Testing Strategy**: Service layer focus with repository mocking + security middleware tests
- **Rate Limiting Strategy**: Dual approach - in-memory (dev) + Redis sliding window (prod)
- **Security Headers Strategy**: Environment-aware progressive enhancement (stricter in prod)
- **IP Detection Strategy**: Real IP with proxy header support for accurate rate limiting
- **Middleware Layering**: Security -> Rate Limit -> CORS -> JWT -> Permission checks
- **Fail-Safe Design**: Rate limiting and security fail-open to maintain availability

## Current TODO Status
- **Foundation**: ✅ Complete (clean architecture + DI)
- **Database**: ✅ Complete (migrations + pooling + UUIDv7)
- **Security**: ✅ Complete ⭐ (RBAC + JWT + rate limiting + security headers)
- **Observability**: 🔄 Partial (Zerolog + Prometheus, missing OpenTelemetry)
- **Testing**: ✅ Complete (handlers + services + middleware)
- **Deployment**: ✅ Complete (Docker + docker-compose)
- **Health Checks**: ✅ Enhanced (dependency validation + metrics)
- **Rate Limiting**: ✅ Complete ⭐ (dual strategy: memory + Redis)
- **Security Headers**: ✅ Complete ⭐ (comprehensive + environment-aware)
- **Documentation**: ✅ Complete (API docs + security guide)

## Next Steps
1. ✅ Complete Docker multi-stage builds
2. ✅ Add docker-compose for full development stack
3. ✅ Advanced rate limiting implementation 
4. ✅ Comprehensive security headers
5. Add OpenTelemetry for distributed tracing (optional for microservices)
6. Consider API versioning strategy (if scaling to multiple versions)
7. Add integration tests for security features
8. Consider adding Redis-based session management (if needed)

## Context for AI Assistant
This is a **mature, production-ready Go backend template** demonstrating enterprise best practices. When working on this project, focus on:
- Maintaining the established clean architecture patterns
- Following the existing security implementations  
- Preserving the comprehensive testing approach
- Keeping consistent with the structured logging patterns

**Do not suggest major architectural changes** - this is a stable, well-designed foundation.
