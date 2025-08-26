package api

import (
	"github.com/barannkoca/banking-backend/internal/middleware"
	"github.com/gin-gonic/gin"
)

// SetupRouter configures the main application router with all middleware and routes
func SetupRouter() *gin.Engine {
	// Create router with custom configuration
	r := gin.New()

	// Set Gin mode based on environment
	gin.SetMode(gin.ReleaseMode) // Production mode

	// Global middleware stack
	r.Use(gin.Recovery()) // Panic recovery

	// Security and CORS middleware
	r.Use(middleware.SecurityHeadersMiddleware())
	r.Use(middleware.CORSMiddleware())

	// Request tracking and logging
	r.Use(middleware.RequestTrackingMiddleware())
	r.Use(middleware.LoggingMiddleware())
	r.Use(middleware.SecurityLoggingMiddleware())

	// Rate limiting (global)
	r.Use(middleware.AdaptiveRateLimitMiddleware())

	// Health check endpoint (no rate limiting)
	healthGroup := r.Group("/health")
	{
		healthGroup.GET("", healthCheckHandler)
		healthGroup.GET("/ready", readinessCheckHandler)
		healthGroup.GET("/live", livenessCheckHandler)
	}

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Public routes (authentication)
		auth := v1.Group("/auth")
		auth.Use(middleware.AuthenticationRateLimitMiddleware()) // Stricter rate limiting for auth
		{
			auth.POST("/register", registerHandler)
			auth.POST("/login", loginHandler)
			auth.POST("/refresh", refreshTokenHandler)
		}

		// Protected routes (require authentication)
		protected := v1.Group("")
		protected.Use(middleware.AuthenticationMiddleware())         // JWT authentication
		protected.Use(middleware.BankingRateLimitMiddleware())       // Banking-specific rate limiting
		protected.Use(middleware.BankingSecurityHeadersMiddleware()) // Enhanced security for banking
		protected.Use(middleware.BankingTrackingMiddleware())        // Enhanced tracking for banking
		{
			// User Management Endpoints
			users := protected.Group("/users")
			{
				users.GET("", getUsersHandler)          // GET /api/v1/users
				users.GET("/:id", getUserHandler)       // GET /api/v1/users/{id}
				users.PUT("/:id", updateUserHandler)    // PUT /api/v1/users/{id}
				users.DELETE("/:id", deleteUserHandler) // DELETE /api/v1/users/{id}
			}

			// Transaction Endpoints
			transactions := protected.Group("/transactions")
			{
				transactions.POST("/credit", creditTransactionHandler)     // POST /api/v1/transactions/credit
				transactions.POST("/debit", debitTransactionHandler)       // POST /api/v1/transactions/debit
				transactions.POST("/transfer", transferTransactionHandler) // POST /api/v1/transactions/transfer
				transactions.GET("/history", getTransactionHistoryHandler) // GET /api/v1/transactions/history
				transactions.GET("/:id", getTransactionHandler)            // GET /api/v1/transactions/{id}
			}

			// Balance Endpoints
			balances := protected.Group("/balances")
			{
				balances.GET("/current", getCurrentBalanceHandler)       // GET /api/v1/balances/current
				balances.GET("/historical", getHistoricalBalanceHandler) // GET /api/v1/balances/historical
				balances.GET("/at-time", getBalanceAtTimeHandler)        // GET /api/v1/balances/at-time
			}
		}

		// Admin routes (require admin role)
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthenticationMiddleware())
		admin.Use(middleware.AdminAuthorizationMiddleware()) // Admin role check
		{
			admin.GET("/users", adminGetUsersHandler)
			admin.GET("/transactions", adminGetTransactionsHandler)
			admin.GET("/audit-logs", adminGetAuditLogsHandler)
			admin.POST("/system/maintenance", adminSystemMaintenanceHandler)
		}
	}

	// Documentation routes
	docs := r.Group("/docs")
	{
		docs.GET("", docsHandler)
		docs.GET("/swagger.json", swaggerJSONHandler)
	}

	return r
}

// Handler placeholders - these will be implemented in separate files
func healthCheckHandler(c *gin.Context) {
	c.JSON(200, gin.H{"status": "healthy", "service": "banking-backend"})
}

