package processing

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/barannkoca/banking-backend/internal/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// ExampleUsage demonstrates how to use the processing system
func ExampleUsage() {
	// Initialize logger
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Create mock services (in real implementation, these would be actual service implementations)
	mockTransactionService := &MockTransactionService{}
	mockBalanceService := &MockBalanceService{}
	mockAuditService := &MockAuditService{}

	// Create worker pool directly (it starts automatically)
	workerPool := NewWorkerPool(5, 500, logger)

	// Example 1: Submit individual transactions
	submitIndividualTransactions(workerPool, mockTransactionService, mockBalanceService, mockAuditService, logger)

	// Example 2: Monitor statistics
	monitorStatistics(workerPool, logger)

	// Example 3: Health check
	performHealthCheck(workerPool, logger)

	// Wait for some processing to complete
	time.Sleep(5 * time.Second)

	// Shutdown gracefully
	if err := workerPool.Shutdown(10 * time.Second); err != nil {
		log.Printf("Shutdown hatası: %v", err)
	}
}

// submitIndividualTransactions demonstrates submitting individual transactions
func submitIndividualTransactions(workerPool *WorkerPool, transactionService *MockTransactionService, balanceService *MockBalanceService, auditService *MockAuditService, logger *zap.Logger) {
	logger.Info("Bireysel işlemler gönderiliyor...")

	// Create sample account IDs
	account1ID := uuid.New()
	account2ID := uuid.New()

	// Create transaction jobs and submit to worker pool
	transferJob := &TransactionJob{
		ID:                 uuid.New(),
		TransactionType:    "transfer",
		FromAccountID:      account1ID,
		ToAccountID:        account2ID,
		Amount:             100.50,
		TransactionService: transactionService,
		BalanceService:     balanceService,
		AuditService:       auditService,
		RetryCount:         0,
		MaxRetries:         3,
		CreatedAt:          time.Now(),
	}

	creditJob := &TransactionJob{
		ID:                 uuid.New(),
		TransactionType:    "credit",
		ToAccountID:        account1ID,
		Amount:             500.00,
		TransactionService: transactionService,
		BalanceService:     balanceService,
		AuditService:       auditService,
		RetryCount:         0,
		MaxRetries:         3,
		CreatedAt:          time.Now(),
	}

	debitJob := &TransactionJob{
		ID:                 uuid.New(),
		TransactionType:    "debit",
		FromAccountID:      account2ID,
		Amount:             75.25,
		TransactionService: transactionService,
		BalanceService:     balanceService,
		AuditService:       auditService,
		RetryCount:         0,
		MaxRetries:         3,
		CreatedAt:          time.Now(),
	}

	// Submit jobs to worker pool
	if err := workerPool.SubmitJob(transferJob); err != nil {
		logger.Error("Transfer işlemi gönderilemedi", zap.Error(err))
	}

	if err := workerPool.SubmitJob(creditJob); err != nil {
		logger.Error("Credit işlemi gönderilemedi", zap.Error(err))
	}

	if err := workerPool.SubmitJob(debitJob); err != nil {
		logger.Error("Debit işlemi gönderilemedi", zap.Error(err))
	}

	logger.Info("Bireysel işlemler gönderildi")
}

// monitorStatistics demonstrates monitoring system statistics
func monitorStatistics(workerPool *WorkerPool, logger *zap.Logger) {
	logger.Info("İstatistikler izleniyor...")

	// Get comprehensive statistics from atomic counters
	stats := workerPool.GetStatistics()

	// Log detailed transaction statistics
	logger.Info("Transaction İstatistikleri",
		zap.Int64("total_transactions", stats["total_transactions"].(int64)),
		zap.Int64("successful_transactions", stats["successful_transactions"].(int64)),
		zap.Int64("failed_transactions", stats["failed_transactions"].(int64)),
		zap.Float64("success_rate", stats["success_rate"].(float64)),
		zap.Float64("total_amount_processed", stats["total_amount_processed"].(float64)),
		zap.Float64("average_amount", stats["average_amount"].(float64)))

	// Log transaction type breakdown
	logger.Info("İşlem Türü Dağılımı",
		zap.Int64("transfer_count", stats["transfer_count"].(int64)),
		zap.Int64("deposit_count", stats["deposit_count"].(int64)),
		zap.Int64("withdraw_count", stats["withdraw_count"].(int64)))

	// Log performance metrics
	logger.Info("Performans Metrikleri",
		zap.Float64("average_processing_time_ms", stats["average_processing_time_ms"].(float64)),
		zap.Float64("fastest_transaction_ms", stats["fastest_transaction_ms"].(float64)),
		zap.Float64("slowest_transaction_ms", stats["slowest_transaction_ms"].(float64)))

	// Log error statistics
	logger.Info("Hata İstatistikleri",
		zap.Int64("validation_errors", stats["validation_errors"].(int64)),
		zap.Int64("insufficient_balance_errors", stats["insufficient_balance_errors"].(int64)),
		zap.Int64("system_errors", stats["system_errors"].(int64)),
		zap.Int64("retry_count", stats["retry_count"].(int64)))

	// Log system health
	logger.Info("Sistem Sağlığı",
		zap.Int("worker_count", stats["worker_count"].(int)),
		zap.Int("queue_length", stats["queue_length"].(int)),
		zap.Int("queue_capacity", stats["queue_capacity"].(int)))
}

