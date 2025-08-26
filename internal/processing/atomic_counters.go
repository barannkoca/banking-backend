package processing

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/barannkoca/banking-backend/internal/models"
	"go.uber.org/zap"
)

// TransactionCounters represents atomic counters for transaction statistics
type TransactionCounters struct {
	// Total transaction counts
	totalTransactions      int64
	successfulTransactions int64
	failedTransactions     int64
	pendingTransactions    int64

	// Transaction type counts
	transferCount int64
	depositCount  int64
	withdrawCount int64

	// Amount tracking
	totalAmountProcessed int64 // Stored in cents to avoid floating point issues
	largestTransaction   int64
	smallestTransaction  int64

	// Performance metrics
	averageProcessingTime int64 // Stored in nanoseconds
	fastestTransaction    int64
	slowestTransaction    int64

	// Error tracking
	validationErrors          int64
	insufficientBalanceErrors int64
	systemErrors              int64

	// Retry tracking
	retryCount int64

	// Thread safety
	mutex sync.RWMutex

	logger *zap.Logger
}

// NewTransactionCounters creates a new transaction counters instance
func NewTransactionCounters(logger *zap.Logger) *TransactionCounters {
	return &TransactionCounters{
		smallestTransaction: 1<<63 - 1, // Max int64
		fastestTransaction:  1<<63 - 1, // Max int64
		logger:              logger,
	}
}

// IncrementTotalTransactions increments the total transaction count
func (tc *TransactionCounters) IncrementTotalTransactions() {
	atomic.AddInt64(&tc.totalTransactions, 1)
}

// IncrementSuccessfulTransactions increments the successful transaction count
func (tc *TransactionCounters) IncrementSuccessfulTransactions() {
	atomic.AddInt64(&tc.successfulTransactions, 1)
}

// IncrementFailedTransactions increments the failed transaction count
func (tc *TransactionCounters) IncrementFailedTransactions() {
	atomic.AddInt64(&tc.failedTransactions, 1)
}

// IncrementPendingTransactions increments the pending transaction count
func (tc *TransactionCounters) IncrementPendingTransactions() {
	atomic.AddInt64(&tc.pendingTransactions, 1)
}

// DecrementPendingTransactions decrements the pending transaction count
func (tc *TransactionCounters) DecrementPendingTransactions() {
	atomic.AddInt64(&tc.pendingTransactions, -1)
}

// IncrementTransactionType increments the count for a specific transaction type
func (tc *TransactionCounters) IncrementTransactionType(transactionType models.TransactionType) {
	switch transactionType {
	case models.TransactionTypeTransfer:
		atomic.AddInt64(&tc.transferCount, 1)
	case models.TransactionTypeDeposit:
		atomic.AddInt64(&tc.depositCount, 1)
	case models.TransactionTypeWithdraw:
		atomic.AddInt64(&tc.withdrawCount, 1)
	}
}

// AddAmountProcessed adds an amount to the total processed amount
func (tc *TransactionCounters) AddAmountProcessed(amount float64) {
	amountInCents := int64(amount * 100)
	atomic.AddInt64(&tc.totalAmountProcessed, amountInCents)

	// Update largest transaction
	for {
		current := atomic.LoadInt64(&tc.largestTransaction)
		if amountInCents <= current {
			break
		}
		if atomic.CompareAndSwapInt64(&tc.largestTransaction, current, amountInCents) {
			break
		}
	}

	// Update smallest transaction
	for {
		current := atomic.LoadInt64(&tc.smallestTransaction)
		if amountInCents >= current {
			break
		}
		if atomic.CompareAndSwapInt64(&tc.smallestTransaction, current, amountInCents) {
			break
		}
	}
}

// RecordProcessingTime records the processing time for a transaction
func (tc *TransactionCounters) RecordProcessingTime(duration time.Duration) {
	durationNs := duration.Nanoseconds()

	// Update average processing time
	for {
		current := atomic.LoadInt64(&tc.averageProcessingTime)
		total := atomic.LoadInt64(&tc.successfulTransactions)
		if total == 0 {
			newAverage := durationNs
			if atomic.CompareAndSwapInt64(&tc.averageProcessingTime, 0, newAverage) {
				break
			}
		} else {
			newAverage := (current*total + durationNs) / (total + 1)
			if atomic.CompareAndSwapInt64(&tc.averageProcessingTime, current, newAverage) {
				break
			}
		}
	}

	// Update fastest transaction
	for {
		current := atomic.LoadInt64(&tc.fastestTransaction)
		if durationNs >= current {
			break
		}
		if atomic.CompareAndSwapInt64(&tc.fastestTransaction, current, durationNs) {
			break
		}
	}

	// Update slowest transaction
	for {
		current := atomic.LoadInt64(&tc.slowestTransaction)
		if durationNs <= current {
			break
		}
		if atomic.CompareAndSwapInt64(&tc.slowestTransaction, current, durationNs) {
			break
		}
	}
}

// IncrementErrorType increments the count for a specific error type
func (tc *TransactionCounters) IncrementErrorType(errorType string) {
	switch errorType {
	case "validation":
		atomic.AddInt64(&tc.validationErrors, 1)
	case "insufficient_balance":
		atomic.AddInt64(&tc.insufficientBalanceErrors, 1)
	case "system":
		atomic.AddInt64(&tc.systemErrors, 1)
	}
}

