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

### 🐳 **Containerization & DevOps**
- **Docker** - Containerization
- **Docker Compose** - Multi-service orchestration
- **Multi-stage builds** - Optimized images
- **Health checks** - Container monitoring

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
├── migrations/          # Database migration dosyaları
├── config/              # Konfigürasyon dosyaları
├── Dockerfile           # Multi-stage Docker build
├── docker-compose.yml   # Docker Compose konfigürasyonu
├── .dockerignore        # Docker build optimizasyonu
├── env.example          # Environment variables template
└── Makefile             # Build ve development komutları
```

## 🚀 Kurulum

### 🐳 **Docker ile Kurulum (Önerilen)**

1. **Repository'yi klonlayın:**
```bash
git clone https://github.com/barannkoca/banking-backend.git
cd banking-backend
```

2. **Environment dosyasını hazırlayın:**
```bash
cp env.example .env
# Gerekirse .env dosyasını düzenleyin
```

3. **Docker servislerini başlatın:**
```bash
# Tüm servisleri başlat (PostgreSQL + Redis + App)
./scripts/docker-run.sh

# Veya manuel olarak
docker-compose up -d
```

4. **Servisleri test edin:**
```bash
# API health check
curl http://localhost:8080/health

# Logları takip edin
docker-compose logs -f
```

### 🔧 **Manuel Kurulum**

1. **Repository'yi klonlayın:**
```bash
git clone https://github.com/barannkoca/banking-backend.git
cd banking-backend
```

2. **Bağımlılıkları yükleyin:**
```bash
go mod tidy
```

3. **Projeyi derleyin:**
```bash
go build -o bin/banking-backend cmd/server/main.go
```

4. **HTTP Server Setup'ı test edin:**
```bash
./scripts/test_http_setup.sh
```

5. **API Endpoints'leri test edin:**
```bash
./scripts/test_api_endpoints.sh
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

### 🚀 HTTP Server Setup

HTTP Server Setup implementasyonu tamamlanmıştır. Detaylar için `docs/HTTP_SERVER_SETUP.md` dosyasını inceleyin.

**Implement edilen özellikler:**
- ✅ Custom Router with Middleware Support
- ✅ CORS and Security Headers
- ✅ Rate Limiting
- ✅ Request Logging and Tracking
- ✅ Authentication and Authorization
- ✅ Role-based Access Control
- ✅ Performance Monitoring
- ✅ Security Logging

### 🐳 Docker Setup

Docker kurulumu tamamlanmıştır. Detaylar için `docs/DOCKER_SETUP.md` dosyasını inceleyin.

**Docker Özellikleri:**
- ✅ Multi-stage Dockerfile
- ✅ Docker Compose (App + PostgreSQL + Redis)
- ✅ Health Checks
- ✅ Database Migration
- ✅ Security (Non-root user)
- ✅ Monitoring
- ✅ Helper Scripts

**Hızlı Başlangıç:**
```bash
# Environment hazırla
cp env.example .env

# Servisleri başlat
./scripts/docker-run.sh

# Test et
curl http://localhost:8080/health
```

**Docker Komutları:**
```bash
# Build
./scripts/docker-build.sh

# Run
./scripts/docker-run.sh

# Cleanup
./scripts/docker-cleanup.sh

# Manuel kontrol
docker-compose up -d
docker-compose down
docker-compose logs -f
```

### 📡 API Endpoints

API endpoint'leri implement edilmiştir. Detaylar için `docs/API_ENDPOINTS.md` dosyasını inceleyin.

**Implement edilen endpoint'ler:**

#### 🔐 Authentication Endpoints
- `POST /api/v1/auth/register` - Kullanıcı kaydı
- `POST /api/v1/auth/login` - Kullanıcı girişi
- `POST /api/v1/auth/refresh` - Token yenileme

#### 👥 User Management Endpoints
- `GET /api/v1/users` - Tüm kullanıcıları listele
- `GET /api/v1/users/{id}` - Kullanıcı bilgilerini getir
- `PUT /api/v1/users/{id}` - Kullanıcı bilgilerini güncelle
- `DELETE /api/v1/users/{id}` - Kullanıcı hesabını sil

#### 💰 Transaction Endpoints
- `POST /api/v1/transactions/credit` - Kredi işlemi
- `POST /api/v1/transactions/debit` - Borç işlemi
- `POST /api/v1/transactions/transfer` - Transfer işlemi
- `GET /api/v1/transactions/history` - İşlem geçmişi
- `GET /api/v1/transactions/{id}` - İşlem detayları

#### 💳 Balance Endpoints
- `GET /api/v1/balances/current` - Mevcut bakiye
- `GET /api/v1/balances/historical` - Geçmiş bakiye
- `GET /api/v1/balances/at-time` - Belirli zamandaki bakiye

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
