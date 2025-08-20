package middleware

import (
	"time"

	"github.com/barannkoca/banking-backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoggingMiddleware logs all HTTP requests with detailed information
func LoggingMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		startTime := time.Now()

		// Process request
		c.Next()

		// Calculate request duration
		duration := time.Since(startTime)

		// Get user ID from context (will be set after authentication)
		userID := getUserIDFromContext(c)

		// Log the request
		logger.LogAPIRequest(
			c.Request.Method,
			c.Request.URL.Path,
			userID,
			c.ClientIP(),
			c.Writer.Status(),
			duration.Milliseconds(),
		)

		// Log additional details for banking operations
		if isBankingEndpoint(c.Request.URL.Path) {
			logger.GetLogger().Info("Banking API Request",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("user_id", userID),
				zap.String("ip", c.ClientIP()),
				zap.Int("status", c.Writer.Status()),
				zap.Duration("duration", duration),
				zap.String("user_agent", c.Request.UserAgent()),
				zap.String("type", "banking_api"),
			)
		}

		// Log slow requests (>1 second)
		if duration > time.Second {
			logger.GetLogger().Warn("Slow Request Detected",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Duration("duration", duration),
				zap.String("ip", c.ClientIP()),
				zap.String("type", "performance"),
			)
		}

		// Log errors
		if c.Writer.Status() >= 400 {
			logger.GetLogger().Error("HTTP Error",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Int("status", c.Writer.Status()),
				zap.String("ip", c.ClientIP()),
				zap.String("user_id", userID),
				zap.String("type", "http_error"),
			)
		}
	})
}

// getUserIDFromContext extracts user ID from Gin context
// This will be set by authentication middleware
func getUserIDFromContext(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if uid, ok := userID.(string); ok {
			return uid
		}
	}
	return "anonymous"
}

// isBankingEndpoint checks if the endpoint is related to banking operations
func isBankingEndpoint(path string) bool {
	bankingPaths := []string{
		"/api/v1/accounts",
		"/api/v1/transactions",
		"/api/v1/transfers",
		"/api/v1/payments",
		"/api/v1/loans",
	}

	for _, bankingPath := range bankingPaths {
		if len(path) >= len(bankingPath) && path[:len(bankingPath)] == bankingPath {
			return true
		}
	}
	return false
}

// SecurityLoggingMiddleware logs security-related events
func SecurityLoggingMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Log suspicious patterns
		userAgent := c.Request.UserAgent()
		ip := c.ClientIP()

		// Check for suspicious user agents
		if isSuspiciousUserAgent(userAgent) {
			logger.LogSecurity(
				"suspicious_user_agent",
				getUserIDFromContext(c),
				"Suspicious user agent detected: "+userAgent,
				"medium",
			)
		}

		// Check for rapid requests (basic rate limiting check)
		if isRapidRequest(c) {
			logger.LogSecurity(
				"rapid_requests",
				getUserIDFromContext(c),
				"Rapid requests detected from IP: "+ip,
				"high",
			)
		}

		c.Next()
	})
}

// isSuspiciousUserAgent checks for known malicious user agents
func isSuspiciousUserAgent(userAgent string) bool {
	suspiciousAgents := []string{
		"sqlmap",
		"nmap",
		"masscan",
		"python-requests", // Too generic, might need adjustment
	}

	for _, suspicious := range suspiciousAgents {
		if len(userAgent) >= len(suspicious) {
			for i := 0; i <= len(userAgent)-len(suspicious); i++ {
				if userAgent[i:i+len(suspicious)] == suspicious {
					return true
				}
			}
		}
	}
	return false
}

// isRapidRequest checks for rapid requests (basic implementation)
// In production, you'd use Redis or similar for rate limiting
func isRapidRequest(c *gin.Context) bool {
	// This is a placeholder - implement proper rate limiting
	// For now, just return false
	return false
}
