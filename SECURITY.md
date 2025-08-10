# Security Implementation Guide

## Rate Limiting

### Overview
This project implements comprehensive rate limiting with two approaches:
- **Development**: In-memory rate limiting using `golang.org/x/time/rate`
- **Production**: Redis-based rate limiting with sliding window algorithm

### Configuration
Set these environment variables to configure rate limiting:

```bash
ENABLE_RATE_LIMIT=true
API_REQUESTS_PER_MINUTE=1000    # General API endpoints
AUTH_REQUESTS_PER_MINUTE=10     # Authentication endpoints
```

### Rate Limit Tiers

#### 1. Authentication Endpoints (`/auth/*`)
- **Development**: 5 requests per minute (burst: 5)
- **Production**: 10 requests per 5 minutes (Redis-based)
- **Purpose**: Prevent brute force attacks on login

#### 2. General API Endpoints (`/api/*`)
- **Development**: 100 requests per minute (burst: 20)  
- **Production**: 1000 requests per 15 minutes (Redis-based)
- **Purpose**: Prevent API abuse and DoS attacks

#### 3. Health Check Endpoints
- **Rate Limiting**: Lighter limits for monitoring services
- **Excluded**: `/metrics` and `/swagger/*` for operational needs

### Response Format
When rate limit is exceeded:
```json
{
    "error": "Too Many Requests",
    "message": "Rate limit exceeded. Please try again later.",
    "retry_after": 300
}
```

## Security Headers

### Implemented Headers

#### 1. XSS Protection
```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
```

#### 2. Content Security Policy
```
Content-Security-Policy: default-src 'self'; connect-src 'self' FRONTEND_URL; img-src 'self' data: https:; font-src 'self' data:; style-src 'self' 'unsafe-inline'; script-src 'self'
```

#### 3. HTTPS and Transport Security
```
Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
```
*Applied only in production or when X-Forwarded-Proto: https*

#### 4. Privacy and Permissions
```
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: geolocation=(), microphone=(), camera=(), payment=()
```

#### 5. Information Disclosure Prevention
```
Server: (removed)
X-Powered-By: (removed)
```

### Environment-Specific Behavior

#### Development
- Basic security headers applied
- CSP allows more flexibility for development
- HSTS not applied (no HTTPS requirement)

#### Production
- Strict security headers
- HSTS enforced
- CSP includes `upgrade-insecure-requests`
- Full privacy policies applied

## Implementation Details

### Rate Limiting Architecture

#### In-Memory (Development)
- Uses `golang.org/x/time/rate` token bucket algorithm
- Automatic cleanup of inactive IPs every 3 minutes
- Thread-safe with mutex protection
- Graceful degradation if memory pressure

#### Redis-Based (Production)
- Sliding window algorithm using Redis Sorted Sets
- Distributed rate limiting across multiple instances  
- Automatic cleanup of expired entries
- Fail-open approach (allow requests if Redis is down)
- Pipeline operations for atomic updates

### Security Headers Implementation
- Applied globally to all routes
- Environment-aware configuration
- CSP dynamically includes frontend URL
- Headers optimized for modern browsers

### Monitoring and Observability
- Rate limit violations logged with IP and endpoint
- Redis failures logged but don't block requests
- Prometheus metrics available at `/metrics`
- Request tracing with X-Request-ID

## Best Practices

### Rate Limiting
1. **Fail Open**: Always allow requests if rate limiting fails
2. **Per-IP Tracking**: Use real client IP (consider proxy headers)
3. **Different Limits**: Authentication endpoints have stricter limits
4. **Cleanup**: Automatic cleanup prevents memory/storage leaks
5. **Monitoring**: Log rate limit violations for analysis

### Security Headers
1. **Environment-Aware**: Stricter in production
2. **CSP Gradual**: Start permissive, tighten over time
3. **HSTS Preload**: Consider HSTS preload list for production domains
4. **Regular Updates**: Review and update security policies

### Configuration
1. **Environment Variables**: All security settings configurable
2. **Sensible Defaults**: Work out-of-the-box with reasonable security
3. **Documentation**: Clear configuration options
4. **Testing**: Test rate limits in staging environment

## Testing Rate Limits

### Development Testing
```bash
# Test auth rate limit (should block after 5 requests/minute)
for i in {1..10}; do curl -X POST localhost:8080/api/auth/login; done

# Test API rate limit (should block after 100 requests/minute)  
for i in {1..150}; do curl localhost:8080/api/health/public; done
```

### Production Testing
Use tools like `ab` or `wrk` to test production rate limits:
```bash
# Apache Bench example
ab -n 100 -c 10 https://yourapi.com/api/auth/login

# wrk example  
wrk -t12 -c400 -d30s https://yourapi.com/api/health/public
```

## Security Checklist

- [ ] Rate limiting enabled and tested
- [ ] Security headers verified with online tools
- [ ] CSP policy tested with your frontend
- [ ] HTTPS enforced in production
- [ ] Rate limit monitoring alerts configured
- [ ] Redis rate limiting tested in production
- [ ] Backup rate limiting (in-memory) tested
- [ ] Security headers tested across all endpoints
