package processing

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/barannkoca/banking-backend/internal/interfaces"
	"github.com/barannkoca/banking-backend/internal/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// WorkerPool represents a pool of workers for processing transactions
type WorkerPool struct {
	workers      []*Worker
	jobQueue     chan *TransactionJob
	results      chan *TransactionResult
	workerCount  int
	maxQueueSize int
	shutdownChan chan struct{}
	wg           sync.WaitGroup
	ctx          context.Context
	cancel       context.CancelFunc
	logger       *zap.Logger

	// Atomic counters for detailed statistics
	counters *TransactionCounters
}

// Worker represents a single worker in the pool
type Worker struct {
	id       int
	pool     *WorkerPool
	jobChan  chan *TransactionJob
	stopChan chan struct{}
	wg       sync.WaitGroup
	logger   *zap.Logger
}

// TransactionJob represents a transaction processing job
type TransactionJob struct {
	ID                 uuid.UUID
	TransactionType    string // "credit", "debit", "transfer"
	FromAccountID      uuid.UUID
	ToAccountID        uuid.UUID
	Amount             float64
	TransactionService interfaces.TransactionService
	BalanceService     interfaces.BalanceService
	AuditService       interfaces.AuditService
	RetryCount         int
	MaxRetries         int
	CreatedAt          time.Time
}

// TransactionResult represents the result of processing a transaction
type TransactionResult struct {
	JobID          uuid.UUID
	Transaction    *models.Transaction
	Success        bool
	Error          error
	ProcessedAt    time.Time
	ProcessingTime time.Duration
	RetryCount     int
}

// NewWorkerPool creates a new worker pool for transaction processing
func NewWorkerPool(workerCount, maxQueueSize int, logger *zap.Logger) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	pool := &WorkerPool{
		workers:      make([]*Worker, workerCount),
		jobQueue:     make(chan *TransactionJob, maxQueueSize),
		results:      make(chan *TransactionResult, maxQueueSize),
		workerCount:  workerCount,
		maxQueueSize: maxQueueSize,
		shutdownChan: make(chan struct{}),
		ctx:          ctx,
		cancel:       cancel,
		logger:       logger,
		counters:     NewTransactionCounters(logger),
	}

	// Initialize workers
	for i := 0; i < workerCount; i++ {
		worker := &Worker{
			id:       i + 1,
			pool:     pool,
			jobChan:  make(chan *TransactionJob, 1),
			stopChan: make(chan struct{}),
			logger:   logger.With(zap.Int("worker_id", i+1)),
		}
		pool.workers[i] = worker
		pool.wg.Add(1)
		go worker.start()
	}

	// Start result processor
	go pool.processResults()

	logger.Info("Worker pool başlatıldı",
		zap.Int("worker_count", workerCount),
		zap.Int("max_queue_size", maxQueueSize))

	return pool
}

// SubmitJob submits a transaction job to the worker pool
func (wp *WorkerPool) SubmitJob(job *TransactionJob) error {
	select {
	case wp.jobQueue <- job:
		// Increment pending transactions counter
		wp.counters.IncrementPendingTransactions()
		wp.logger.Debug("İş kuyruğa eklendi",
			zap.String("job_id", job.ID.String()),
			zap.String("transaction_type", job.TransactionType))
		return nil
	case <-wp.ctx.Done():
		return fmt.Errorf("worker pool kapatılmış")
	default:
		return fmt.Errorf("iş kuyruğu dolu")
	}
}

// SubmitBatch submits multiple transaction jobs
func (wp *WorkerPool) SubmitBatch(jobs []*TransactionJob) error {
	for _, job := range jobs {
		if err := wp.SubmitJob(job); err != nil {
			return fmt.Errorf("iş gönderilirken hata: %w", err)
		}
	}
	return nil
}

// GetStatistics returns current pool statistics
func (wp *WorkerPool) GetStatistics() map[string]interface{} {
	// Get detailed transaction statistics from counters
	transactionStats := wp.counters.GetStatistics()

	// Add worker pool specific statistics
	poolStats := map[string]interface{}{
		"worker_count":   wp.workerCount,
		"max_queue_size": wp.maxQueueSize,
		"queue_length":   len(wp.jobQueue),
		"queue_capacity": cap(wp.jobQueue),
		"active_workers": len(wp.workers),
	}

	// Merge statistics
	for key, value := range poolStats {
		transactionStats[key] = value
	}

	return transactionStats
}

// Shutdown gracefully shuts down the worker pool
func (wp *WorkerPool) Shutdown(timeout time.Duration) error {
	wp.logger.Info("Worker pool kapatılıyor...")

	// Signal shutdown
	close(wp.shutdownChan)
	wp.cancel()

	// Wait for workers to finish with timeout
	done := make(chan struct{})
	go func() {
		wp.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		wp.logger.Info("Worker pool başarıyla kapatıldı")
		return nil
	case <-time.After(timeout):
		wp.logger.Warn("Worker pool kapatma zaman aşımı")
		return fmt.Errorf("worker pool kapatma zaman aşımı")
	}
}

// processResults processes results from workers
func (wp *WorkerPool) processResults() {
	for result := range wp.results {
		// Record transaction in atomic counters (simplified)
		wp.counters.RecordTransaction(nil, result.Success, result.ProcessingTime, result.Error)

		if !result.Success {
			wp.logger.Error("İşlem başarısız",
				zap.String("job_id", result.JobID.String()),
				zap.Error(result.Error),
				zap.Int("retry_count", result.RetryCount))
		} else {
			wp.logger.Info("İşlem başarılı",
				zap.String("job_id", result.JobID.String()),
				zap.Duration("processing_time", result.ProcessingTime))
		}
	}
}

