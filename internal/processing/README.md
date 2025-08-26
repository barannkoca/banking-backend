# 🚀 Transaction Processing System

Bu klasör, bankacılık backend'i için gelişmiş transaction processing sistemi içerir. Sistem, yüksek performanslı, ölçeklenebilir ve thread-safe transaction işleme yetenekleri sağlar.

## 📋 Özellikler

### 🔄 **Worker Pool**
- Çoklu worker goroutine'leri ile concurrent transaction processing
- Configurable worker sayısı ve queue boyutu
- Graceful shutdown desteği
- Retry mekanizması ile exponential backoff
- Real-time istatistikler



### ⚡ **Atomic Counters**
- Thread-safe transaction istatistikleri
- Real-time performance metrics
- Time-based analytics (hourly, daily, monthly)
- Error tracking ve categorization
- User activity monitoring



### 🎛️ **Processing Manager**
- Tüm bileşenleri yöneten merkezi manager
- Unified API for all processing operations
- Health monitoring ve diagnostics
- Statistics aggregation
- Graceful shutdown orchestration

## 🏗️ Mimari

```
ProcessingManager
├── WorkerPool
│   ├── Workers (goroutines)
│   ├── Job Queue
│   └── Results Channel
└── TransactionCounters
    ├── Atomic Counters
    ├── Performance Metrics
    └── Time-based Analytics
```

## 🚀 Kullanım

### Temel Kurulum

```go
// Logger oluştur
logger, _ := zap.NewDevelopment()

// Konfigürasyon
config := DefaultProcessingConfig()
config.WorkerCount = 10
config.MaxQueueSize = 1000

// Services (gerçek implementasyonlar)
transactionService := &YourTransactionService{}
balanceService := &YourBalanceService{}
auditService := &YourAuditService{}
userService := &YourUserService{}

// Processing Manager oluştur
manager := NewProcessingManager(
    config,
    transactionService,
    balanceService,
    auditService,
    userService,
    logger,
)

// Başlat
if err := manager.Start(); err != nil {
    log.Fatal(err)
}

// Graceful shutdown
defer manager.Shutdown()
```

### Transaction Gönderme

```go
// Bireysel transaction
transaction := &models.Transaction{
    ID:         uuid.New(),
    FromUserID: &fromUserID,
    ToUserID:   &toUserID,
    Amount:     100.50,
    Type:       models.TransactionTypeTransfer,
    Status:     models.TransactionStatusPending,
    Reference:  "Transfer",
    CreatedAt:  time.Now(),
}

// Transaction gönder
err := manager.SubmitTransaction(transaction)
```



### İstatistikler ve Monitoring

```go
// Comprehensive statistics
stats := manager.GetStatistics()

// Transaction-specific statistics
txStats := manager.GetTransactionStatistics()

// Queue status
queueStatus := manager.GetQueueStatus()

// Health check
health := manager.HealthCheck()
```

## ⚙️ Konfigürasyon

### ProcessingConfig

```go
type ProcessingConfig struct {
    // Worker Pool
    WorkerCount    int           // Worker goroutine sayısı
    MaxQueueSize   int           // Maksimum queue boyutu
    
    // Transaction Queue
    MaxHighPriority   int        // Yüksek öncelik queue boyutu
    MaxNormalPriority int        // Normal öncelik queue boyutu
    MaxLowPriority    int        // Düşük öncelik queue boyutu
    
    // Batch Processor
    MaxConcurrentTasks int       // Maksimum concurrent task sayısı
    MaxBatchSize       int       // Maksimum batch boyutu
    DefaultTimeout     time.Duration // Default timeout
    
    // General
    ShutdownTimeout time.Duration // Shutdown timeout
}
```

### Default Konfigürasyon

```go
config := DefaultProcessingConfig()
// WorkerCount: 10
// MaxQueueSize: 1000
// MaxHighPriority: 100
// MaxNormalPriority: 500
// MaxLowPriority: 1000
// MaxConcurrentTasks: 5
// MaxBatchSize: 100
// DefaultTimeout: 30s
// ShutdownTimeout: 60s
```

## 📊 İstatistikler

### Worker Pool İstatistikleri
- `processed_count`: İşlenen iş sayısı
- `failed_count`: Başarısız iş sayısı
- `active_workers`: Aktif worker sayısı
- `queue_size`: Kuyruk boyutu

### Transaction İstatistikleri
- `total_transactions`: Toplam transaction sayısı
- `successful_transactions`: Başarılı transaction sayısı
- `failed_transactions`: Başarısız transaction sayısı
- `success_rate`: Başarı oranı (%)
- `average_processing_time_ms`: Ortalama işlem süresi
- `total_amount_processed`: Toplam işlenen tutar



## 🔍 Health Check

Sistem sürekli olarak kendi sağlığını kontrol eder:

- **Worker Pool**: Aktif worker sayısı kontrolü
- **Queue**: Worker pool kuyruk boyutu kontrolü

- **Performance**: Processing time ve error rate kontrolü

## 🛡️ Thread Safety

Tüm bileşenler thread-safe olarak tasarlanmıştır:

- **Atomic Counters**: `sync/atomic` kullanımı
- **Mutex Protection**: Critical section'lar için `sync.RWMutex`
- **Channel Communication**: Goroutine'ler arası güvenli iletişim
- **Context Cancellation**: Graceful shutdown için context kullanımı

## 🔄 Retry Mekanizması

Sistem, başarısız işlemler için otomatik retry mekanizması içerir:

- **Exponential Backoff**: Her retry'da artan bekleme süresi
- **Max Retries**: Maksimum retry sayısı (default: 3)
- **Retry Count Tracking**: Her işlem için retry sayısı takibi
- **Priority Degradation**: Başarısız işlemler düşük önceliğe geçer

## 📈 Performance Optimizations



- **Concurrent Processing**: Paralel işlem yapma
- **Memory Management**: Efficient memory kullanımı
- **Connection Pooling**: Database connection pooling

## 🚨 Error Handling

- **Graceful Degradation**: Hata durumunda sistem çalışmaya devam eder
- **Error Categorization**: Hataları türlerine göre kategorize etme
- **Audit Logging**: Tüm hataları audit log'a kaydetme
- **Circuit Breaker**: Aşırı hata durumunda koruma

## 🔧 Monitoring ve Logging

- **Structured Logging**: Zap logger ile structured logging
- **Metrics Collection**: Real-time metrics toplama
- **Health Endpoints**: Health check endpoint'leri
- **Performance Monitoring**: Processing time ve throughput monitoring

## 📝 Örnek Kullanım

Detaylı örnek kullanım için `example_usage.go` dosyasına bakın. Bu dosya:

- Sistem kurulumu
- Transaction gönderme
- İstatistik monitoring
- Health check
- Mock service implementasyonları

içerir.

## 🤝 Katkıda Bulunma

1. Kod standartlarına uyun
2. Test coverage'ı koruyun
3. Documentation'ı güncelleyin
4. Performance impact'i değerlendirin
5. Thread safety'yi test edin

## 📄 Lisans

Bu proje MIT lisansı altında lisanslanmıştır.
