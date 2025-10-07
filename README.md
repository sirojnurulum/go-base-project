# Go Base Project - Auth & RBAC Template

A production-ready Go backend template with comprehensive authentication and role-based access control (RBAC) features. This template provides a solid foundation for building scalable backend applications.

## Features

### 🔐 Authentication & Authorization
- **JWT Token Authentication** with refresh token support
- **Google OAuth 2.0** integration
- **Multi-organization support** with context switching
- **Hierarchical RBAC** system with granular permissions
- **Permission-based middleware** for route protection

### 🏗️ Architecture
- **Clean Architecture** with clear separation of concerns
- **Repository Pattern** for data access abstraction
- **Dependency Injection** for loose coupling
- **Middleware-based** request processing
- **Comprehensive logging** with structured JSON output

### 🚀 Infrastructure
- **Redis** for caching and session management
- **PostgreSQL** with GORM for database operations
- **Rate limiting** with configurable storage backends
- **Health checks** (public and private endpoints)
- **Prometheus metrics** integration
- **OpenTelemetry** tracing support
- **Swagger/OpenAPI** documentation

### 🛡️ Security
- **Security headers** middleware
- **CORS** configuration
- **Request ID** tracking
- **Input validation** and sanitization
- **SQL injection** protection via GORM

## Quick Start

### Prerequisites
- Go 1.24.5 or higher
- PostgreSQL database
- Redis server (optional, for caching and rate limiting)

### Environment Setup

1. Copy the environment example:
```bash
cp .env.example .env
```

2. Configure your environment variables:
```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=your_db_user
DB_PASSWORD=your_db_password
DB_NAME=your_db_name
DB_SSL_MODE=disable

# Redis Configuration (optional)
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-here
JWT_ACCESS_TOKEN_EXPIRE_MINUTES=60
JWT_REFRESH_TOKEN_EXPIRE_DAYS=7

# Google OAuth (optional)
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8080/api/auth/google/callback

# Server Configuration
PORT=8080
ENV=development
FRONTEND_URL=http://localhost:3000

# Rate Limiting
RATE_LIMIT_RPS=100
RATE_LIMIT_BURST=200
RATE_LIMIT_STORAGE=memory  # or "redis"

# Security
ENABLE_SECURITY_HEADERS=true
```

### Database Setup

1. Create the database:
```sql
CREATE DATABASE your_db_name;
```

2. Run migrations:
```bash
# Apply migrations in order
psql -h localhost -U your_db_user -d your_db_name -f migrations/001_create_extensions_and_functions.sql
psql -h localhost -U your_db_user -d your_db_name -f migrations/002_create_rbac_and_organization_tables.sql
psql -h localhost -U your_db_user -d your_db_name -f migrations/003_create_users_and_user_organization_tables.sql
```

3. Seed initial data (optional):
```bash
go run cmd/api/main.go --seed
```

### Running the Application

```bash
# Install dependencies
go mod tidy

# Build the application
go build -o bin/main ./cmd/api

# Run the application
./bin/main
```

The server will start on `http://localhost:8080` (or your configured PORT).

## API Documentation

Once the server is running, access the Swagger documentation at:
- **Swagger UI**: `http://localhost:8080/swagger/index.html`

## Project Structure

```
.
├── cmd/
│   └── api/
│       └── main.go                 # Application entry point
├── internal/
│   ├── app/
│   │   └── app.go                  # Application initialization
│   ├── bootstrap/                  # Dependency injection setup
│   │   ├── handlers.go
│   │   ├── repositories.go
│   │   └── services.go
│   ├── config/
│   │   └── config.go               # Configuration management
│   ├── constant/                   # Application constants
│   ├── dto/                        # Data Transfer Objects
│   ├── handler/                    # HTTP request handlers
│   ├── middleware/                 # Custom middleware
│   ├── model/                      # Database models
│   ├── repository/                 # Data access layer
│   ├── router/
│   │   └── router.go               # Route definitions
│   ├── seeder/                     # Database seeders
│   ├── service/                    # Business logic layer
│   ├── util/                       # Utility functions
│   └── validator/                  # Custom validators
├── migrations/                     # Database migrations
├── platform/                      # External platform integrations
│   ├── database/
│   ├── logger/
│   └── redis/
└── docs/                          # Swagger documentation
```

## Key Components

### Authentication Flow

1. **Login**: POST `/api/auth/login`
   - Email/password authentication
   - Returns access and refresh tokens

2. **Google OAuth**: GET `/api/auth/google/login`
   - Redirects to Google OAuth consent
   - Callback: GET `/api/auth/google/callback`

3. **Token Refresh**: POST `/api/auth/refresh`
   - Renew access token using refresh token

4. **Current User**: GET `/api/auth/me`
   - Get authenticated user information

### RBAC System

#### Permissions
- Granular permissions like `users:read`, `users:create`, `roles:assign`
- Hierarchical organization-based permissions
- Platform-level vs organization-level access

#### Roles
- Predefined roles with permission sets
- Custom role creation for organizations
- Role assignment within organization context

#### Organizations
- Multi-tenant organization support
- User membership with organization-specific roles
- Organization context switching

### Middleware

- **JWT Authentication**: Validates and extracts user from JWT tokens
- **Permission Check**: Enforces permission-based access control
- **Organization Context**: Extracts organization context from routes
- **Rate Limiting**: Configurable rate limiting with Redis/memory storage
- **Security Headers**: Adds security headers to responses
- **Request Logging**: Structured logging with request tracking

## Usage Examples

### Adding New Features

1. **Create a new model** in `internal/model/`
2. **Define DTOs** in `internal/dto/`
3. **Implement repository** in `internal/repository/`
4. **Create service** in `internal/service/`
5. **Add handler** in `internal/handler/`
6. **Register routes** in `internal/router/router.go`
7. **Update bootstrap** files to wire dependencies

### Adding Custom Permissions

```go
// In your service or handler
if !authService.HasPermission(userID, "your-feature:action") {
    return echo.NewHTTPError(http.StatusForbidden, "Insufficient permissions")
}
```

### Organization-Scoped Resources

```go
// Use organization context middleware
orgRoutes := api.Group("/organizations/:orgId", m.JWT, m.OrganizationContext())
orgRoutes.GET("/resources", handler.GetOrgResources, m.RequirePermission("resources:read"))
```

## Development

### Testing

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...
```

### Building for Production

```bash
# Build for Linux
GOOS=linux GOARCH=amd64 go build -o bin/main-linux ./cmd/api

# Build with optimizations
go build -ldflags="-s -w" -o bin/main ./cmd/api
```

## Customization

This template is designed to be easily customizable:

1. **Remove unused features**: Delete components you don't need
2. **Add business logic**: Extend the service layer with your domain logic
3. **Modify auth flow**: Customize authentication to fit your requirements
4. **Extend RBAC**: Add more permission granularity as needed
5. **Add integrations**: Extend platform layer for external services

## License

This project is licensed under the MIT License - see the LICENSE file for details.
