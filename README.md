# Go Enterprise Backend Template

A production-ready Go backend template built with clean architecture, comprehensive security, and enterprise-grade features. This template provides a solid foundation for building scalable web applications with authentication, authorization, and organization management.

## 🚀 GitHub Template

This repository is a GitHub template. Click the "Use this template" button to create a new repository with the same structure and configurations.

## ✨ Features

### Core Features
- **Authentication & Authorization** - JWT-based auth with Google OAuth integration
- **Role-Based Access Control (RBAC)** - Hierarchical permission system with role inheritance
- **Organization Management** - Multi-tenant architecture with hierarchical organizations
- **User Management** - Complete user lifecycle with security validations
- **Security Hardening** - Enterprise-level security headers, rate limiting, and input validation

### Technical Features  
- **Clean Architecture** - Layered architecture with proper separation of concerns
- **Database Migrations** - Schema versioning with Goose
- **API Documentation** - Auto-generated Swagger/OpenAPI documentation
- **Structured Logging** - JSON logging with request tracing
- **Hot Reload** - Development environment with Air
- **Health Checks** - Comprehensive health monitoring
- **Metrics** - Prometheus metrics integration
- **Caching** - Redis-based caching layer

## 🛠 Tech Stack

| Category | Technology |
|----------|------------|
| **Language** | Go 1.21+ |
| **Framework** | Echo v4 |
| **Database** | PostgreSQL with GORM |
| **Cache** | Redis |
| **Authentication** | JWT + Google OAuth |
| **Migrations** | Goose |
| **Documentation** | Swagger/OpenAPI |
| **Logging** | Zerolog |
| **Monitoring** | Prometheus |
| **Development** | Air (hot reload) |

## 🏗 Architecture

```
cmd/api/                 # Application entrypoint
internal/
├── app/                 # Application setup and lifecycle
├── bootstrap/           # Dependency injection setup
├── config/              # Configuration management
├── constant/            # Application constants
├── dto/                 # Data Transfer Objects
├── handler/             # HTTP handlers (controllers)
├── middleware/          # Custom middleware
├── model/               # Domain models
├── repository/          # Data access layer
├── router/              # Route definitions
├── service/             # Business logic layer
├── seeder/              # Database seeders
├── util/                # Utility functions
└── validator/           # Request validation
migrations/              # Database migrations
pkg/                     # Shared packages
platform/               # External service connections
```

## 🚦 Getting Started

### Prerequisites

```bash
# Required installations
go version      # Go 1.21+
docker --version # Docker for PostgreSQL/Redis
make --version   # Make for task automation

# Development tools (auto-installed via Makefile)
air             # Hot reload
goose           # Database migrations  
swag            # API documentation
```

### Quick Start

1. **Create from template:**
   ```bash
   # Click "Use this template" on GitHub or
   git clone https://github.com/your-username/go-base-project.git
   cd go-base-project
   ```

2. **Setup environment:**
   ```bash
   cp .env.example .env
   # Edit .env with your configurations
   ```

3. **Start development:**
   ```bash
   make dev-setup    # Install tools & start services
   make migrate-up   # Run database migrations
   make seed         # Seed initial data
   make dev          # Start development server
   ```

4. **Access the application:**
   - API: http://localhost:8080
   - Swagger: http://localhost:8080/swagger/index.html
   - Health: http://localhost:8080/api/health/public

## 📝 Environment Configuration

Key environment variables:

```bash
# Server
APP_NAME=your-app-name
APP_ENV=development
APP_PORT=8080
APP_TIMEOUT=30

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your-password
DB_NAME=your-database

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# JWT
JWT_SECRET=your-super-secret-key

# Google OAuth (optional)
GOOGLE_CLIENT_ID=your-client-id
GOOGLE_CLIENT_SECRET=your-secret

# Frontend (CORS)
FRONTEND_URL=http://localhost:3000
```

## 🔧 Available Commands

```bash
# Development
make dev            # Start development server with hot reload
make dev-setup      # Install development tools
make build          # Build production binary
make clean          # Clean build artifacts

# Database
make db-up          # Start PostgreSQL & Redis
make db-down        # Stop database services
make migrate-up     # Run all migrations
make migrate-down   # Rollback last migration
make migrate-reset  # Reset database
make seed           # Seed initial data

# Documentation
make docs           # Generate Swagger documentation
make docs-serve     # Serve documentation locally

# Quality
make test           # Run all tests
make test-coverage  # Run tests with coverage
make lint           # Run golangci-lint
make format         # Format code

# Docker
make docker-build   # Build Docker image
make docker-run     # Run in Docker
```

## 🔐 Authentication & Authorization

### User Roles Hierarchy

