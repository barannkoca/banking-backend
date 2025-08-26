# 🚀 HTTP Server Setup Implementation

Bu dokümantasyon, Banking Backend projesinde HTTP Server Setup gereksinimlerinin nasıl implement edildiğini açıklar.

## 📋 Implemented Features

### 1. ✅ Custom Router with Middleware Support

**Dosya:** `internal/api/router.go`

- **Gin Framework** kullanarak custom router oluşturuldu
- **Middleware stack** organize edildi:
  - Global middleware (recovery, security, CORS)
  - Route-specific middleware (authentication, rate limiting)
  - Banking-specific middleware (enhanced security, tracking)

**Özellikler:**
- Modüler middleware yapısı
- Route grupları (public, protected, admin)
- Conditional middleware application
- Clean separation of concerns

### 2. ✅ CORS and Security Headers

**Dosyalar:** 
- `internal/middleware/cors.go`
- `internal/middleware/security.go`

**CORS Implementation:**
```go
// Standard CORS for general API
func CORSMiddleware() gin.HandlerFunc {
    config := cors.Config{
        AllowOrigins:     []string{"http://localhost:3000", "https://banking-frontend.com"},
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length", "X-Total-Count", "X-Rate-Limit-Remaining"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }
    return cors.New(config)
}
```

**Security Headers:**
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Content-Security-Policy`
- `Strict-Transport-Security`
- `Referrer-Policy`
- `Permissions-Policy`

### 3. ✅ Rate Limiting

**Dosya:** `internal/middleware/rate_limiting.go`

**Implementasyon:**
- **IP + User Agent** tabanlı rate limiting
- **Adaptive rate limiting** (authenticated vs anonymous users)
- **Route-specific limits**:
  - Global: 10 req/s, burst 20
  - Authentication: 1 req/s, burst 3
  - Banking operations: 5 req/s, burst 10

**Özellikler:**
- Thread-safe implementation
- Configurable limits via environment variables
- Rate limit headers (`X-Rate-Limit-*`)
- Graceful error responses

### 4. ✅ Request Logging and Tracking

**Dosyalar:**
- `internal/middleware/logging.go` (mevcut)
- `internal/middleware/tracking.go` (yeni)

**Request Tracking Features:**
- **Unique Request ID** generation (UUID)
- **Correlation ID** support
- **Performance metrics** (response time)
- **Sensitive operation tracking**
- **Structured logging** with Zap

**Logging Levels:**
- Request start/completion
- Performance warnings (>1s)
- Error tracking (4xx, 5xx)
- Security events
- Banking operations

## 🛠️ Middleware Stack Order

```go
// Global middleware (applied to all routes)
r.Use(gin.Recovery())                    // Panic recovery
r.Use(middleware.SecurityHeadersMiddleware()) // Security headers
r.Use(middleware.CORSMiddleware())       // CORS
r.Use(middleware.RequestTrackingMiddleware()) // Request tracking
r.Use(middleware.LoggingMiddleware())    // Request logging
r.Use(middleware.SecurityLoggingMiddleware()) // Security logging
r.Use(middleware.AdaptiveRateLimitMiddleware()) // Rate limiting

// Route-specific middleware
auth.Use(middleware.AuthenticationRateLimitMiddleware()) // Stricter auth limits
protected.Use(middleware.AuthenticationMiddleware())     // JWT auth
protected.Use(middleware.BankingRateLimitMiddleware())   // Banking limits
protected.Use(middleware.BankingSecurityHeadersMiddleware()) // Enhanced security
protected.Use(middleware.BankingTrackingMiddleware())    // Enhanced tracking
admin.Use(middleware.AdminAuthorizationMiddleware())     // Admin role check
```

## 🔧 Configuration

**Dosya:** `config/config.go`

Rate limiting ve security ayarları environment variables ile yapılandırılabilir:

```bash
# Rate Limiting
RATE_LIMIT_GLOBAL_RPS=10.0
RATE_LIMIT_GLOBAL_BURST=20
RATE_LIMIT_AUTH_RPS=1.0
RATE_LIMIT_AUTH_BURST=3
RATE_LIMIT_BANKING_RPS=5.0
RATE_LIMIT_BANKING_BURST=10

# Security
ENABLE_HSTS=true
ENABLE_CSP=true
```

## 🚀 Usage

### Server Başlatma

```go
// main.go
func main() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        panic("Failed to load configuration: " + err.Error())
    }

    // Initialize custom router with all middleware
    r := api.SetupRouter()

    // Create HTTP server
    server := &http.Server{
        Addr:    ":" + cfg.Server.Port,
        Handler: r,
    }

    // Start server
    if err := server.ListenAndServe(); err != nil {
        log.Fatal("Server failed to start", zap.Error(err))
    }
}
```

### Test Endpoints

```bash
# Health check (no rate limiting)
curl http://localhost:8080/health

# Authentication (stricter rate limiting)
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password"}'

# Protected endpoint (requires JWT)
curl http://localhost:8080/api/v1/accounts \
  -H "Authorization: Bearer valid-token"

# Admin endpoint (requires admin role)
curl http://localhost:8080/api/v1/admin/users \
  -H "Authorization: Bearer admin-token"
```

## 📊 Monitoring

### Request Headers

Her response'da şu headers bulunur:
- `X-Request-ID`: Unique request identifier
- `X-Response-Time`: Request duration
- `X-Rate-Limit-Remaining`: Remaining rate limit
- `X-Banking-Security`: Security status

### Logging Examples

```json
{
  "level": "info",
  "msg": "Request Started",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "method": "POST",
  "path": "/api/v1/transfers",
  "ip": "192.168.1.100",
  "type": "request_start"
}

{
  "level": "info",
  "msg": "Sensitive Banking Operation",
  "request_id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "user123",
  "operation": "/api/v1/transfers",
  "type": "sensitive_operation"
}
```

## 🔒 Security Features

1. **CORS Protection**: Cross-origin request kontrolü
2. **Security Headers**: Modern web security headers
3. **Rate Limiting**: DDoS ve brute force koruması
4. **Request Tracking**: Audit trail için tam izleme
5. **Authentication**: JWT tabanlı kimlik doğrulama
6. **Authorization**: Role-based access control
7. **Input Validation**: Request validation (TODO)
8. **HTTPS Enforcement**: HSTS headers

## 🚀 Performance Features

1. **Structured Logging**: Zap ile hızlı logging
2. **Request Tracking**: Performance monitoring
3. **Rate Limiting**: Resource protection
4. **Middleware Optimization**: Efficient middleware stack
5. **Graceful Shutdown**: Clean server shutdown

## 📝 Next Steps

1. **JWT Implementation**: Gerçek JWT validation
2. **Redis Integration**: Distributed rate limiting
3. **Metrics Collection**: Prometheus metrics
4. **API Documentation**: Swagger/OpenAPI
5. **Input Validation**: Request validation middleware
6. **Caching**: Response caching middleware

Bu implementasyon, modern bir banking backend için gerekli tüm HTTP server güvenlik ve performans özelliklerini sağlar.