// IncrementRetryCount increments the retry count
func (tc *TransactionCounters) IncrementRetryCount() {
	atomic.AddInt64(&tc.retryCount, 1)
}

// GetStatistics returns all current statistics
func (tc *TransactionCounters) GetStatistics() map[string]interface{} {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()

	// Calculate success rate
	total := atomic.LoadInt64(&tc.totalTransactions)
	successful := atomic.LoadInt64(&tc.successfulTransactions)
	successRate := 0.0
	if total > 0 {
		successRate = float64(successful) / float64(total) * 100
	}

	// Calculate average amount
	totalAmount := atomic.LoadInt64(&tc.totalAmountProcessed)
	averageAmount := 0.0
	if successful > 0 {
		averageAmount = float64(totalAmount) / float64(successful) / 100.0
	}

	return map[string]interface{}{
		// Transaction counts
		"total_transactions":      atomic.LoadInt64(&tc.totalTransactions),
		"successful_transactions": atomic.LoadInt64(&tc.successfulTransactions),
		"failed_transactions":     atomic.LoadInt64(&tc.failedTransactions),
		"pending_transactions":    atomic.LoadInt64(&tc.pendingTransactions),
		"success_rate":            successRate,

		// Transaction type counts
		"transfer_count": atomic.LoadInt64(&tc.transferCount),
		"deposit_count":  atomic.LoadInt64(&tc.depositCount),
		"withdraw_count": atomic.LoadInt64(&tc.withdrawCount),

		// Amount statistics
		"total_amount_processed": float64(totalAmount) / 100.0,
		"average_amount":         averageAmount,
		"largest_transaction":    float64(atomic.LoadInt64(&tc.largestTransaction)) / 100.0,
		"smallest_transaction":   float64(atomic.LoadInt64(&tc.smallestTransaction)) / 100.0,

		// Performance metrics
		"average_processing_time_ms": float64(atomic.LoadInt64(&tc.averageProcessingTime)) / 1000000.0,
		"fastest_transaction_ms":     float64(atomic.LoadInt64(&tc.fastestTransaction)) / 1000000.0,
		"slowest_transaction_ms":     float64(atomic.LoadInt64(&tc.slowestTransaction)) / 1000000.0,

		// Error counts
		"validation_errors":           atomic.LoadInt64(&tc.validationErrors),
		"insufficient_balance_errors": atomic.LoadInt64(&tc.insufficientBalanceErrors),
		"system_errors":               atomic.LoadInt64(&tc.systemErrors),

		// Retry count
		"retry_count": atomic.LoadInt64(&tc.retryCount),
	}
}

// Reset resets all counters to zero
func (tc *TransactionCounters) Reset() {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	// Reset all atomic counters
	atomic.StoreInt64(&tc.totalTransactions, 0)
	atomic.StoreInt64(&tc.successfulTransactions, 0)
	atomic.StoreInt64(&tc.failedTransactions, 0)
	atomic.StoreInt64(&tc.pendingTransactions, 0)

	atomic.StoreInt64(&tc.transferCount, 0)
	atomic.StoreInt64(&tc.depositCount, 0)
	atomic.StoreInt64(&tc.withdrawCount, 0)

	atomic.StoreInt64(&tc.totalAmountProcessed, 0)
	atomic.StoreInt64(&tc.largestTransaction, 0)
	atomic.StoreInt64(&tc.smallestTransaction, 1<<63-1)

	atomic.StoreInt64(&tc.averageProcessingTime, 0)
	atomic.StoreInt64(&tc.fastestTransaction, 1<<63-1)
	atomic.StoreInt64(&tc.slowestTransaction, 0)

	atomic.StoreInt64(&tc.validationErrors, 0)
	atomic.StoreInt64(&tc.insufficientBalanceErrors, 0)
	atomic.StoreInt64(&tc.systemErrors, 0)

	atomic.StoreInt64(&tc.retryCount, 0)

	tc.logger.Info("Tüm transaction sayaçları sıfırlandı")
}

// RecordTransaction records a complete transaction with all its metrics
func (tc *TransactionCounters) RecordTransaction(transaction *models.Transaction, success bool, processingTime time.Duration, err error) {
	// Increment total transactions
	tc.IncrementTotalTransactions()

	// Increment transaction type
	tc.IncrementTransactionType(transaction.Type)

	// Record amount
	tc.AddAmountProcessed(transaction.Amount)

	// Record processing time
	tc.RecordProcessingTime(processingTime)

	if success {
		tc.IncrementSuccessfulTransactions()
		tc.DecrementPendingTransactions()
	} else {
		tc.IncrementFailedTransactions()
		tc.DecrementPendingTransactions()

		// Record error type
		if err != nil {
			errorType := "system"
			if err.Error() == "yetersiz bakiye" {
				errorType = "insufficient_balance"
			} else if err.Error() == "işlem doğrulama hatası" {
				errorType = "validation"
			}
			tc.IncrementErrorType(errorType)
		}
	}
}
