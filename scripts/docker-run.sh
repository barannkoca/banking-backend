#!/bin/bash

# Docker Run Script for Banking Backend
# Bu script Docker Compose ile tüm servisleri başlatır

set -e

echo "🚀 Banking Backend Docker Run Script"
echo "===================================="

# Renk kodları
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Fonksiyonlar
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Docker'ın çalışıp çalışmadığını kontrol et
if ! docker info > /dev/null 2>&1; then
    log_error "Docker çalışmıyor. Lütfen Docker'ı başlatın."
    exit 1
fi

# Docker Compose'un varlığını kontrol et
if ! command -v docker-compose &> /dev/null; then
    log_error "docker-compose bulunamadı. Lütfen Docker Compose'u yükleyin."
    exit 1
fi

# Environment dosyasını kontrol et
if [ ! -f ".env" ]; then
    log_warning ".env dosyası bulunamadı. env.example'dan kopyalanıyor..."
    if [ -f "env.example" ]; then
        cp env.example .env
        log_success ".env dosyası oluşturuldu."
    else
        log_error "env.example dosyası da bulunamadı!"
        exit 1
    fi
fi

# Önceki container'ları temizle
log_info "Önceki container'lar temizleniyor..."
docker-compose down --remove-orphans

# Servisleri başlat
log_info "Servisler başlatılıyor..."
docker-compose up --build -d

# Servislerin hazır olmasını bekle
log_info "Servislerin hazır olması bekleniyor..."
sleep 10

# Health check
log_info "Health check yapılıyor..."

# PostgreSQL health check
if docker-compose exec -T postgres pg_isready -U barankoca -d banking_db > /dev/null 2>&1; then
    log_success "PostgreSQL hazır!"
else
    log_warning "PostgreSQL henüz hazır değil."
fi

# Redis health check
if docker-compose exec -T redis redis-cli ping > /dev/null 2>&1; then
    log_success "Redis hazır!"
else
    log_warning "Redis henüz hazır değil."
fi

# Application health check
sleep 5
if curl -f http://localhost:8080/health > /dev/null 2>&1; then
    log_success "Banking Backend API hazır!"
else
    log_warning "Banking Backend API henüz hazır değil."
fi

echo ""
log_success "Tüm servisler başlatıldı!"
echo ""
echo "🌐 Servisler:"
echo "  - Banking Backend API: http://localhost:8080"
echo "  - PostgreSQL: localhost:5432"
echo "  - Redis: localhost:6379"
echo ""
echo "📋 Kullanışlı komutlar:"
echo "  - Logları görüntüle: docker-compose logs -f"
echo "  - Servisleri durdur: docker-compose down"
echo "  - Container'lara bağlan: docker-compose exec banking-backend sh"
echo "  - Database'e bağlan: docker-compose exec postgres psql -U barankoca -d banking_db"
echo "  - Redis'e bağlan: docker-compose exec redis redis-cli"
echo ""
echo "🔍 API Test:"
echo "  curl http://localhost:8080/health"
