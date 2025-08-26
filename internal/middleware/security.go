package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

// SecurityHeadersMiddleware adds security headers to all responses
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Security Headers
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Content Security Policy for banking application
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' https:; connect-src 'self' https:; frame-ancestors 'none';")

		// Strict Transport Security (HSTS)
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")

		// Custom headers for banking security
		c.Header("X-Banking-Security", "enabled")
		c.Header("X-Request-ID", generateRequestID())

		c.Next()
	})
}

// BankingSecurityHeadersMiddleware adds additional security headers for banking operations
func BankingSecurityHeadersMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Standard security headers
		SecurityHeadersMiddleware()(c)

		// Additional banking-specific headers
		c.Header("X-Banking-Version", "1.0.0")
		c.Header("X-Transaction-Security", "enabled")
		c.Header("X-Fraud-Protection", "enabled")

		// Cache control for sensitive data
		c.Header("Cache-Control", "no-store, no-cache, must-revalidate, private")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")

		c.Next()
	})
}

// generateRequestID generates a unique request ID for tracking
func generateRequestID() string {
	// In production, use a proper UUID generator
	// For now, using a simple timestamp-based ID
	return "req-" + string(time.Now().UnixNano())
}
