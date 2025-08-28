package v1

import (
	"net/http"
	"strconv"
	"time"

	"github.com/barannkoca/banking-backend/internal/interfaces"
	"github.com/barannkoca/banking-backend/internal/middleware"
	"github.com/barannkoca/banking-backend/internal/models"
	"github.com/barannkoca/banking-backend/internal/processing"
	"github.com/barannkoca/banking-backend/internal/services"
	"github.com/barannkoca/banking-backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TransactionHandler handles transaction-related requests
type TransactionHandler struct {
	transactionService *services.TransactionService
	balanceService     *services.BalanceService
	auditService       interfaces.AuditService
	workerPool         *processing.WorkerPool
}

// NewTransactionHandler creates a new TransactionHandler instance
func NewTransactionHandler(
	transactionService *services.TransactionService,
	balanceService *services.BalanceService,
	auditService interfaces.AuditService,
	workerPool *processing.WorkerPool,
) *TransactionHandler {
	return &TransactionHandler{
		transactionService: transactionService,
		balanceService:     balanceService,
		auditService:       auditService,
		workerPool:         workerPool,
	}
}

// CreditTransaction handles POST /api/v1/transactions/credit
func (h *TransactionHandler) CreditTransaction(c *gin.Context) {
	// Get current user from context
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Authentication required",
			"message": "Kimlik doğrulama gerekli",
		})
		return
	}

	// Parse request body
	var req models.DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Warn("Invalid credit request",
			zap.String("user_id", currentUserID.(string)),
			zap.String("ip", c.ClientIP()),
			zap.Error(err),
			zap.String("type", "credit_validation_error"),
		)

		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"message": "Geçersiz para yatırma verisi",
		})
		return
	}

	// Validate amount
	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid amount",
			"message": "Tutar sıfırdan büyük olmalıdır",
		})
		return
	}

	// Parse user ID
	userID, err := uuid.Parse(currentUserID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"message": "Geçersiz kullanıcı ID'si",
		})
		return
	}

	// Create transaction job
	job := &processing.TransactionJob{
		ID:                 uuid.New(),
		TransactionType:    "credit",
		ToAccountID:        userID,
		Amount:             req.Amount,
		TransactionService: h.transactionService,
		BalanceService:     h.balanceService,
		AuditService:       h.auditService,
		RetryCount:         0,
		MaxRetries:         3,
		CreatedAt:          time.Now(),
	}

	// Submit job to worker pool
	if err := h.workerPool.SubmitJob(job); err != nil {
		logger.GetLogger().Error("Failed to submit credit job",
			zap.String("user_id", userID.String()),
			zap.Float64("amount", req.Amount),
			zap.Error(err),
			zap.String("type", "credit_job_submit_error"),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to process credit transaction",
			"message": "Para yatırma işlemi başlatılamadı",
		})
		return
	}

	logger.GetLogger().Info("Credit transaction submitted",
		zap.String("job_id", job.ID.String()),
		zap.String("user_id", userID.String()),
		zap.Float64("amount", req.Amount),
		zap.String("ip", c.ClientIP()),
		zap.String("type", "credit_submitted"),
	)

	c.JSON(http.StatusAccepted, gin.H{
		"message":    "Para yatırma işlemi başlatıldı",
		"job_id":     job.ID.String(),
		"amount":     req.Amount,
		"status":     "processing",
		"created_at": job.CreatedAt,
	})
}

