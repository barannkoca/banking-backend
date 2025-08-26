package middleware

import (
	"fmt"
	"time"

	"github.com/barannkoca/banking-backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// RequestTrackingMiddleware adds request tracking and correlation IDs
func RequestTrackingMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		startTime := time.Now()

		// Generate unique request ID
		requestID := uuid.New().String()
		c.Header("X-Request-ID", requestID)

		// Set request ID in context for logging
		c.Set("request_id", requestID)
		c.Set("start_time", startTime)

		// Add correlation ID if provided
		if correlationID := c.GetHeader("X-Correlation-ID"); correlationID != "" {
			c.Set("correlation_id", correlationID)
			c.Header("X-Correlation-ID", correlationID)
		}

		// Log request start
		logger.GetLogger().Info("Request Started",
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.String("type", "request_start"),
		)

		c.Next()

		// Log request completion
		duration := time.Since(startTime)
		status := c.Writer.Status()

		logger.GetLogger().Info("Request Completed",
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", status),
			zap.Duration("duration", duration),
			zap.String("type", "request_complete"),
		)

		// Add performance headers
		c.Header("X-Response-Time", duration.String())
		c.Header("X-Request-Duration", fmt.Sprintf("%d", duration.Milliseconds()))
	})
}

// BankingTrackingMiddleware provides enhanced tracking for banking operations
func BankingTrackingMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Apply basic tracking
		RequestTrackingMiddleware()(c)

		// Add banking-specific tracking
		userID := getUserIDFromContext(c)
		requestID, _ := c.Get("request_id")

		// Track sensitive operations
		if isSensitiveOperation(c.Request.URL.Path) {
			logger.GetLogger().Info("Sensitive Banking Operation",
				zap.String("request_id", requestID.(string)),
				zap.String("user_id", userID),
				zap.String("operation", c.Request.URL.Path),
				zap.String("ip", c.ClientIP()),
				zap.String("type", "sensitive_operation"),
			)
		}

		// Add transaction tracking headers
		c.Header("X-Banking-Operation", c.Request.URL.Path)
		c.Header("X-User-ID", userID)

		c.Next()
	})
}

// isSensitiveOperation checks if the operation is sensitive
func isSensitiveOperation(path string) bool {
	sensitivePaths := []string{
		"/api/v1/transfers",
		"/api/v1/payments",
		"/api/v1/accounts/balance",
		"/api/v1/transactions",
		"/api/v1/loans",
	}

	for _, sensitivePath := range sensitivePaths {
		if len(path) >= len(sensitivePath) && path[:len(sensitivePath)] == sensitivePath {
			return true
		}
	}
	return false
}
