package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/barannkoca/banking-backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// AuthenticationMiddleware validates JWT tokens and sets user context
func AuthenticationMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Authorization header required",
				"message": "Please provide a valid JWT token",
			})
			c.Abort()
			return
		}

		// Check if it's a Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid authorization format",
				"message": "Authorization header must be in format: Bearer <token>",
			})
			c.Abort()
			return
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Empty token",
				"message": "JWT token cannot be empty",
			})
			c.Abort()
			return
		}

		// Validate JWT token (placeholder - implement actual JWT validation)
		userID, userRole, err := validateJWTToken(token)
		if err != nil {
			logger.GetLogger().Warn("Invalid JWT token",
				zap.String("token", token[:10]+"..."), // Log only first 10 chars for security
				zap.String("ip", c.ClientIP()),
				zap.Error(err),
				zap.String("type", "auth_error"),
			)

			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid token",
				"message": "JWT token is invalid or expired",
			})
			c.Abort()
			return
		}

		// Set user context
		c.Set("user_id", userID)
		c.Set("user_role", userRole)
		c.Set("authenticated", true)

		// Log successful authentication
		logger.GetLogger().Info("User authenticated",
			zap.String("user_id", userID),
			zap.String("user_role", userRole),
			zap.String("ip", c.ClientIP()),
			zap.String("type", "auth_success"),
		)

		c.Next()
	})
}

// AdminAuthorizationMiddleware checks if user has admin role
func AdminAuthorizationMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Check if user is authenticated
		if !isAuthenticated(c) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Authentication required",
				"message": "Please authenticate first",
			})
			c.Abort()
			return
		}

		// Check if user has admin role
		userRole := getUserRoleFromContext(c)
		if userRole != "admin" {
			logger.GetLogger().Warn("Unauthorized admin access attempt",
				zap.String("user_id", getUserIDFromContext(c)),
				zap.String("user_role", userRole),
				zap.String("ip", c.ClientIP()),
				zap.String("path", c.Request.URL.Path),
				zap.String("type", "auth_unauthorized"),
			)

			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Insufficient permissions",
				"message": "Admin role required for this operation",
			})
			c.Abort()
			return
		}

		c.Next()
	})
}

// ManagerAuthorizationMiddleware checks if user has manager or admin role
func ManagerAuthorizationMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Check if user is authenticated
		if !isAuthenticated(c) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Authentication required",
				"message": "Please authenticate first",
			})
			c.Abort()
			return
		}

		// Check if user has manager or admin role
		userRole := getUserRoleFromContext(c)
		if userRole != "manager" && userRole != "admin" {
			logger.GetLogger().Warn("Unauthorized manager access attempt",
				zap.String("user_id", getUserIDFromContext(c)),
				zap.String("user_role", userRole),
				zap.String("ip", c.ClientIP()),
				zap.String("path", c.Request.URL.Path),
				zap.String("type", "auth_unauthorized"),
			)

			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Insufficient permissions",
				"message": "Manager or admin role required for this operation",
			})
			c.Abort()
			return
		}

		c.Next()
	})
}

// validateJWTToken validates JWT token and returns user info (placeholder)
func validateJWTToken(token string) (userID string, userRole string, err error) {
	// TODO: Implement actual JWT validation
	// For now, return mock data
	if token == "valid-token" {
		return "user123", "customer", nil
	}
	if token == "admin-token" {
		return "admin123", "admin", nil
	}
	if token == "manager-token" {
		return "manager123", "manager", nil
	}

	return "", "", fmt.Errorf("invalid token")
}

// isAuthenticated checks if user is authenticated
func isAuthenticated(c *gin.Context) bool {
	authenticated, exists := c.Get("authenticated")
	if !exists {
		return false
	}
	return authenticated.(bool)
}

// getUserRoleFromContext extracts user role from Gin context
func getUserRoleFromContext(c *gin.Context) string {
	if userRole, exists := c.Get("user_role"); exists {
		if role, ok := userRole.(string); ok {
			return role
		}
	}
	return "anonymous"
}
