package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter holds rate limiting configuration
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	r        rate.Limit
	b        int
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(r rate.Limit, b int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		r:        r,
		b:        b,
	}
}

// getLimiter returns the rate limiter for the given key
func (rl *RateLimiter) getLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(rl.r, rl.b)
		rl.limiters[key] = limiter
	}

	return limiter
}

// RateLimitMiddleware provides rate limiting based on IP address
func RateLimitMiddleware(requestsPerSecond float64, burst int) gin.HandlerFunc {
	limiter := NewRateLimiter(rate.Limit(requestsPerSecond), burst)

	return gin.HandlerFunc(func(c *gin.Context) {
		key := getClientKey(c)
		limiter := limiter.getLimiter(key)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "Rate limit exceeded",
				"message":     "Too many requests, please try again later",
				"retry_after": time.Now().Add(time.Second).Unix(),
			})
			c.Abort()
			return
		}

		// Add rate limit headers
		c.Header("X-Rate-Limit-Limit", fmt.Sprintf("%.0f", requestsPerSecond))
		c.Header("X-Rate-Limit-Remaining", fmt.Sprintf("%d", limiter.Burst()))
		c.Header("X-Rate-Limit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Second).Unix()))

		c.Next()
	})
}

// BankingRateLimitMiddleware provides stricter rate limiting for banking operations
func BankingRateLimitMiddleware() gin.HandlerFunc {
	// Stricter limits for banking operations
	return RateLimitMiddleware(5.0, 10) // 5 requests per second, burst of 10
}

// AuthenticationRateLimitMiddleware provides rate limiting for authentication endpoints
func AuthenticationRateLimitMiddleware() gin.HandlerFunc {
	// Very strict limits for authentication to prevent brute force
	return RateLimitMiddleware(1.0, 3) // 1 request per second, burst of 3
}

// getClientKey returns a unique key for the client (IP + User Agent)
func getClientKey(c *gin.Context) string {
	ip := c.ClientIP()
	userAgent := c.Request.UserAgent()
	return fmt.Sprintf("%s:%s", ip, userAgent)
}

// AdaptiveRateLimitMiddleware provides adaptive rate limiting based on user behavior
func AdaptiveRateLimitMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Check if user is authenticated
		userID := getUserIDFromContext(c)

		var requestsPerSecond float64
		var burst int

		if userID != "anonymous" {
			// Authenticated users get higher limits
			requestsPerSecond = 10.0
			burst = 20
		} else {
			// Anonymous users get lower limits
			requestsPerSecond = 2.0
			burst = 5
		}

		// Apply rate limiting
		RateLimitMiddleware(requestsPerSecond, burst)(c)
	})
}
