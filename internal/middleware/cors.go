package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware provides CORS configuration for the banking API
func CORSMiddleware() gin.HandlerFunc {
	config := cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "https://banking-frontend.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-Requested-With", "X-API-Key"},
		ExposeHeaders:    []string{"Content-Length", "X-Total-Count", "X-Rate-Limit-Remaining"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}

	return cors.New(config)
}

// BankingCORSMiddleware provides stricter CORS for banking operations
func BankingCORSMiddleware() gin.HandlerFunc {
	config := cors.Config{
		AllowOrigins:     []string{"https://secure-banking.com", "https://banking-app.com"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Requested-With", "X-API-Key", "X-Transaction-ID"},
		ExposeHeaders:    []string{"X-Transaction-ID", "X-Rate-Limit-Remaining", "X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           1 * time.Hour, // Shorter max age for security
	}

	return cors.New(config)
}
