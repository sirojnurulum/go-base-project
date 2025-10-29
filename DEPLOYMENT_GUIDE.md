# 🚀 Go Base Project - Deployment Guide

## Quick Domain Change Guide

### When You Get a New Domain

Only **2 environment variables** need to be changed:

```bash
# In your .env file
FRONTEND_URL=https://app.yourdomain.com
BACKEND_URL=https://api.yourdomain.com
```

**That's it!** Everything else (CORS, OAuth callbacks, Swagger docs) will automatically use these URLs.

---

## Step-by-Step Deployment

### 1. Prepare Environment File

```bash
cp .env.example .env
```

### 2. Update URLs

```bash
# Edit .env
FRONTEND_URL=https://app.yourproject.com
BACKEND_URL=https://api.yourproject.com
```

### 3. Configure Database

```bash
DATABASE_URL=postgres://user:pass@your-db-host:5432/yourproject_db?sslmode=require
```

### 4. Set Security Credentials

```bash
# Generate strong JWT secret
JWT_SECRET=$(openssl rand -base64 32)

# Set admin password
ADMIN_DEFAULT_PASSWORD=your-secure-password
```

### 5. Configure Google OAuth (Optional)

```bash
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret
```

**Important**: Update Google Console redirect URI to:
```
https://api.yourproject.com/api/auth/google/callback
```

### 6. Enable HTTPS Settings

```bash
COOKIE_SECURE=true
ENABLE_SECURITY_HEADERS=true
```

### 7. Build and Run

```bash
make build
./bin/main
```

Or with Docker:

```bash
docker-compose up -d
```

---

## Domain Migration Checklist

When migrating to a new domain:

- [ ] Update `FRONTEND_URL` in `.env`
- [ ] Update `BACKEND_URL` in `.env`
- [ ] Update Google OAuth redirect URIs (if using OAuth)
- [ ] Update DNS records
- [ ] Update SSL certificates
- [ ] Restart backend service
- [ ] Test CORS from frontend
- [ ] Test OAuth login flow
- [ ] Verify Swagger docs accessible

---

## Common Deployment Scenarios

### Scenario 1: Local Development

```env
FRONTEND_URL=http://localhost:5173
BACKEND_URL=http://localhost:8080
COOKIE_SECURE=false
LOGGER_CONSOLE=true
```

### Scenario 2: Staging with IP

```env
FRONTEND_URL=http://staging-ip:5173
BACKEND_URL=http://staging-ip:8080
COOKIE_SECURE=false
ENABLE_DETAILED_TRACING=true
```

### Scenario 3: Production with Domain

```env
FRONTEND_URL=https://app.yourproject.com
BACKEND_URL=https://api.yourproject.com
COOKIE_SECURE=true
ENABLE_SECURITY_HEADERS=true
ENABLE_DETAILED_TRACING=false
RATE_LIMIT_STORAGE=redis
```

### Scenario 4: Production with Subdomain

```env
FRONTEND_URL=https://yourproject.com
BACKEND_URL=https://api.yourproject.com
COOKIE_SECURE=true
ENABLE_SECURITY_HEADERS=true
```

---

## Troubleshooting

### CORS Errors

**Problem**: Frontend gets CORS error

**Check**:
1. Is `FRONTEND_URL` exactly matching the frontend origin?
2. Check browser console for the exact origin being sent
3. Verify backend logs show correct `AllowedOrigins`

**Fix**:
```bash
# Make sure URLs match exactly (including protocol and port)
FRONTEND_URL=https://app.yourproject.com  # No trailing slash
```

### OAuth Redirect Fails

**Problem**: Google OAuth returns error after login

**Check**:
1. Google Console redirect URI matches `{BACKEND_URL}/api/auth/google/callback`
2. `BACKEND_URL` is accessible from internet
3. OAuth credentials are correct

**Fix**:
```bash
# Update in Google Console:
Redirect URI: https://api.yourproject.com/api/auth/google/callback

# Update in .env:
BACKEND_URL=https://api.yourproject.com
GOOGLE_CLIENT_ID=your-id
GOOGLE_CLIENT_SECRET=your-secret
```

### Cookie Not Being Set

**Problem**: Authentication cookie not working

**Check**:
1. If using HTTPS, `COOKIE_SECURE=true`
2. If using HTTP (dev), `COOKIE_SECURE=false`
3. `COOKIE_SAME_SITE` setting

**Fix**:
```bash
# For HTTPS production
COOKIE_SECURE=true
COOKIE_SAME_SITE=lax

# For HTTP development
COOKIE_SECURE=false
COOKIE_SAME_SITE=lax
```

---

## Health Check

After deployment, verify:

```bash
# Public health check
curl https://api.yourproject.com/api/health/public

# Swagger docs
open https://api.yourproject.com/swagger/index.html

# CORS test from browser console (on frontend page)
fetch('https://api.yourproject.com/api/health/public', {
  method: 'GET',
  credentials: 'include'
}).then(r => r.json()).then(console.log)
```

---

## Environment Variables Reference

### Required
- `DATABASE_URL` - PostgreSQL connection string
- `JWT_SECRET` - Secret key for JWT tokens
- `FRONTEND_URL` - Frontend application URL
- `BACKEND_URL` - Backend API URL

### Optional
- `PORT` - Server port (default: 8080)
- `REDIS_URL` - Redis connection string
- `GOOGLE_CLIENT_ID` - Google OAuth client ID
- `GOOGLE_CLIENT_SECRET` - Google OAuth client secret
- `ADMIN_DEFAULT_PASSWORD` - Initial admin password
- `COOKIE_SECURE` - Cookie secure flag
- `ENABLE_SECURITY_HEADERS` - Enable security headers
- `ENABLE_DETAILED_TRACING` - Enable detailed tracing
- `LOGGER_CONSOLE` - Use console logger
- `RATE_LIMIT_RPS` - Rate limit requests per second
- `RATE_LIMIT_BURST` - Rate limit burst size
- `RATE_LIMIT_STORAGE` - Rate limit storage (memory/redis)

---

**Last Updated**: October 29, 2025  
**Version**: 1.0  
**Status**: Production-Ready ✅
