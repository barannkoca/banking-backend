# 🐳 Docker Setup Guide

Bu dokümantasyon Banking Backend projesinin Docker ile nasıl çalıştırılacağını açıklar.

## 📋 Gereksinimler

- Docker 20.10+
- Docker Compose 2.0+
- En az 4GB RAM
- En az 2GB disk alanı

## 🚀 Hızlı Başlangıç

### 1. Environment Dosyasını Hazırlayın

```bash
# Environment dosyasını kopyalayın
cp env.example .env

# Gerekirse .env dosyasını düzenleyin
nano .env
```

### 2. Docker Compose ile Çalıştırın

```bash
# Tüm servisleri başlatın
docker-compose up -d

# Logları takip edin
docker-compose logs -f
```

### 3. Servisleri Test Edin

```bash
# API health check
curl http://localhost:8080/health

# PostgreSQL bağlantısı
docker-compose exec postgres psql -U barankoca -d banking_db -c "SELECT version();"

# Redis bağlantısı
docker-compose exec redis redis-cli ping
```

## 🛠️ Detaylı Kurulum

### Multi-Stage Dockerfile

Proje multi-stage Dockerfile kullanır:

- **Stage 1 (Builder)**: Go uygulamasını derler
- **Stage 2 (Runtime)**: Minimal Alpine Linux ile çalıştırır

### Docker Compose Servisleri

#### 1. Banking Backend Application
- **Port**: 8080
- **Image**: Custom build
- **Health Check**: `/health` endpoint
- **Dependencies**: PostgreSQL, Redis

#### 2. PostgreSQL Database
- **Port**: 5432
- **Image**: postgres:15-alpine
- **Database**: banking_db
- **User**: barankoca
- **Password**: secure_password_123
- **Health Check**: pg_isready

#### 3. Redis Cache
- **Port**: 6379
- **Image**: redis:7-alpine
- **Memory Limit**: 256MB
- **Policy**: allkeys-lru
- **Health Check**: redis-cli ping

#### 4. Database Migration
- **Image**: migrate/migrate:v4.18.3
- **Purpose**: Database schema migration
- **Runs**: Once on startup

## 📜 Yardımcı Scriptler

### Docker Build Script
```bash
./scripts/docker-build.sh
```
- Docker image'ını build eder
- Image'ı test eder
- Boyut bilgilerini gösterir

### Docker Run Script
```bash
./scripts/docker-run.sh
```
- Tüm servisleri başlatır
- Health check yapar
- Kullanışlı komutları gösterir

### Docker Cleanup Script
```bash
./scripts/docker-cleanup.sh
```
- Container'ları temizler
- Kullanılmayan image'ları siler
- Volume'ları temizler

## 🔧 Konfigürasyon

### Environment Variables

| Variable | Default | Açıklama |
|----------|---------|----------|
| `ENVIRONMENT` | development | Çalışma ortamı |
| `SERVER_PORT` | 8080 | API portu |
| `DB_HOST` | postgres | Database host |
| `DB_PORT` | 5432 | Database port |
| `DB_USER` | barankoca | Database kullanıcısı |
| `DB_PASSWORD` | secure_password_123 | Database şifresi |
| `DB_NAME` | banking_db | Database adı |
| `REDIS_HOST` | redis | Redis host |
| `REDIS_PORT` | 6379 | Redis port |
| `JWT_SECRET` | your-super-secure... | JWT secret key |

### Volume'lar

- `postgres_data`: PostgreSQL verileri
- `redis_data`: Redis verileri

### Network

- `banking-network`: Tüm servisler arası iletişim

## 🐛 Troubleshooting

### Port Çakışması
```bash
# Port kullanımını kontrol edin
lsof -i :8080
lsof -i :5432
lsof -i :6379

# Farklı port kullanın
docker-compose up -d --scale banking-backend=0
# docker-compose.yml'de port değiştirin
```

### Database Bağlantı Sorunu
```bash
# PostgreSQL loglarını kontrol edin
docker-compose logs postgres

# Database'e manuel bağlanın
docker-compose exec postgres psql -U barankoca -d banking_db
```

### Redis Bağlantı Sorunu
```bash
# Redis loglarını kontrol edin
docker-compose logs redis

# Redis'e manuel bağlanın
docker-compose exec redis redis-cli
```

### Memory Sorunu
```bash
# Container resource kullanımını kontrol edin
docker stats

# Memory limit artırın (docker-compose.yml)
services:
  redis:
    deploy:
      resources:
        limits:
          memory: 512M
```

## 📊 Monitoring

### Container Durumu
```bash
# Tüm container'ları listele
docker-compose ps

# Resource kullanımı
docker stats

# Logları takip et
docker-compose logs -f banking-backend
```

### Health Check
```bash
# API health
curl http://localhost:8080/health

# Database health
docker-compose exec postgres pg_isready -U barankoca -d banking_db

# Redis health
docker-compose exec redis redis-cli ping
```

## 🔄 Güncelleme

### Code Değişikliği
```bash
# Yeni image build et
docker-compose build banking-backend

# Servisleri yeniden başlat
docker-compose up -d banking-backend
```

### Database Migration
```bash
# Migration çalıştır
docker-compose run --rm migrate

# Yeni migration ekle
# migrations/ klasörüne yeni dosya ekleyin
```

## 🧹 Temizlik

### Geçici Temizlik
```bash
# Sadece container'ları durdur
docker-compose down

# Volume'ları da sil
docker-compose down -v
```

### Tam Temizlik
```bash
# Cleanup script kullanın
./scripts/docker-cleanup.sh
```

## 📚 Ek Kaynaklar

- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [PostgreSQL Docker Image](https://hub.docker.com/_/postgres)
- [Redis Docker Image](https://hub.docker.com/_/redis)
- [Go Docker Best Practices](https://docs.docker.com/language/golang/)

## 🤝 Katkıda Bulunma

Docker setup'ında iyileştirme önerileriniz varsa:

1. Issue açın
2. Pull request gönderin
3. Dokümantasyonu güncelleyin

---

**Not**: Production ortamında mutlaka güvenlik ayarlarını gözden geçirin ve güçlü şifreler kullanın.