```
Super Admin (Level 0)
├── Platform Admin (Level 1)  
│   ├── Company Admin (Level 2)
│   │   └── Store Manager (Level 3)
│   └── Company User (Level 2)
└── Platform User (Level 1)
```

### API Authentication

```bash
# Login
POST /api/auth/login
{
  "email": "admin@example.com", 
  "password": "password123"
}

# Use JWT token in subsequent requests
Authorization: Bearer <jwt-token>
```

### Permission System

Permissions follow the pattern: `resource:action`

```
users:create, users:read, users:update, users:delete
roles:create, roles:assign
organizations:create, organizations:manage
dashboard:view
```

## 🏢 Organization Management

### Organization Types

- **Platform** - Top-level organization
- **Company** - Under platform
- **Store** - Under company

### Hierarchy Rules

- Users can belong to multiple organizations
- Permissions inherit down the hierarchy
- Company admins can manage store-level users
- Platform admins can manage company-level users

### API Examples

```bash
# Create organization
POST /api/admin/organizations
{
  "name": "ACME Corp",
  "type": "company", 
  "parent_id": "platform-uuid"
}

# Join organization
POST /api/admin/organizations/{id}/join
```

## 📊 Monitoring & Observability

### Health Checks

- `GET /api/health/public` - Basic health status
- `GET /api/health/private` - Detailed health with database/redis status

### Metrics

Prometheus metrics available at `/metrics`:
- HTTP request duration and count
- Database connection pool stats  
- Redis operations metrics
- Custom business metrics

### Logging

Structured JSON logging with request tracing:
```json
{
  "level": "info",
  "time": "2024-01-01T10:00:00Z",
  "request_id": "uuid-here",
  "method": "POST",
  "path": "/api/users",
  "status": 201,
  "duration": 45.2
}
```

## 🛡️ Security Features

- **Rate Limiting** - Redis-based with configurable limits
- **Security Headers** - HSTS, CSP, X-Frame-Options, etc.
- **Input Validation** - Comprehensive request validation
- **SQL Injection Prevention** - Parameterized queries with GORM
- **XSS Protection** - Output encoding and CSP
- **CORS Configuration** - Configurable cross-origin policies

## 🧪 Testing

```bash
make test              # Run all tests
make test-coverage     # With coverage report
make test-integration  # Integration tests
make test-e2e         # End-to-end tests
```

Test structure:
```
internal/
├── handler/
│   ├── user_handler.go
│   └── user_handler_test.go
├── service/  
│   ├── user_service.go
│   └── user_service_test.go
└── repository/
    ├── user_repository.go
    └── user_repository_test.go
```

## 📦 Deployment

### Docker Deployment

```bash
# Build image
make docker-build

# Run with docker-compose
docker-compose up -d
```

### Production Checklist

- [ ] Environment variables configured
- [ ] Database migrations applied
- [ ] SSL/TLS certificates installed
- [ ] Reverse proxy configured (Nginx/Traefik)
- [ ] Monitoring setup (Prometheus/Grafana)
- [ ] Log aggregation configured
- [ ] Backup strategy implemented
- [ ] Security scan completed

## 🤝 Contributing

1. Fork the repository
2. Create feature branch: `git checkout -b feature/amazing-feature`
3. Commit changes: `git commit -m 'Add amazing feature'`
4. Push to branch: `git push origin feature/amazing-feature`
5. Open Pull Request

### Coding Standards

- Follow Go best practices
- Write tests for new features
- Update documentation
- Use conventional commits
- Run `make lint` before committing

## 📋 API Documentation

Full API documentation is available at `/swagger/index.html` when running the server.

### Key Endpoints

| Endpoint | Method | Description |
|----------|---------|-------------|
| `/api/auth/login` | POST | User login |
| `/api/auth/refresh` | POST | Refresh JWT token |
| `/api/admin/users` | GET,POST,PUT,DELETE | User management |
| `/api/admin/roles` | GET,POST,PUT | Role management |  
| `/api/admin/organizations` | GET,POST,PUT,DELETE | Organization management |

## 🐛 Troubleshooting

### Common Issues

**Database connection failed**
```bash
# Check if PostgreSQL is running
make db-up
# Verify connection string in .env
```

**Migration errors**
```bash 
# Reset and rerun migrations
make migrate-reset
make migrate-up
```

**Port already in use**
```bash
# Kill process on port 8080
lsof -ti:8080 | xargs kill -9
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Echo Framework](https://echo.labstack.com/) - High performance Go web framework
- [GORM](https://gorm.io/) - Fantastic ORM library for Go
- [Goose](https://github.com/pressly/goose) - Database migration tool
- [Air](https://github.com/cosmtrek/air) - Live reload for Go apps

---

**Built with ❤️ using Go**