// DebitTransaction handles POST /api/v1/transactions/debit
func (h *TransactionHandler) DebitTransaction(c *gin.Context) {
	// Get current user from context
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Authentication required",
			"message": "Kimlik doğrulama gerekli",
		})
		return
	}

	// Parse request body
	var req models.WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Warn("Invalid debit request",
			zap.String("user_id", currentUserID.(string)),
			zap.String("ip", c.ClientIP()),
			zap.Error(err),
			zap.String("type", "debit_validation_error"),
		)

		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"message": "Geçersiz para çekme verisi",
		})
		return
	}

	// Validate amount
	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid amount",
			"message": "Tutar sıfırdan büyük olmalıdır",
		})
		return
	}

	// Parse user ID
	userID, err := uuid.Parse(currentUserID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"message": "Geçersiz kullanıcı ID'si",
		})
		return
	}

	// Check if user has sufficient balance
	canPerform, err := h.transactionService.CanPerformTransaction(c.Request.Context(), userID, req.Amount)
	if err != nil {
		logger.GetLogger().Error("Failed to check balance",
			zap.String("user_id", userID.String()),
			zap.Error(err),
			zap.String("type", "balance_check_error"),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to check balance",
			"message": "Bakiye kontrol edilemedi",
		})
		return
	}

	if !canPerform {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Insufficient balance",
			"message": "Yetersiz bakiye",
		})
		return
	}

	// Create transaction job
	job := &processing.TransactionJob{
		ID:                 uuid.New(),
		TransactionType:    "debit",
		FromAccountID:      userID,
		Amount:             req.Amount,
		TransactionService: h.transactionService,
		BalanceService:     h.balanceService,
		AuditService:       h.auditService,
		RetryCount:         0,
		MaxRetries:         3,
		CreatedAt:          time.Now(),
	}

	// Submit job to worker pool
	if err := h.workerPool.SubmitJob(job); err != nil {
		logger.GetLogger().Error("Failed to submit debit job",
			zap.String("user_id", userID.String()),
			zap.Float64("amount", req.Amount),
			zap.Error(err),
			zap.String("type", "debit_job_submit_error"),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to process debit transaction",
			"message": "Para çekme işlemi başlatılamadı",
		})
		return
	}

	logger.GetLogger().Info("Debit transaction submitted",
		zap.String("job_id", job.ID.String()),
		zap.String("user_id", userID.String()),
		zap.Float64("amount", req.Amount),
		zap.String("ip", c.ClientIP()),
		zap.String("type", "debit_submitted"),
	)

	c.JSON(http.StatusAccepted, gin.H{
		"message":    "Para çekme işlemi başlatıldı",
		"job_id":     job.ID.String(),
		"amount":     req.Amount,
		"status":     "processing",
		"created_at": job.CreatedAt,
	})
}

// TransferTransaction handles POST /api/v1/transactions/transfer
func (h *TransactionHandler) TransferTransaction(c *gin.Context) {
	// Get current user from context
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Authentication required",
			"message": "Kimlik doğrulama gerekli",
		})
		return
	}

	// Parse request body
	var req models.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Warn("Invalid transfer request",
			zap.String("user_id", currentUserID.(string)),
			zap.String("ip", c.ClientIP()),
			zap.Error(err),
			zap.String("type", "transfer_validation_error"),
		)

		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"message": "Geçersiz transfer verisi",
		})
		return
	}

	// Validate amount
	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid amount",
			"message": "Tutar sıfırdan büyük olmalıdır",
		})
		return
	}

	// Parse user IDs
	fromUserID, err := uuid.Parse(currentUserID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"message": "Geçersiz kullanıcı ID'si",
		})
		return
	}

	// Check if user is trying to transfer to themselves
	if fromUserID == req.ToUserID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Cannot transfer to same account",
			"message": "Kendinize transfer yapamazsınız",
		})
		return
	}

	// Check if user has sufficient balance
	canPerform, err := h.transactionService.CanPerformTransaction(c.Request.Context(), fromUserID, req.Amount)
	if err != nil {
		logger.GetLogger().Error("Failed to check balance",
			zap.String("user_id", fromUserID.String()),
			zap.Error(err),
			zap.String("type", "balance_check_error"),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to check balance",
			"message": "Bakiye kontrol edilemedi",
		})
		return
	}

	if !canPerform {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Insufficient balance",
			"message": "Yetersiz bakiye",
		})
		return
	}

	// Create transaction job
	job := &processing.TransactionJob{
		ID:                 uuid.New(),
		TransactionType:    "transfer",
		FromAccountID:      fromUserID,
		ToAccountID:        req.ToUserID,
		Amount:             req.Amount,
		TransactionService: h.transactionService,
		BalanceService:     h.balanceService,
		AuditService:       h.auditService,
		RetryCount:         0,
		MaxRetries:         3,
		CreatedAt:          time.Now(),
	}

	// Submit job to worker pool
	if err := h.workerPool.SubmitJob(job); err != nil {
		logger.GetLogger().Error("Failed to submit transfer job",
			zap.String("from_user_id", fromUserID.String()),
			zap.String("to_user_id", req.ToUserID.String()),
			zap.Float64("amount", req.Amount),
			zap.Error(err),
			zap.String("type", "transfer_job_submit_error"),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to process transfer transaction",
			"message": "Transfer işlemi başlatılamadı",
		})
		return
	}

	logger.GetLogger().Info("Transfer transaction submitted",
		zap.String("job_id", job.ID.String()),
		zap.String("from_user_id", fromUserID.String()),
		zap.String("to_user_id", req.ToUserID.String()),
		zap.Float64("amount", req.Amount),
		zap.String("ip", c.ClientIP()),
		zap.String("type", "transfer_submitted"),
	)

	c.JSON(http.StatusAccepted, gin.H{
		"message":      "Transfer işlemi başlatıldı",
		"job_id":       job.ID.String(),
		"from_user_id": fromUserID.String(),
		"to_user_id":   req.ToUserID.String(),
		"amount":       req.Amount,
		"status":       "processing",
		"created_at":   job.CreatedAt,
	})
}

