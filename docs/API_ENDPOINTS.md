# 🚀 API Endpoints Documentation

Bu dokümantasyon, Banking Backend API'sinin tüm endpoint'lerini detaylandırır.

## 📋 API Overview

**Base URL:** `http://localhost:8080/api/v1`  
**Authentication:** JWT Bearer Token  
**Content-Type:** `application/json`

## 🔐 Authentication Endpoints

### POST /api/v1/auth/register
Kullanıcı kaydı oluşturur.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securepassword123",
  "first_name": "John",
  "last_name": "Doe",
  "phone": "+905551234567"
}
```

**Response:**
```json
{
  "message": "User registration endpoint",
  "endpoint": "POST /api/v1/auth/register",
  "description": "Register a new user account"
}
```

### POST /api/v1/auth/login
Kullanıcı girişi yapar ve JWT token döner.

**Request Body:**
```json
{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Response:**
```json
{
  "message": "User login endpoint",
  "endpoint": "POST /api/v1/auth/login",
  "description": "Authenticate user and return JWT token"
}
```

### POST /api/v1/auth/refresh
JWT token'ı yeniler.

**Request Body:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Response:**
```json
{
  "message": "Token refresh endpoint",
  "endpoint": "POST /api/v1/auth/refresh",
  "description": "Refresh JWT token"
}
```

## 👥 User Management Endpoints

*Bu endpoint'ler authentication gerektirir.*

### GET /api/v1/users
Tüm kullanıcıları listeler (Admin only).

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "message": "Get all users endpoint",
  "endpoint": "GET /api/v1/users",
  "description": "Retrieve list of all users (admin only)"
}
```

### GET /api/v1/users/{id}
Belirli bir kullanıcının bilgilerini getirir.

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "message": "Get user by ID endpoint",
  "endpoint": "GET /api/v1/users/123",
  "description": "Retrieve user information by ID",
  "user_id": "123"
}
```

### PUT /api/v1/users/{id}
Kullanıcı bilgilerini günceller.

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Request Body:**
```json
{
  "first_name": "John",
  "last_name": "Smith",
  "phone": "+905551234567"
}
```

**Response:**
```json
{
  "message": "Update user endpoint",
  "endpoint": "PUT /api/v1/users/123",
  "description": "Update user information",
  "user_id": "123"
}
```

