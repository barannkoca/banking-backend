#!/bin/bash

# Docker Build Script for Banking Backend
# Bu script Docker image'ını build eder ve test eder

set -e

echo "🐳 Banking Backend Docker Build Script"
echo "======================================"

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

log_info "Docker build işlemi başlatılıyor..."

# Docker image'ını build et
log_info "Docker image build ediliyor..."
docker build -t banking-backend:latest .

if [ $? -eq 0 ]; then
    log_success "Docker image başarıyla build edildi!"
else
    log_error "Docker image build edilemedi!"
    exit 1
fi

# Image boyutunu göster
log_info "Image boyutu:"
docker images banking-backend:latest --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}"

# Image'ı test et
log_info "Image test ediliyor..."
docker run --rm banking-backend:latest --help > /dev/null 2>&1

if [ $? -eq 0 ]; then
    log_success "Image test edildi ve çalışıyor!"
else
    log_warning "Image test edilemedi, ancak build başarılı."
fi

log_success "Docker build işlemi tamamlandı!"
echo ""
echo "Kullanım:"
echo "  docker run -p 8080:8080 banking-backend:latest"
echo "  docker-compose up"
