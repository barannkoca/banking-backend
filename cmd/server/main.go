package main

import (
	"net/http"
	"time"

	"github.com/barannkoca/banking-backend/config"
	"github.com/barannkoca/banking-backend/internal/api"
	"github.com/barannkoca/banking-backend/internal/database"
	"github.com/barannkoca/banking-backend/internal/processing"
	"github.com/barannkoca/banking-backend/internal/repository"
	"github.com/barannkoca/banking-backend/internal/services"
	"github.com/barannkoca/banking-backend/pkg/graceful"
	"github.com/barannkoca/banking-backend/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}

	// Initialize logger
	if err := logger.InitLogger(cfg.App.Environment); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}

	log := logger.GetLogger()
	log.Info("Starting Banking Backend Server...",
		zap.String("environment", cfg.App.Environment),
		zap.String("version", "1.0.0"),
	)

	// Initialize Database
	if err := database.InitDatabase(); err != nil {
		log.Fatal("Failed to initialize database",
			zap.Error(err),
			zap.String("type", "database_error"),
		)
	}

	// Run database migrations
	if err := database.AutoMigrate(); err != nil {
		log.Fatal("Failed to run database migrations",
			zap.Error(err),
			zap.String("type", "migration_error"),
		)
	}

	// Seed database with initial data (development only)
	if err := database.SeedData(); err != nil {
		log.Warn("Failed to seed database",
			zap.Error(err),
			zap.String("type", "seed_error"),
		)
	}

	// Initialize repositories and services
	userRepo := repository.NewUserRepository(database.GetDB())
	transactionRepo := repository.NewTransactionRepository(database.GetDB())
	balanceRepo := repository.NewBalanceRepository(database.GetDB())

	// Initialize Redis cache service
	cacheService, err := services.NewRedisCacheService("localhost:6379", "", 0)
	if err != nil {
		log.Warn("Failed to initialize Redis cache, continuing without cache",
			zap.Error(err),
			zap.String("type", "cache_init_warning"),
		)
		cacheService = nil
	} else {
		log.Info("Redis cache service initialized successfully")
	}

	// Initialize audit service
	auditService := services.NewAuditService(log)

	userService := services.NewUserService(userRepo, auditService)
	transactionService := services.NewTransactionService(transactionRepo, balanceRepo, auditService, cacheService, log)
	balanceService := services.NewBalanceService(balanceRepo, auditService, cacheService)

	// Initialize worker pool for transaction processing
	workerPool := processing.NewWorkerPool(5, 100, log) // 5 workers, 100 job queue size

	// Initialize custom router with all middleware
	r := api.SetupRouter(userService, transactionService, balanceService.(*services.BalanceService), auditService, workerPool)

	// Create HTTP server
	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}

	// Initialize graceful shutdown handler
	shutdownHandler := graceful.NewShutdownHandler(server, 30*time.Second)

	// Add cleanup tasks for banking system
	shutdownHandler.AddCleanupTask(func() error {
		log.Info("Shutting down worker pool...")
		return workerPool.Shutdown(30 * time.Second)
	})
	shutdownHandler.AddCleanupTask(graceful.CleanupTransactionQueue())
	shutdownHandler.AddCleanupTask(graceful.CleanupAuditLogs())

	// Add database cleanup
	shutdownHandler.AddCleanupTask(graceful.CleanupDatabaseConnections(database.GetDB()))

	// Start server in a goroutine
	go func() {
		log.Info("Server starting",
			zap.String("address", server.Addr),
			zap.String("type", "server_start"),
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server failed to start",
				zap.Error(err),
				zap.String("type", "server_error"),
			)
		}
	}()

	log.Info("Banking Backend Server started successfully",
		zap.String("address", server.Addr),
		zap.String("status", "running"),
	)

	// Listen for shutdown signals (this blocks)
	shutdownHandler.ListenForShutdown()

	log.Info("Banking Backend Server stopped")
}
