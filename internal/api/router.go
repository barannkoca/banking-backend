package api

import (
	"time"

	v1 "github.com/barannkoca/banking-backend/internal/api/v1"
	"github.com/barannkoca/banking-backend/internal/interfaces"
	"github.com/barannkoca/banking-backend/internal/middleware"
	"github.com/barannkoca/banking-backend/internal/processing"
	"github.com/barannkoca/banking-backend/internal/services"
	"github.com/gin-gonic/gin"
)

// SetupRouter configures the main application router with all middleware and routes
func SetupRouter(
	userService *services.UserService,
	transactionService *services.TransactionService,
	balanceService *services.BalanceService,
	auditService interfaces.AuditService,
	workerPool *processing.WorkerPool,
) *gin.Engine {
	// Create router with custom configuration
	r := gin.New()

	// Set Gin mode based on environment
	gin.SetMode(gin.ReleaseMode) // Production mode

	// Initialize handlers
	authHandler := v1.NewAuthHandler(userService)
	userHandler := v1.NewUserHandler(userService)
	transactionHandler := v1.NewTransactionHandler(transactionService, balanceService, auditService, workerPool)
	balanceHandler := v1.NewBalanceHandler(balanceService)

	// Global middleware stack
	r.Use(gin.Recovery()) // Panic recovery

	// Security and CORS middleware
	r.Use(middleware.SecurityHeadersMiddleware())
	r.Use(middleware.CORSMiddleware())

	// Request tracking and logging
	r.Use(middleware.RequestTrackingMiddleware())
	r.Use(middleware.LoggingMiddleware())
	r.Use(middleware.SecurityLoggingMiddleware())
	r.Use(middleware.PerformanceMonitorMiddleware())

	// Rate limiting (global)
	r.Use(middleware.AdaptiveRateLimitMiddleware())

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Public routes (authentication)
		auth := v1.Group("/auth")
		auth.Use(middleware.AuthenticationRateLimitMiddleware()) // Stricter rate limiting for auth
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
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
				users.GET("", userHandler.GetUsers)          // GET /api/v1/users
				users.GET("/:id", userHandler.GetUser)       // GET /api/v1/users/{id}
				users.PUT("/:id", userHandler.UpdateUser)    // PUT /api/v1/users/{id}
				users.DELETE("/:id", userHandler.DeleteUser) // DELETE /api/v1/users/{id}
			}

			// Transaction Endpoints
			transactions := protected.Group("/transactions")
			{
				transactions.POST("/credit", transactionHandler.CreditTransaction)     // POST /api/v1/transactions/credit
				transactions.POST("/debit", transactionHandler.DebitTransaction)       // POST /api/v1/transactions/debit
				transactions.POST("/transfer", transactionHandler.TransferTransaction) // POST /api/v1/transactions/transfer
				transactions.GET("/history", transactionHandler.GetTransactionHistory) // GET /api/v1/transactions/history
				transactions.GET("/:id", transactionHandler.GetTransaction)            // GET /api/v1/transactions/{id}
			}

			// Balance Endpoints
			balances := protected.Group("/balances")
			{
				balances.GET("/current", balanceHandler.GetCurrentBalance)       // GET /api/v1/balances/current
				balances.GET("/historical", balanceHandler.GetHistoricalBalance) // GET /api/v1/balances/historical
				balances.GET("/at-time", balanceHandler.GetBalanceAtTime)        // GET /api/v1/balances/at-time
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

	// Health check endpoints
	health := r.Group("/health")
	{
		health.GET("", healthCheckHandler)
		health.GET("/ready", readinessCheckHandler)
		health.GET("/live", livenessCheckHandler)
		health.GET("/cache", cacheHealthCheckHandler)
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

func cacheHealthCheckHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":    "cache_healthy",
		"timestamp": time.Now().UTC(),
		"cache":     "redis",
		"message":   "Cache health check endpoint - implement cache service access",
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