// performHealthCheck demonstrates performing health checks
func performHealthCheck(workerPool *WorkerPool, logger *zap.Logger) {
	logger.Info("Sağlık kontrolü yapılıyor...")

	// Simple health check for worker pool
	queueLength := len(workerPool.jobQueue)
	queueCapacity := cap(workerPool.jobQueue)
	workerCount := len(workerPool.workers)

	status := "healthy"
	if queueLength >= queueCapacity*9/10 { // 90% full
		status = "warning"
	}
	if workerCount == 0 {
		status = "unhealthy"
	}

	logger.Info("Worker Pool sağlık kontrolü sonucu",
		zap.String("status", status),
		zap.Int("queue_length", queueLength),
		zap.Int("queue_capacity", queueCapacity),
		zap.Int("worker_count", workerCount),
		zap.Time("timestamp", time.Now()))

	if status != "healthy" {
		logger.Warn("Worker Pool sağlık durumu uyarı veriyor",
			zap.String("status", status),
			zap.Int("queue_length", queueLength),
			zap.Int("worker_count", workerCount))
	}
}

// Mock services for demonstration purposes

type MockTransactionService struct{}

func (m *MockTransactionService) Credit(ctx context.Context, accountID uuid.UUID, amount float64) error {
	// Simulate processing time
	time.Sleep(100 * time.Millisecond)
	return nil
}

func (m *MockTransactionService) Debit(ctx context.Context, accountID uuid.UUID, amount float64) error {
	// Simulate processing time
	time.Sleep(100 * time.Millisecond)
	return nil
}

func (m *MockTransactionService) Transfer(ctx context.Context, fromAccountID, toAccountID uuid.UUID, amount float64) error {
	// Simulate processing time
	time.Sleep(100 * time.Millisecond)
	return nil
}

type MockBalanceService struct{}

func (m *MockBalanceService) GetBalance(ctx context.Context, accountID uuid.UUID) (float64, error) {
	return 1000.0, nil
}

func (m *MockBalanceService) UpdateBalance(ctx context.Context, accountID uuid.UUID, amount float64) error {
	return nil
}

func (m *MockBalanceService) SafeUpdateBalance(ctx context.Context, accountID uuid.UUID, amount float64) error {
	return nil
}

func (m *MockBalanceService) GetBalanceHistory(ctx context.Context, accountID uuid.UUID) ([]models.BalanceHistory, error) {
	return []models.BalanceHistory{}, nil
}

func (m *MockBalanceService) CalculateAvailableBalance(ctx context.Context, accountID uuid.UUID) (float64, error) {
	return 1000.0, nil
}

type MockAuditService struct{}

func (m *MockAuditService) LogUserActivity(ctx context.Context, userID uuid.UUID, action, entityType, entityID, details string) error {
	return nil
}

func (m *MockAuditService) LogTransactionActivity(ctx context.Context, transaction *models.Transaction, action, details string) error {
	return nil
}

func (m *MockAuditService) LogSystemActivity(ctx context.Context, action, details string) error {
	return nil
}

func (m *MockAuditService) GetAuditLogsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.AuditLog, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAuditService) GetAuditLogsByEntityID(ctx context.Context, entityID string, limit, offset int) ([]*models.AuditLog, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAuditService) GetAuditLogsByAction(ctx context.Context, action string, limit, offset int) ([]*models.AuditLog, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAuditService) GetAllAuditLogs(ctx context.Context, limit, offset int) ([]*models.AuditLog, error) {
	return nil, fmt.Errorf("not implemented")
}

func (m *MockAuditService) DeleteOldAuditLogs(ctx context.Context, olderThanDays int) error {
	return fmt.Errorf("not implemented")
}

func (m *MockAuditService) GetAuditStatistics(ctx context.Context) (map[string]int64, error) {
	return nil, fmt.Errorf("not implemented")
}
