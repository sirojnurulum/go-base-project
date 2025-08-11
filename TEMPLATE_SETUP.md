# Template Setup Guide

## 🚀 Quick Template Customization

After creating a new repository from this template, follow these steps to customize it for your project:

### 1. Project Rename

Replace all instances of "beresin-backend" with your project name:

```bash
# Replace in go.mod
sed -i '' 's/beresin-backend/your-project-name/g' go.mod

# Replace in all Go files
find . -name "*.go" -type f -exec sed -i '' 's/beresin-backend/your-project-name/g' {} +

# Replace in docker files
sed -i '' 's/beresin-backend/your-project-name/g' Dockerfile
sed -i '' 's/beresin-backend/your-project-name/g' docker-compose.yml

# Replace in Makefile
sed -i '' 's/beresin-backend/your-project-name/g' Makefile
```

### 2. Environment Configuration

```bash
# Copy example environment file
cp .env.example .env

# Update with your values
APP_NAME=your-project-name
APP_ENV=development
DB_NAME=your_database
JWT_SECRET=your-super-secret-key-here
```

### 3. Database Setup

```bash
# Update database name in .env
DB_NAME=your_project_db

# Create database
createdb your_project_db

# Run migrations
make migrate-up

# Seed initial data
make seed
```

### 4. Google OAuth Setup (Optional)

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select existing
3. Enable Google+ API
4. Create OAuth 2.0 credentials
5. Add to .env:
   ```
   GOOGLE_CLIENT_ID=your-client-id
   GOOGLE_CLIENT_SECRET=your-client-secret
   GOOGLE_REDIRECT_URL=http://localhost:8080/api/auth/google/callback
   ```

### 5. First Run

```bash
# Install development tools
make dev-setup

# Start development server
make dev
```

### 6. Verify Installation

- API: http://localhost:8080/api/health/public
- Swagger: http://localhost:8080/swagger/index.html
- Admin login: admin@example.com / password123

## 🔧 Template Features Included

✅ JWT Authentication with Google OAuth  
✅ Role-Based Access Control (RBAC)  
✅ Organization Management System  
✅ User Management with Security Validations  
✅ Database Migrations with Goose  
✅ API Documentation with Swagger  
✅ Structured Logging with Zerolog  
✅ Hot Reload Development Environment  
✅ Health Checks & Monitoring  
✅ Redis Caching Layer  
✅ Rate Limiting & Security Headers  
✅ Clean Architecture Pattern  
✅ Comprehensive Error Handling  

## 📝 Next Steps

1. Customize business logic in `/internal/service/`
2. Add your domain models to `/internal/model/`
3. Create API endpoints in `/internal/handler/`
4. Add database migrations in `/migrations/`
5. Update API documentation
6. Write tests for your features

## 🤝 Support

If you need help customizing this template:
1. Check the main README.md
2. Review the API documentation at `/swagger`
3. Look at existing examples in the codebase
4. Open an issue on the template repository

Happy coding! 🚀
