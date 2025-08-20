package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

// InitLogger initializes the global logger instance
func InitLogger(env string) error {
	var config zap.Config

	if env == "production" {
		// Production: JSON format, Info level
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	} else {
		// Development: Console format, Debug level
		config = zap.NewDevelopmentConfig()
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Customize encoder config for banking system
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.MessageKey = "message"
	config.EncoderConfig.LevelKey = "level"
	config.EncoderConfig.CallerKey = "caller"

	// Add file output for production
	if env == "production" {
		config.OutputPaths = []string{"stdout", "logs/banking.log"}
		config.ErrorOutputPaths = []string{"stderr", "logs/banking-error.log"}
	}

	var err error
	Logger, err = config.Build(zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	if err != nil {
		return err
	}

	// Create logs directory if it doesn't exist
	if env == "production" {
		if err := os.MkdirAll("logs", 0755); err != nil {
			return err
		}
	}

	return nil
}

// GetLogger returns the global logger instance
func GetLogger() *zap.Logger {
	if Logger == nil {
		// Fallback to a basic logger if not initialized
		Logger, _ = zap.NewDevelopment()
	}
	return Logger
}

// Sync flushes buffered log entries
func Sync() {
	if Logger != nil {
		Logger.Sync()
	}
}

// Banking specific logging functions

// LogTransaction logs banking transaction details
func LogTransaction(transactionID, fromAccount, toAccount, userID string, amount float64, status string) {
	Logger.Info("Banking Transaction",
		zap.String("transaction_id", transactionID),
		zap.String("from_account", fromAccount),
		zap.String("to_account", toAccount),
		zap.String("user_id", userID),
		zap.Float64("amount", amount),
		zap.String("status", status),
		zap.String("type", "transaction"),
	)
}

// LogAuth logs authentication events
func LogAuth(userID, action, ip string, success bool) {
	level := zap.InfoLevel
	if !success {
		level = zap.WarnLevel
	}

	Logger.Log(level, "Authentication Event",
		zap.String("user_id", userID),
		zap.String("action", action),
		zap.String("ip_address", ip),
		zap.Bool("success", success),
		zap.String("type", "auth"),
	)
}

// LogSecurity logs security-related events
func LogSecurity(event, userID, details string, severity string) {
	var level zapcore.Level
	switch severity {
	case "critical":
		level = zap.ErrorLevel
	case "high":
		level = zap.WarnLevel
	default:
		level = zap.InfoLevel
	}

	Logger.Log(level, "Security Event",
		zap.String("event", event),
		zap.String("user_id", userID),
		zap.String("details", details),
		zap.String("severity", severity),
		zap.String("type", "security"),
	)
}

// LogAPIRequest logs HTTP API requests
func LogAPIRequest(method, path, userID, ip string, statusCode int, duration int64) {
	Logger.Info("API Request",
		zap.String("method", method),
		zap.String("path", path),
		zap.String("user_id", userID),
		zap.String("ip_address", ip),
		zap.Int("status_code", statusCode),
		zap.Int64("duration_ms", duration),
		zap.String("type", "api"),
	)
}