// GetTransactionHistory handles GET /api/v1/transactions/history
func (h *TransactionHandler) GetTransactionHistory(c *gin.Context) {
	// Get current user from context
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Authentication required",
			"message": "Kimlik doğrulama gerekli",
		})
		return
	}

	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")
	transactionType := c.Query("type")
	status := c.Query("status")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 || limit > 100 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Parse user ID
	userID, err := uuid.Parse(currentUserID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"message": "Geçersiz kullanıcı ID'si",
		})
		return
	}

	// Get transaction history from service
	transactions, err := h.transactionService.GetTransactionHistory(c.Request.Context(), userID, limit, offset, transactionType, status)
	if err != nil {
		// Increment error count for performance monitoring
		middleware.IncrementErrorCount(c)

		logger.GetLogger().Error("Failed to get transaction history",
			zap.String("user_id", userID.String()),
			zap.Error(err),
			zap.String("type", "transaction_history_error"),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve transaction history",
			"message": "İşlem geçmişi alınamadı",
		})
		return
	}

	// Convert to response format
	var transactionResponses []*models.TransactionResponse
	for _, transaction := range transactions {
		transactionResponses = append(transactionResponses, transaction.ToResponse())
	}

	logger.GetLogger().Info("Transaction history retrieved",
		zap.String("user_id", userID.String()),
		zap.Int("count", len(transactionResponses)),
		zap.String("ip", c.ClientIP()),
		zap.String("type", "transaction_history_success"),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "İşlem geçmişi başarıyla getirildi",
		"data":    transactionResponses,
		"pagination": gin.H{
			"limit":  limit,
			"offset": offset,
			"count":  len(transactionResponses),
		},
	})
}

// GetTransaction handles GET /api/v1/transactions/{id}
func (h *TransactionHandler) GetTransaction(c *gin.Context) {
	// Get current user from context
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Authentication required",
			"message": "Kimlik doğrulama gerekli",
		})
		return
	}

	// Get transaction ID from URL parameter
	transactionIDStr := c.Param("id")
	transactionID, err := uuid.Parse(transactionIDStr)
	if err != nil {
		logger.GetLogger().Warn("Invalid transaction ID format",
			zap.String("transaction_id", transactionIDStr),
			zap.String("ip", c.ClientIP()),
			zap.Error(err),
			zap.String("type", "transaction_id_validation_error"),
		)

		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid transaction ID",
			"message": "Geçersiz işlem ID'si",
		})
		return
	}

	// Parse user ID
	userID, err := uuid.Parse(currentUserID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid user ID",
			"message": "Geçersiz kullanıcı ID'si",
		})
		return
	}

	// Get transaction from service
	transaction, err := h.transactionService.GetTransactionByID(c.Request.Context(), transactionID)
	if err != nil {
		logger.GetLogger().Error("Failed to get transaction",
			zap.String("transaction_id", transactionID.String()),
			zap.Error(err),
			zap.String("type", "transaction_get_error"),
		)

		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Transaction not found",
			"message": "İşlem bulunamadı",
		})
		return
	}

	// Check if user has access to this transaction
	if (transaction.FromUserID != nil && *transaction.FromUserID != userID) &&
		(transaction.ToUserID != nil && *transaction.ToUserID != userID) {
		logger.GetLogger().Warn("Unauthorized transaction access attempt",
			zap.String("user_id", userID.String()),
			zap.String("transaction_id", transactionID.String()),
			zap.String("ip", c.ClientIP()),
			zap.String("type", "transaction_access_unauthorized"),
		)

		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Access denied",
			"message": "Bu işleme erişim izniniz yok",
		})
		return
	}

	logger.GetLogger().Info("Transaction retrieved successfully",
		zap.String("transaction_id", transactionID.String()),
		zap.String("user_id", userID.String()),
		zap.String("ip", c.ClientIP()),
		zap.String("type", "transaction_get_success"),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "İşlem başarıyla getirildi",
		"data":    transaction.ToResponse(),
	})
}