func readinessCheckHandler(c *gin.Context) {
	c.JSON(200, gin.H{"status": "ready", "service": "banking-backend"})
}

func livenessCheckHandler(c *gin.Context) {
	c.JSON(200, gin.H{"status": "alive", "service": "banking-backend"})
}

// Authentication Handlers
func registerHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message":     "User registration endpoint",
		"endpoint":    "POST /api/v1/auth/register",
		"description": "Register a new user account",
	})
}

func loginHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message":     "User login endpoint",
		"endpoint":    "POST /api/v1/auth/login",
		"description": "Authenticate user and return JWT token",
	})
}

func refreshTokenHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message":     "Token refresh endpoint",
		"endpoint":    "POST /api/v1/auth/refresh",
		"description": "Refresh JWT token",
	})
}

// User Management Handlers
func getUsersHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message":     "Get all users endpoint",
		"endpoint":    "GET /api/v1/users",
		"description": "Retrieve list of all users (admin only)",
	})
}

func getUserHandler(c *gin.Context) {
	userID := c.Param("id")
	c.JSON(200, gin.H{
		"message":     "Get user by ID endpoint",
		"endpoint":    "GET /api/v1/users/" + userID,
		"description": "Retrieve user information by ID",
		"user_id":     userID,
	})
}

func updateUserHandler(c *gin.Context) {
	userID := c.Param("id")
	c.JSON(200, gin.H{
		"message":     "Update user endpoint",
		"endpoint":    "PUT /api/v1/users/" + userID,
		"description": "Update user information",
		"user_id":     userID,
	})
}

func deleteUserHandler(c *gin.Context) {
	userID := c.Param("id")
	c.JSON(200, gin.H{
		"message":     "Delete user endpoint",
		"endpoint":    "DELETE /api/v1/users/" + userID,
		"description": "Delete user account",
		"user_id":     userID,
	})
}

// Transaction Handlers
func creditTransactionHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message":     "Credit transaction endpoint",
		"endpoint":    "POST /api/v1/transactions/credit",
		"description": "Add money to account (credit transaction)",
	})
}

func debitTransactionHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message":     "Debit transaction endpoint",
		"endpoint":    "POST /api/v1/transactions/debit",
		"description": "Remove money from account (debit transaction)",
	})
}

func transferTransactionHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message":     "Transfer transaction endpoint",
		"endpoint":    "POST /api/v1/transactions/transfer",
		"description": "Transfer money between accounts",
	})
}

func getTransactionHistoryHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message":     "Transaction history endpoint",
		"endpoint":    "GET /api/v1/transactions/history",
		"description": "Get transaction history for user",
	})
}

func getTransactionHandler(c *gin.Context) {
	transactionID := c.Param("id")
	c.JSON(200, gin.H{
		"message":        "Get transaction by ID endpoint",
		"endpoint":       "GET /api/v1/transactions/" + transactionID,
		"description":    "Retrieve specific transaction details",
		"transaction_id": transactionID,
	})
}

// Balance Handlers
func getCurrentBalanceHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message":     "Current balance endpoint",
		"endpoint":    "GET /api/v1/balances/current",
		"description": "Get current account balance",
	})
}

func getHistoricalBalanceHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message":     "Historical balance endpoint",
		"endpoint":    "GET /api/v1/balances/historical",
		"description": "Get historical balance data",
	})
}

func getBalanceAtTimeHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message":     "Balance at time endpoint",
		"endpoint":    "GET /api/v1/balances/at-time",
		"description": "Get account balance at specific time",
	})
}

func adminGetUsersHandler(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Admin get users endpoint"})
}

func adminGetTransactionsHandler(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Admin get transactions endpoint"})
}

func adminGetAuditLogsHandler(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Admin get audit logs endpoint"})
}

func adminSystemMaintenanceHandler(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Admin system maintenance endpoint"})
}

func docsHandler(c *gin.Context) {
	c.JSON(200, gin.H{"message": "API documentation"})
}

func swaggerJSONHandler(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Swagger JSON"})
}
