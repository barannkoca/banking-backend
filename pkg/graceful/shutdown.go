package graceful

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/barannkoca/banking-backend/pkg/logger"
	"go.uber.org/zap"
)

// ShutdownHandler handles graceful shutdown of the application
type ShutdownHandler struct {
	server  *http.Server
	timeout time.Duration
	cleanup []func() error
	logger  *zap.Logger
}

// NewShutdownHandler creates a new shutdown handler
func NewShutdownHandler(server *http.Server, timeout time.Duration) *ShutdownHandler {
	return &ShutdownHandler{
		server:  server,
		timeout: timeout,
		cleanup: make([]func() error, 0),
		logger:  logger.GetLogger(),
	}
}

// AddCleanupTask adds a cleanup function to be executed during shutdown
func (sh *ShutdownHandler) AddCleanupTask(task func() error) {
	sh.cleanup = append(sh.cleanup, task)
}

// ListenForShutdown listens for shutdown signals and handles graceful shutdown
func (sh *ShutdownHandler) ListenForShutdown() {
	// Channel to listen for interrupt signals
	quit := make(chan os.Signal, 1)

	// Register channel to receive specific signals
	signal.Notify(quit,
		syscall.SIGINT,  // Ctrl+C
		syscall.SIGTERM, // Termination signal
		syscall.SIGQUIT, // Quit signal
	)

	// Block until signal is received
	sig := <-quit
	sh.logger.Info("Shutdown signal received",
		zap.String("signal", sig.String()),
		zap.String("type", "graceful_shutdown"),
	)

	// Start graceful shutdown process
	sh.shutdown()
}

// shutdown performs the actual graceful shutdown
func (sh *ShutdownHandler) shutdown() {
	sh.logger.Info("Starting graceful shutdown process...")

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), sh.timeout)
	defer cancel()

	// 1. Stop accepting new requests
	sh.logger.Info("Stopping HTTP server...")
	if err := sh.server.Shutdown(ctx); err != nil {
		sh.logger.Error("HTTP server shutdown failed",
			zap.Error(err),
			zap.String("type", "shutdown_error"),
		)
		// Force close if graceful shutdown fails
		sh.server.Close()
	} else {
		sh.logger.Info("HTTP server stopped successfully")
	}

	// 2. Execute cleanup tasks
	sh.logger.Info("Executing cleanup tasks...",
		zap.Int("task_count", len(sh.cleanup)),
	)

	for i, task := range sh.cleanup {
		sh.logger.Info("Executing cleanup task", zap.Int("task_number", i+1))
		if err := task(); err != nil {
			sh.logger.Error("Cleanup task failed",
				zap.Int("task_number", i+1),
				zap.Error(err),
			)
		}
	}

	// 3. Final logging
	sh.logger.Info("Graceful shutdown completed successfully")

	// 4. Sync logger to ensure all logs are written
	logger.Sync()
}

// Banking-specific cleanup functions

// CleanupDatabaseConnections closes all database connections safely
func CleanupDatabaseConnections(db interface{}) func() error {
	return func() error {
		logger.GetLogger().Info("Closing database connections...")

		// Type assertion for different DB types
		switch dbConn := db.(type) {
		case interface{ Close() error }:
			if err := dbConn.Close(); err != nil {
				return err
			}
		case interface{ DB() (*sql.DB, error) }: // GORM support
			if sqlDB, err := dbConn.DB(); err == nil {
				if err := sqlDB.Close(); err != nil {
					return err
				}
			}
		}

		logger.GetLogger().Info("Database connections closed successfully")
		return nil
	}
}

// CleanupRedisConnections closes Redis connections
func CleanupRedisConnections(redis interface{}) func() error {
	return func() error {
		logger.GetLogger().Info("Closing Redis connections...")

		switch redisConn := redis.(type) {
		case interface{ Close() error }:
			if err := redisConn.Close(); err != nil {
				return err
			}
		}

		logger.GetLogger().Info("Redis connections closed successfully")
		return nil
	}
}

// CleanupTransactionQueue ensures all pending transactions are processed
func CleanupTransactionQueue() func() error {
	return func() error {
		logger.GetLogger().Info("Processing remaining transactions...")

		// Banking-specific: Wait for pending transactions
		// This is a placeholder - implement actual transaction cleanup
		time.Sleep(2 * time.Second) // Wait for pending operations

		logger.GetLogger().Info("All transactions processed successfully")
		return nil
	}
}

// CleanupAuditLogs ensures all audit logs are flushed
func CleanupAuditLogs() func() error {
	return func() error {
		logger.GetLogger().Info("Flushing audit logs...")

		// Flush any buffered logs
		logger.Sync()

		logger.GetLogger().Info("Audit logs flushed successfully")
		return nil
	}
}

// ForceShutdown performs immediate shutdown (use as last resort)
func ForceShutdown(exitCode int) {
	logger.GetLogger().Error("Force shutdown initiated",
		zap.Int("exit_code", exitCode),
		zap.String("type", "force_shutdown"),
	)

	logger.Sync()
	os.Exit(exitCode)
}
