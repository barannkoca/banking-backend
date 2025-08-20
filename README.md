# 🏦 Banking Backend

Go ile geliştirilmiş modern bankacılık backend API'si.

## 🚀 Özellikler

### 🔐 **Authentication & Authorization**
- JWT tabanlı kimlik doğrulama
- Role-based authorization (Admin, Manager, Customer)
- Password hashing (bcrypt)
- Session management (Redis)

### 💳 **Banking Operations**
- Hesap yönetimi ve oluşturma
- Para transferi işlemleri
- Credit/Debit operations
- Transaction rollback mechanism
- Multi-currency support

### 📊 **Monitoring & Security**
- İşlem geçmişi ve audit trails
- Rate limiting (API koruması)
- CORS security
- Structured logging (Zap)
- Performance monitoring

### 🔄 **Advanced Features**
- Database migrations
- Redis caching
- Concurrent transaction processing
- Worker pools
- Graceful shutdown

## 🛠️ Teknolojiler

### 🚀 **Core Framework**
- **Go 1.25+** - Modern Go version
- **Gin Web Framework v1.10.1** - HTTP router ve middleware
- **Zap v1.27.0** - Ultra-fast structured logging

### 🔐 **Authentication & Security**
- **JWT v5.3.0** - JSON Web Tokens için authentication
- **bcrypt** - Güvenli password hashing
- **UUID v1.6.0** - Unique identifier generation
- **CORS v1.7.6** - Cross-Origin Resource Sharing
- **Secure Headers v1.1.2** - Security headers middleware

### 🗄️ **Database & Storage**
- **PostgreSQL** - Primary database
- **lib/pq v1.10.9** - PostgreSQL driver
- **SQLx v1.4.0** - SQL extensions ve struct mapping
- **Migrate v4.18.3** - Database migrations
- **Redis v8.11.5** - Caching ve session storage

### ⚙️ **Configuration & Utils**
- **godotenv v1.5.1** - Environment variables management
- **Validator v10.26.0** - Request validation
- **Rate Limiter v3.11.2** - API rate limiting

### 🛡️ **Performance & Monitoring**
- **Prometheus metrics** - System monitoring
- **Structured logging** - Audit trails
- **Worker pools** - Concurrent processing

## 📁 Proje Yapısı

```
├── cmd/server/          # Ana uygulama giriş noktası
├── internal/            # Proje içi kodlar
│   ├── api/v1/          # HTTP handler'ları ve route'lar
│   ├── models/          # Veritabanı modelleri
│   ├── services/        # İş mantığı katmanı
│   ├── database/        # Veritabanı işlemleri
│   └── middleware/      # Ara katmanlar
├── pkg/                 # Dışa açık paketler
│   ├── utils/           # Yardımcı fonksiyonlar
│   └── logger/          # Loglama sistemi
├── docs/                # API dokümantasyonu
├── scripts/             # Build ve deploy scriptleri
└── config/              # Konfigürasyon dosyaları
```

## 🚀 Kurulum

1. **Repository'yi klonlayın:**
```bash
git clone https://github.com/barannkoca/banking-backend.git
cd banking-backend
```

2. **Bağımlılıkları yükleyin:**
```bash
go mod tidy
```

**Ana Dependencies:**
```bash
# Core Framework
go get github.com/gin-gonic/gin@v1.10.1
go get go.uber.org/zap@v1.27.0

# Authentication & Security  
go get github.com/golang-jwt/jwt/v5@v5.3.0
go get golang.org/x/crypto/bcrypt
go get github.com/google/uuid@v1.6.0
go get github.com/gin-contrib/cors@v1.7.6

# Database
go get github.com/lib/pq@v1.10.9
go get github.com/jmoiron/sqlx@v1.4.0
go get github.com/golang-migrate/migrate/v4@v4.18.3

# Configuration & Performance
go get github.com/joho/godotenv@v1.5.1
go get github.com/go-redis/redis/v8@v8.11.5
go get github.com/ulule/limiter/v3@v3.11.2
```

3. **Environment dosyasını oluşturun:**
```bash
cp .env.example .env
```

4. **Uygulamayı çalıştırın:**
```bash
go run cmd/server/main.go
```

## 📝 API Dokümantasyonu

API dokümantasyonu `docs/` klasöründe bulunur. Swagger UI için `/docs` endpoint'ini ziyaret edin.

## 🤝 Katkıda Bulunma

1. Fork yapın
2. Feature branch oluşturun (`git checkout -b feature/amazing-feature`)
3. Commit yapın (`git commit -m 'Add amazing feature'`)
4. Push yapın (`git push origin feature/amazing-feature`)
5. Pull Request oluşturun

## 📄 Lisans

Bu proje MIT lisansı altında lisanslanmıştır.

## 👨‍💻 Geliştirici

**Baran Koca** - [GitHub](https://github.com/barannkoca)
