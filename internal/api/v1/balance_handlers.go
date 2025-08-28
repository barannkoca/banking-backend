package v1

import (
	"net/http"
	"strconv"
	"time"

	"github.com/barannkoca/banking-backend/internal/middleware"
	"github.com/barannkoca/banking-backend/internal/models"
	"github.com/barannkoca/banking-backend/internal/services"
	"github.com/barannkoca/banking-backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// BalanceHandler handles balance-related requests
type BalanceHandler struct {
	balanceService *services.BalanceService
}

// NewBalanceHandler creates a new BalanceHandler instance
func NewBalanceHandler(balanceService *services.BalanceService) *BalanceHandler {
	return &BalanceHandler{
		balanceService: balanceService,
	}
}

// GetCurrentBalance handles GET /api/v1/balances/current
func (h *BalanceHandler) GetCurrentBalance(c *gin.Context) {
	// Get current user from context
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Authentication required",
			"message": "Kimlik doğrulama gerekli",
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

	// Get current balance
	balance, err := h.balanceService.GetBalance(c.Request.Context(), userID)
	if err != nil {
		// Increment error count for performance monitoring
		middleware.IncrementErrorCount(c)

		logger.GetLogger().Error("Failed to get current balance",
			zap.String("user_id", userID.String()),
			zap.Error(err),
			zap.String("type", "balance_get_error"),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve balance",
			"message": "Bakiye alınamadı",
		})
		return
	}

	// Get cache statistics for performance monitoring
	if metrics := middleware.GetPerformanceMetrics(c); metrics != nil {
		// Cache hit/miss will be automatically tracked by Redis service
		// We can add additional cache-specific metrics here if needed
	}

	// Get available balance (considering holds, pending transactions, etc.)
	availableBalance, err := h.balanceService.CalculateAvailableBalance(c.Request.Context(), userID)
	if err != nil {
		logger.GetLogger().Error("Failed to calculate available balance",
			zap.String("user_id", userID.String()),
			zap.Error(err),
			zap.String("type", "available_balance_calc_error"),
		)

		// Don't fail the request, just use current balance
		availableBalance = balance
	}

	logger.GetLogger().Info("Current balance retrieved",
		zap.String("user_id", userID.String()),
		zap.Float64("balance", balance),
		zap.Float64("available_balance", availableBalance),
		zap.String("ip", c.ClientIP()),
		zap.String("type", "balance_get_success"),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Mevcut bakiye başarıyla getirildi",
		"data": gin.H{
			"user_id":           userID.String(),
			"current_balance":   balance,
			"available_balance": availableBalance,
			"currency":          "TRY",
			"last_updated":      time.Now(),
		},
	})
}

// GetHistoricalBalance handles GET /api/v1/balances/historical
func (h *BalanceHandler) GetHistoricalBalance(c *gin.Context) {
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
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

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

	// Parse date range if provided
	var startDate, endDate *time.Time
	if startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = &parsed
		}
	}
	if endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = &parsed
		}
	}

	// Get balance history
	history, err := h.balanceService.GetBalanceHistory(c.Request.Context(), userID)
	if err != nil {
		logger.GetLogger().Error("Failed to get balance history",
			zap.String("user_id", userID.String()),
			zap.Error(err),
			zap.String("type", "balance_history_error"),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to retrieve balance history",
			"message": "Bakiye geçmişi alınamadı",
		})
		return
	}

	// Filter by date range if provided
	var filteredHistory []models.BalanceHistory
	for _, entry := range history {
		if startDate != nil && entry.CreatedAt.Before(*startDate) {
			continue
		}
		if endDate != nil && entry.CreatedAt.After(*endDate) {
			continue
		}
		filteredHistory = append(filteredHistory, entry)
	}

	// Apply pagination
	start := offset
	end := start + limit
	if start >= len(filteredHistory) {
		filteredHistory = []models.BalanceHistory{}
	} else if end > len(filteredHistory) {
		filteredHistory = filteredHistory[start:]
	} else {
		filteredHistory = filteredHistory[start:end]
	}

	logger.GetLogger().Info("Balance history retrieved",
		zap.String("user_id", userID.String()),
		zap.Int("count", len(filteredHistory)),
		zap.String("ip", c.ClientIP()),
		zap.String("type", "balance_history_success"),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Bakiye geçmişi başarıyla getirildi",
		"data": gin.H{
			"user_id": userID.String(),
			"history": filteredHistory,
			"pagination": gin.H{
				"limit":  limit,
				"offset": offset,
				"count":  len(filteredHistory),
				"total":  len(history),
			},
		},
	})
}

// GetBalanceAtTime handles GET /api/v1/balances/at-time
func (h *BalanceHandler) GetBalanceAtTime(c *gin.Context) {
	// Get current user from context
	currentUserID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Authentication required",
			"message": "Kimlik doğrulama gerekli",
		})
		return
	}

	// Get timestamp parameter
	timestampStr := c.Query("timestamp")
	if timestampStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Timestamp parameter required",
			"message": "Zaman damgası parametresi gerekli",
		})
		return
	}

	// Parse timestamp
	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		// Try alternative formats
		formats := []string{"2006-01-02T15:04:05Z", "2006-01-02 15:04:05", "2006-01-02"}
		parsed := false

		for _, format := range formats {
			if t, err := time.Parse(format, timestampStr); err == nil {
				timestamp = t
				parsed = true
				break
			}
		}

		if !parsed {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid timestamp format",
				"message": "Geçersiz zaman damgası formatı",
			})
			return
		}
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

	// Get balance history to find balance at specific time
	history, err := h.balanceService.GetBalanceHistory(c.Request.Context(), userID)
	if err != nil {
		logger.GetLogger().Error("Failed to get balance history for time calculation",
			zap.String("user_id", userID.String()),
			zap.Time("timestamp", timestamp),
			zap.Error(err),
			zap.String("type", "balance_at_time_error"),
		)

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to calculate balance at time",
			"message": "Belirtilen zamandaki bakiye hesaplanamadı",
		})
		return
	}

	// Find the balance at the specified time
	// This is a simplified implementation - in a real system, you'd have a more sophisticated
	// way to calculate historical balances
	var balanceAtTime float64
	var found bool

	// Sort history by creation time and find the closest entry before the timestamp
	for _, entry := range history {
		if entry.CreatedAt.Before(timestamp) || entry.CreatedAt.Equal(timestamp) {
			balanceAtTime = entry.NewAmount
			found = true
		}
	}

	// If no historical data found, return current balance
	if !found {
		currentBalance, err := h.balanceService.GetBalance(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to get current balance",
				"message": "Mevcut bakiye alınamadı",
			})
			return
		}
		balanceAtTime = currentBalance
	}

	logger.GetLogger().Info("Balance at time retrieved",
		zap.String("user_id", userID.String()),
		zap.Time("timestamp", timestamp),
		zap.Float64("balance", balanceAtTime),
		zap.String("ip", c.ClientIP()),
		zap.String("type", "balance_at_time_success"),
	)

	c.JSON(http.StatusOK, gin.H{
		"message": "Belirtilen zamandaki bakiye başarıyla getirildi",
		"data": gin.H{
			"user_id":    userID.String(),
			"timestamp":  timestamp,
			"balance":    balanceAtTime,
			"currency":   "TRY",
			"calculated": found,
		},
	})
}