### DELETE /api/v1/users/{id}
Kullanıcı hesabını siler.

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "message": "Delete user endpoint",
  "endpoint": "DELETE /api/v1/users/123",
  "description": "Delete user account",
  "user_id": "123"
}
```

## 💰 Transaction Endpoints

*Bu endpoint'ler authentication gerektirir.*

### POST /api/v1/transactions/credit
Hesaba para ekler (kredi işlemi). Worker pool ile asenkron olarak işlenir.

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Request Body:**
```json
{
  "amount": 1000.50,
  "reference": "Salary deposit"
}
```

**Response:**
```json
{
  "message": "Para yatırma işlemi başlatıldı",
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "amount": 1000.50,
  "status": "processing",
  "created_at": "2024-01-15T10:30:00Z"
}
```

### POST /api/v1/transactions/debit
Hesaptan para çıkarır (borç işlemi). Worker pool ile asenkron olarak işlenir.

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Request Body:**
```json
{
  "amount": 250.75,
  "reference": "ATM withdrawal"
}
```

**Response:**
```json
{
  "message": "Para çekme işlemi başlatıldı",
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "amount": 250.75,
  "status": "processing",
  "created_at": "2024-01-15T10:30:00Z"
}
```

### POST /api/v1/transactions/transfer
Hesaplar arası para transferi yapar. Worker pool ile asenkron olarak işlenir.

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Request Body:**
```json
{
  "to_user_id": "550e8400-e29b-41d4-a716-446655440000",
  "amount": 500.00,
  "reference": "Transfer to savings account"
}
```

**Response:**
```json
{
  "message": "Transfer işlemi başlatıldı",
  "job_id": "550e8400-e29b-41d4-a716-446655440000",
  "from_user_id": "current-user-id",
  "to_user_id": "550e8400-e29b-41d4-a716-446655440000",
  "amount": 500.00,
  "status": "processing",
  "created_at": "2024-01-15T10:30:00Z"
}
```

### GET /api/v1/transactions/history
Kullanıcının işlem geçmişini getirir.

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Query Parameters:**
- `limit`: Sayfa başına kayıt sayısı (default: 50, max: 100)
- `offset`: Başlangıç pozisyonu (default: 0)
- `type`: İşlem tipi (deposit, withdraw, transfer)
- `status`: İşlem durumu (pending, completed, failed)

**Response:**
```json
{
  "message": "İşlem geçmişi başarıyla getirildi",
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "from_user_id": null,
      "to_user_id": "user-id",
      "amount": 1000.50,
      "type": "deposit",
      "status": "completed",
      "reference": "Salary deposit",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ],
  "pagination": {
    "limit": 50,
    "offset": 0,
    "count": 1
  }
}
```

### GET /api/v1/transactions/{id}
Belirli bir işlemin detaylarını getirir.

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "message": "İşlem başarıyla getirildi",
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "from_user_id": "user-id-1",
    "to_user_id": "user-id-2",
    "amount": 500.00,
    "type": "transfer",
    "status": "completed",
    "reference": "Transfer to savings account",
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

## 💳 Balance Endpoints

*Bu endpoint'ler authentication gerektirir.*

### GET /api/v1/balances/current
Mevcut hesap bakiyesini getirir.

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response:**
```json
{
  "message": "Mevcut bakiye başarıyla getirildi",
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "current_balance": 1500.75,
    "available_balance": 1500.75,
    "currency": "TRY",
    "last_updated": "2024-01-15T10:30:00Z"
  }
}
```

### GET /api/v1/balances/historical
Geçmiş bakiye verilerini getirir.

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Query Parameters:**
- `limit`: Sayfa başına kayıt sayısı (default: 50, max: 100)
- `offset`: Başlangıç pozisyonu (default: 0)
- `start_date`: Başlangıç tarihi (YYYY-MM-DD format)
- `end_date`: Bitiş tarihi (YYYY-MM-DD format)

**Response:**
```json
{
  "message": "Bakiye geçmişi başarıyla getirildi",
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "history": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "user_id": "550e8400-e29b-41d4-a716-446655440000",
        "previous_amount": 1000.00,
        "new_amount": 1500.75,
        "change_amount": 500.75,
        "change_type": "BALANCE_CREDIT",
        "transaction_id": "550e8400-e29b-41d4-a716-446655440000",
        "created_at": "2024-01-15T10:30:00Z"
      }
    ],
    "pagination": {
      "limit": 50,
      "offset": 0,
      "count": 1,
      "total": 1
    }
  }
}
```

### GET /api/v1/balances/at-time
Belirli bir zamandaki bakiyeyi getirir.

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Query Parameters:**
- `timestamp`: Zaman damgası (RFC3339, ISO 8601, YYYY-MM-DD formatları desteklenir)

**Response:**
```json
{
  "message": "Belirtilen zamandaki bakiye başarıyla getirildi",
  "data": {
    "user_id": "550e8400-e29b-41d4-a716-446655440000",
    "timestamp": "2024-01-15T10:30:00Z",
    "balance": 1500.75,
    "currency": "TRY",
    "calculated": true
  }
}
```

## 🔧 Health Check Endpoints

### GET /health
Servisin sağlık durumunu kontrol eder.

**Response:**
```json
{
  "status": "healthy",
  "service": "banking-backend"
}
```

### GET /health/ready
Servisin hazır olup olmadığını kontrol eder.

**Response:**
```json
{
  "status": "ready",
  "service": "banking-backend"
}
```

### GET /health/live
Servisin canlı olup olmadığını kontrol eder.

**Response:**
```json
{
  "status": "alive",
  "service": "banking-backend"
}
```

## 🛡️ Security Features

### Authentication
- JWT Bearer Token authentication
- Token expiration handling
- Refresh token mechanism

### Authorization
- Role-based access control (Admin, Manager, Customer)
- Endpoint-level permissions
- Resource ownership validation

### Rate Limiting
- Global rate limiting: 10 req/s, burst 20
- Authentication endpoints: 1 req/s, burst 3
- Banking operations: 5 req/s, burst 10

### Security Headers
- `X-Content-Type-Options: nosniff`
- `X-Frame-Options: DENY`
- `X-XSS-Protection: 1; mode=block`
- `Content-Security-Policy`
- `Strict-Transport-Security`
- `X-Request-ID`: Unique request identifier

## 📊 Response Headers

Her response'da şu headers bulunur:
- `X-Request-ID`: Unique request identifier
- `X-Response-Time`: Request duration
- `X-Rate-Limit-Remaining`: Remaining rate limit
- `X-Banking-Security`: Security status

## 🔍 Error Responses

### 400 Bad Request
```json
{
  "error": "Validation failed",
  "message": "Invalid request data",
  "details": ["Field 'email' is required"]
}
```

### 401 Unauthorized
```json
{
  "error": "Authentication required",
  "message": "Please provide valid JWT token"
}
```

### 403 Forbidden
```json
{
  "error": "Insufficient permissions",
  "message": "Admin role required for this operation"
}
```

### 404 Not Found
```json
{
  "error": "Resource not found",
  "message": "User with ID '123' not found"
}
```

### 429 Too Many Requests
```json
{
  "error": "Rate limit exceeded",
  "message": "Too many requests, please try again later",
  "retry_after": 1640995200
}
```

### 500 Internal Server Error
```json
{
  "error": "Internal server error",
  "message": "An unexpected error occurred"
}
```

## 🚀 Testing

API endpoint'lerini test etmek için:

```bash
# Test script'ini çalıştır
./scripts/test_api_endpoints.sh

# Manuel test
curl -X POST http://localhost:8080/api/v1/auth/register
curl -H "Authorization: Bearer valid-token" http://localhost:8080/api/v1/users
```

## 📝 Next Steps

1. **Business Logic Implementation**: Gerçek iş mantığı implementasyonu
2. **Database Integration**: Veritabanı işlemleri
3. **Request Validation**: Input validation middleware
4. **Error Handling**: Comprehensive error handling
5. **API Documentation**: Swagger/OpenAPI documentation
6. **Testing**: Unit ve integration testleri
