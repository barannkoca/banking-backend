# Redis Cache Implementation

Bu dokümantasyon, Banking Backend projesinde Redis cache implementasyonunu açıklar.

## 📋 İçindekiler

- [Genel Bakış](#genel-bakış)
- [Kurulum](#kurulum)
- [Kullanım](#kullanım)
- [API Referansı](#api-referansı)
- [Performance Monitoring](#performance-monitoring)
- [Test](#test)
- [Troubleshooting](#troubleshooting)

## 🎯 Genel Bakış

Redis cache sistemi, banking uygulamasının performansını artırmak için kullanılır. Aşağıdaki veriler cache'lenir:

- **Balance verileri**: Kullanıcı bakiyeleri
- **Transaction verileri**: İşlem geçmişi
- **User verileri**: Kullanıcı bilgileri
- **Cache istatistikleri**: Hit/miss oranları

### Özellikler

- ✅ Thread-safe operations
- ✅ Automatic serialization/deserialization
- ✅ Cache statistics tracking
- ✅ Performance monitoring integration
- ✅ Banking-specific operations
- ✅ Graceful error handling

## 🚀 Kurulum

### 1. Redis Server Kurulumu

```bash
# macOS (Homebrew)
brew install redis
brew services start redis

# Ubuntu/Debian
sudo apt-get install redis-server
sudo systemctl start redis-server

# CentOS/RHEL
sudo yum install redis
sudo systemctl start redis
```

### 2. Go Dependencies

```bash
go get github.com/redis/go-redis/v9
```

### 3. Konfigürasyon

Redis bağlantı ayarları `main.go` dosyasında yapılandırılır:

```go
cacheService, err := services.NewRedisCacheService("localhost:6379", "", 0)
```

**Parametreler:**
- `addr`: Redis server adresi (varsayılan: localhost:6379)
- `password`: Redis şifresi (varsayılan: boş)
- `db`: Redis veritabanı numarası (varsayılan: 0)

## 🔧 Kullanım

### Basic Operations

```go
// Cache service'i al
cacheService := services.NewRedisCacheService("localhost:6379", "", 0)

// Veri kaydet
err := cacheService.Set(ctx, "key", "value", 5*time.Minute)

// Veri al
value, err := cacheService.Get(ctx, "key")

// Veri sil
err := cacheService.Delete(ctx, "key")
```

### Banking-Specific Operations

```go
// Balance cache
balance, err := cacheService.GetBalance(ctx, userID)
err = cacheService.SetBalance(ctx, userID, 1000.50, 5*time.Minute)
err = cacheService.InvalidateBalance(ctx, userID)

// Transaction cache
transaction, err := cacheService.GetTransaction(ctx, transactionID)
err = cacheService.SetTransaction(ctx, transactionID, transaction, 10*time.Minute)
err = cacheService.InvalidateTransaction(ctx, transactionID)

// User transactions cache
transactions, err := cacheService.GetUserTransactions(ctx, userID, 10, 0)
err = cacheService.SetUserTransactions(ctx, userID, transactions, 5*time.Minute)
err = cacheService.InvalidateUserTransactions(ctx, userID)
```

## 📚 API Referansı

### CacheService Interface

```go
type CacheService interface {
    // Basic operations
    Get(ctx context.Context, key string) (interface{}, error)
    Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
    Delete(ctx context.Context, key string) error
    Exists(ctx context.Context, key string) (bool, error)

    // Hash operations
    HGet(ctx context.Context, key, field string) (string, error)
    HSet(ctx context.Context, key string, values map[string]interface{}) error
    HGetAll(ctx context.Context, key string) (map[string]string, error)
    HDel(ctx context.Context, key string, fields ...string) error

    // List operations
    LPush(ctx context.Context, key string, values ...interface{}) error
    RPush(ctx context.Context, key string, values ...interface{}) error
    LPop(ctx context.Context, key string) (string, error)
    RPop(ctx context.Context, key string) (string, error)
    LRange(ctx context.Context, key string, start, stop int64) ([]string, error)

    // Set operations
    SAdd(ctx context.Context, key string, members ...interface{}) error
    SRem(ctx context.Context, key string, members ...interface{}) error
    SMembers(ctx context.Context, key string) ([]string, error)
    SIsMember(ctx context.Context, key string, member interface{}) (bool, error)

    // Utility operations
    FlushDB(ctx context.Context) error
    Ping(ctx context.Context) error
    Close() error

    // Banking-specific operations
    GetBalance(ctx context.Context, userID string) (float64, error)
    SetBalance(ctx context.Context, userID string, balance float64, expiration time.Duration) error
    InvalidateBalance(ctx context.Context, userID string) error

    GetTransaction(ctx context.Context, transactionID string) (interface{}, error)
    SetTransaction(ctx context.Context, transactionID string, transaction interface{}, expiration time.Duration) error
    InvalidateTransaction(ctx context.Context, transactionID string) error

    GetUserTransactions(ctx context.Context, userID string, limit, offset int) ([]interface{}, error)
    SetUserTransactions(ctx context.Context, userID string, transactions []interface{}, expiration time.Duration) error
    InvalidateUserTransactions(ctx context.Context, userID string) error

    // Cache statistics
    GetStats() map[string]interface{}
}
```

### Cache Key Patterns

```go
// Balance keys
"balance:{user_id}"

// Transaction keys
"transaction:{transaction_id}"

// User transaction keys
"user_transactions:{user_id}:{limit}:{offset}"

// User keys
"user:{user_id}"
```

## 📊 Performance Monitoring

Cache sistemi, performance monitoring middleware ile entegre edilmiştir:

### Cache Metrics

- **Hits**: Başarılı cache istekleri
- **Misses**: Cache'de bulunamayan istekler
- **Sets**: Cache'e yazma işlemleri
- **Deletes**: Cache'den silme işlemleri
- **Errors**: Cache hataları
- **Hit Rate**: Hit oranı (%)

### Performance Headers

Cache işlemleri sırasında aşağıdaki HTTP header'ları eklenir:

```
X-Cache-Hit: true/false
X-Cache-Response-Time: 1.23ms
X-Cache-Keys: balance:user123
```

### Logging

Cache işlemleri Zap logger ile loglanır:

```json
{
  "level": "info",
  "msg": "Cache operation completed",
  "operation": "get",
  "key": "balance:user123",
  "hit": true,
  "duration": "1.23ms"
}
```

## 🧪 Test

### Test Script'i Çalıştırma

```bash
# Redis cache test'ini çalıştır
./scripts/test_redis_cache.sh
```

### Manuel Test

```bash
# Redis bağlantısını test et
redis-cli ping

# Cache health endpoint'ini test et
curl http://localhost:8080/health/cache

# Cache istatistiklerini görüntüle
redis-cli info memory
redis-cli dbsize
```

### Test Coverage

Test script'i aşağıdaki alanları test eder:

- ✅ Redis bağlantısı
- ✅ Basic cache operations (GET, SET, DELETE)
- ✅ Banking-specific operations
- ✅ Cache health endpoint
- ✅ Cache statistics
- ✅ Performance testing

## 🔍 Troubleshooting

### Yaygın Sorunlar

#### 1. Redis Bağlantı Hatası

```
Error: failed to connect to Redis: dial tcp localhost:6379: connect: connection refused
```

**Çözüm:**
```bash
# Redis'i başlat
brew services start redis

# Bağlantıyı test et
redis-cli ping
```

#### 2. Cache Service Başlatılamıyor

```
Error: Redis cache service initialization failed
```

**Çözüm:**
- Redis server'ın çalıştığından emin ol
- Bağlantı parametrelerini kontrol et
- Firewall ayarlarını kontrol et

#### 3. Cache Hit Rate Düşük

**Çözüm:**
- Cache TTL sürelerini artır
- Cache key pattern'lerini optimize et
- Cache invalidation stratejisini gözden geçir

### Debug Komutları

```bash
# Redis log'larını görüntüle
tail -f /var/log/redis/redis-server.log

# Redis memory kullanımını kontrol et
redis-cli info memory

# Cache key'lerini listele
redis-cli keys "*"

# Belirli pattern'deki key'leri listele
redis-cli keys "balance:*"
```

### Performance Optimizasyonu

#### 1. Memory Kullanımı

```bash
# Memory kullanımını optimize et
redis-cli config set maxmemory-policy allkeys-lru
redis-cli config set maxmemory 100mb
```

#### 2. Connection Pool

```go
// Connection pool boyutunu artır
client := redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    PoolSize: 20, // Varsayılan: 10
})
```

#### 3. Cache TTL Stratejisi

```go
// Balance cache: 5 dakika
cacheService.SetBalance(ctx, userID, balance, 5*time.Minute)

// Transaction cache: 10 dakika
cacheService.SetTransaction(ctx, transactionID, transaction, 10*time.Minute)

// User cache: 30 dakika
cacheService.CacheUser(ctx, user, 1800)
```

## 📈 Monitoring ve Alerting

### Cache Metrics Dashboard

Cache performansını izlemek için aşağıdaki metrikleri takip edin:

- **Hit Rate**: %80+ hedef
- **Response Time**: <5ms hedef
- **Memory Usage**: <80% hedef
- **Error Rate**: <1% hedef

### Alerting Rules

```yaml
# Cache hit rate düşük
- alert: CacheHitRateLow
  expr: cache_hit_rate < 0.8
  for: 5m

# Cache response time yüksek
- alert: CacheResponseTimeHigh
  expr: cache_response_time > 0.005
  for: 2m

# Cache memory kullanımı yüksek
- alert: CacheMemoryHigh
  expr: cache_memory_usage > 0.8
  for: 5m
```

## 🔐 Güvenlik

### Redis Güvenlik Önerileri

1. **Authentication**: Redis şifresi kullan
2. **Network Security**: Redis'i sadece localhost'ta çalıştır
3. **SSL/TLS**: Production'da SSL kullan
4. **Key Patterns**: Güvenli key pattern'leri kullan

### Güvenli Konfigürasyon

```go
// Production Redis konfigürasyonu
cacheService, err := services.NewRedisCacheService(
    "redis.example.com:6379",
    "strong-password",
    0,
)
```

## 📝 Changelog

### v1.0.0 (2025-08-28)
- ✅ Initial Redis cache implementation
- ✅ Basic cache operations
- ✅ Banking-specific cache methods
- ✅ Performance monitoring integration
- ✅ Cache statistics tracking
- ✅ Health check endpoint
- ✅ Comprehensive test suite

## 🤝 Katkıda Bulunma

1. Fork yap
2. Feature branch oluştur (`git checkout -b feature/redis-cache-improvement`)
3. Commit yap (`git commit -am 'Add Redis cache improvement'`)
4. Push yap (`git push origin feature/redis-cache-improvement`)
5. Pull Request oluştur

## 📞 Destek

Redis cache ile ilgili sorunlar için:

- 📧 Email: support@banking-backend.com
- 🐛 Issues: GitHub Issues
- 📖 Docs: Bu dokümantasyon
- 💬 Chat: Slack #redis-cache
