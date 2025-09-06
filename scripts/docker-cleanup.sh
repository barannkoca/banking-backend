#!/bin/bash

# Docker Cleanup Script for Banking Backend
# Bu script Docker container'larını, image'larını ve volume'larını temizler

set -e

echo "🧹 Banking Backend Docker Cleanup Script"
echo "======================================="

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

# Kullanıcıdan onay al
echo -e "${YELLOW}Bu script aşağıdaki işlemleri yapacak:${NC}"
echo "  - Tüm container'ları durduracak ve silecek"
echo "  - Banking Backend image'larını silecek"
echo "  - Kullanılmayan volume'ları silecek"
echo "  - Kullanılmayan network'leri silecek"
echo ""
read -p "Devam etmek istiyor musunuz? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    log_info "İşlem iptal edildi."
    exit 0
fi

# Docker'ın çalışıp çalışmadığını kontrol et
if ! docker info > /dev/null 2>&1; then
    log_error "Docker çalışmıyor. Lütfen Docker'ı başlatın."
    exit 1
fi

# Container'ları durdur ve sil
log_info "Container'lar durduruluyor ve siliniyor..."
docker-compose down --remove-orphans --volumes

# Banking Backend image'larını sil
log_info "Banking Backend image'ları siliniyor..."
docker images | grep banking-backend | awk '{print $3}' | xargs -r docker rmi -f

# Kullanılmayan image'ları sil
log_info "Kullanılmayan image'lar siliniyor..."
docker image prune -f

# Kullanılmayan container'ları sil
log_info "Kullanılmayan container'lar siliniyor..."
docker container prune -f

# Kullanılmayan volume'ları sil
log_info "Kullanılmayan volume'lar siliniyor..."
docker volume prune -f

# Kullanılmayan network'leri sil
log_info "Kullanılmayan network'ler siliniyor..."
docker network prune -f

# Sistem temizliği
log_info "Docker sistem temizliği yapılıyor..."
docker system prune -f

log_success "Temizlik işlemi tamamlandı!"
echo ""
echo "📊 Temizlik sonrası durum:"
echo "  - Container'lar: $(docker ps -a --format 'table {{.Names}}' | wc -l | tr -d ' ') adet"
echo "  - Image'lar: $(docker images --format 'table {{.Repository}}' | wc -l | tr -d ' ') adet"
echo "  - Volume'lar: $(docker volume ls --format 'table {{.Name}}' | wc -l | tr -d ' ') adet"
echo "  - Network'ler: $(docker network ls --format 'table {{.Name}}' | wc -l | tr -d ' ') adet"