// start starts a worker
func (w *Worker) start() {
	defer w.pool.wg.Done()
	defer close(w.jobChan)

	w.logger.Info("Worker başlatıldı")

	for {
		select {
		case job := <-w.pool.jobQueue:
			w.processJob(job)
		case <-w.stopChan:
			w.logger.Info("Worker durduruldu")
			return
		case <-w.pool.ctx.Done():
			w.logger.Info("Worker context iptal edildi")
			return
		}
	}
}

// processJob processes a single transaction job
func (w *Worker) processJob(job *TransactionJob) {
	startTime := time.Now()

	w.logger.Debug("İşlem işleniyor",
		zap.String("job_id", job.ID.String()),
		zap.String("transaction_type", job.TransactionType),
		zap.Float64("amount", job.Amount))

	// Process the transaction based on its type
	var err error
	switch job.TransactionType {
	case "credit":
		err = w.processCredit(job)
	case "debit":
		err = w.processDebit(job)
	case "transfer":
		err = w.processTransfer(job)
	default:
		err = fmt.Errorf("desteklenmeyen işlem türü: %s", job.TransactionType)
	}

	processingTime := time.Since(startTime)

	// Create result
	result := &TransactionResult{
		JobID:          job.ID,
		Transaction:    nil, // Artık transaction objesi yok
		Success:        err == nil,
		Error:          err,
		ProcessedAt:    time.Now(),
		ProcessingTime: processingTime,
		RetryCount:     job.RetryCount,
	}

	// Send result
	select {
	case w.pool.results <- result:
		// Result sent successfully
	default:
		w.logger.Warn("Sonuç kuyruğu dolu, sonuç atıldı",
			zap.String("job_id", job.ID.String()))
	}

	// Handle retry logic
	if err != nil && job.RetryCount < job.MaxRetries {
		w.handleRetry(job, err)
	}
}

// processCredit processes a credit transaction
func (w *Worker) processCredit(job *TransactionJob) error {
	ctx := context.Background()

	// Process the credit using the service
	err := job.TransactionService.Credit(ctx, job.ToAccountID, job.Amount)

	if err != nil {
		// Log audit trail for failed transaction
		job.AuditService.LogSystemActivity(ctx, "CREDIT_FAILED", fmt.Sprintf("Credit failed for account %s: %v", job.ToAccountID, err))
		return fmt.Errorf("credit işlemi başarısız: %w", err)
	}

	// Log audit trail for successful transaction
	job.AuditService.LogSystemActivity(ctx, "CREDIT_SUCCESS", fmt.Sprintf("Credit successful for account %s: %f", job.ToAccountID, job.Amount))

	return nil
}

// processDebit processes a debit transaction
func (w *Worker) processDebit(job *TransactionJob) error {
	ctx := context.Background()

	// Process the debit using the service
	err := job.TransactionService.Debit(ctx, job.FromAccountID, job.Amount)

	if err != nil {
		// Log audit trail for failed transaction
		job.AuditService.LogSystemActivity(ctx, "DEBIT_FAILED", fmt.Sprintf("Debit failed for account %s: %v", job.FromAccountID, err))
		return fmt.Errorf("debit işlemi başarısız: %w", err)
	}

	// Log audit trail for successful transaction
	job.AuditService.LogSystemActivity(ctx, "DEBIT_SUCCESS", fmt.Sprintf("Debit successful for account %s: %f", job.FromAccountID, job.Amount))

	return nil
}

// processTransfer processes a transfer transaction
func (w *Worker) processTransfer(job *TransactionJob) error {
	ctx := context.Background()

	// Process the transfer using the service
	err := job.TransactionService.Transfer(ctx, job.FromAccountID, job.ToAccountID, job.Amount)

	if err != nil {
		// Log audit trail for failed transaction
		job.AuditService.LogSystemActivity(ctx, "TRANSFER_FAILED", fmt.Sprintf("Transfer failed from %s to %s: %v", job.FromAccountID, job.ToAccountID, err))
		return fmt.Errorf("transfer işlemi başarısız: %w", err)
	}

	// Log audit trail for successful transaction
	job.AuditService.LogSystemActivity(ctx, "TRANSFER_SUCCESS", fmt.Sprintf("Transfer successful from %s to %s: %f", job.FromAccountID, job.ToAccountID, job.Amount))

	return nil
}

// handleRetry handles retry logic for failed jobs
func (w *Worker) handleRetry(job *TransactionJob, err error) {
	job.RetryCount++

	// Exponential backoff
	backoffDuration := time.Duration(job.RetryCount*job.RetryCount) * time.Second
	if backoffDuration > 30*time.Second {
		backoffDuration = 30 * time.Second
	}

	w.logger.Info("İşlem yeniden deneniyor",
		zap.String("job_id", job.ID.String()),
		zap.String("transaction_type", job.TransactionType),
		zap.Int("retry_count", job.RetryCount),
		zap.Duration("backoff_duration", backoffDuration),
		zap.Error(err))

	// Schedule retry
	time.AfterFunc(backoffDuration, func() {
		select {
		case w.pool.jobQueue <- job:
			// Increment retry counter
			w.pool.counters.IncrementRetryCount()
		case <-w.pool.ctx.Done():
			w.logger.Warn("Retry işlemi iptal edildi, worker pool kapatılıyor")
		}
	})
}

// Stop stops a specific worker
func (w *Worker) Stop() {
	close(w.stopChan)
}
