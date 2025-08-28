package middleware

import (
	"bytes"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/barannkoca/banking-backend/pkg/logger"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// PerformanceMetrics holds performance data for a request
type PerformanceMetrics struct {
	Method          string
	Path            string
	Status          int
	Duration        time.Duration
	RequestSize     int64
	ResponseSize    int64
	UserAgent       string
	ClientIP        string
	RequestID       string
	StartTime       time.Time
	EndTime         time.Time
	DatabaseQueries int
	CacheHits       int
	CacheMisses     int
	ErrorCount      int
}

// PerformanceOptions defines configuration options for performance monitoring
type PerformanceOptions struct {
	Enabled      bool
	ExcludePaths []string
	IncludePaths []string
	Thresholds   *PerformanceThresholds
}

// PerformanceMonitorMiddleware provides comprehensive performance monitoring
func PerformanceMonitorMiddleware() gin.HandlerFunc {
	return PerformanceMonitorWithOptions(PerformanceOptions{
		Enabled: true,
		ExcludePaths: []string{
			"/health",
			"/health/ready",
			"/health/live",
		},
		IncludePaths: []string{}, // Boş ise tüm path'ler dahil
	})
}

// PerformanceMonitorWithOptions creates middleware with custom options
func PerformanceMonitorWithOptions(options PerformanceOptions) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if monitoring is enabled
		if !options.Enabled {
			c.Next()
			return
		}

		// Check if path should be excluded
		if shouldExcludePath(c.Request.URL.Path, options.ExcludePaths) {
			c.Next()
			return
		}

		// Check if path should be included (if IncludePaths is not empty)
		if len(options.IncludePaths) > 0 && !shouldIncludePath(c.Request.URL.Path, options.IncludePaths) {
			c.Next()
			return
		}

		start := time.Now()

		// Get request ID from context (set by tracking middleware)
		requestID := c.GetString("request_id")
		if requestID == "" {
			requestID = "unknown"
		}

		// Create metrics struct
		metrics := &PerformanceMetrics{
			Method:    c.Request.Method,
			Path:      c.Request.URL.Path,
			UserAgent: c.Request.UserAgent(),
			ClientIP:  c.ClientIP(),
			RequestID: requestID,
			StartTime: start,
		}

		// Calculate request size
		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			metrics.RequestSize = int64(len(bodyBytes))
			// Restore body for other middleware/handlers
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Create a custom response writer to capture response size
		responseWriter := &responseWriter{
			ResponseWriter: c.Writer,
			metrics:        metrics,
		}
		c.Writer = responseWriter

		// Add metrics to context for other middleware/handlers to update
		c.Set("performance_metrics", metrics)

		// Process request
		c.Next()

		// Calculate final metrics
		metrics.EndTime = time.Now()
		metrics.Duration = metrics.EndTime.Sub(metrics.StartTime)
		metrics.Status = c.Writer.Status()

		// Log performance metrics
		logPerformanceMetrics(metrics)

		// Add performance headers to response
		addPerformanceHeaders(c, metrics)
	}
}

// shouldExcludePath checks if the path should be excluded from monitoring
func shouldExcludePath(path string, excludePaths []string) bool {
	for _, excludePath := range excludePaths {
		if path == excludePath || (len(excludePath) > 1 && excludePath[len(excludePath)-1] == '*' &&
			strings.HasPrefix(path, excludePath[:len(excludePath)-1])) {
			return true
		}
	}
	return false
}

// shouldIncludePath checks if the path should be included in monitoring
func shouldIncludePath(path string, includePaths []string) bool {
	for _, includePath := range includePaths {
		if path == includePath || (len(includePath) > 1 && includePath[len(includePath)-1] == '*' &&
			strings.HasPrefix(path, includePath[:len(includePath)-1])) {
			return true
		}
	}
	return false
}

// responseWriter wraps gin.ResponseWriter to capture response size
type responseWriter struct {
	gin.ResponseWriter
	metrics *PerformanceMetrics
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.metrics.ResponseSize += int64(len(b))
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteString(s string) (int, error) {
	w.metrics.ResponseSize += int64(len(s))
	return w.ResponseWriter.WriteString(s)
}

// logPerformanceMetrics logs performance data with appropriate log levels
func logPerformanceMetrics(metrics *PerformanceMetrics) {
	log := logger.GetLogger()

	// Determine log level based on performance
	var logLevel zapcore.Level
	var logMessage string

	switch {
	case metrics.Duration > 5*time.Second:
		logLevel = zapcore.ErrorLevel
		logMessage = "Very slow request detected"
	case metrics.Duration > 2*time.Second:
		logLevel = zapcore.WarnLevel
		logMessage = "Slow request detected"
	case metrics.Duration > 1*time.Second:
		logLevel = zapcore.InfoLevel
		logMessage = "Moderate request time"
	default:
		logLevel = zapcore.DebugLevel
		logMessage = "Fast request"
	}

	// Create log fields
	fields := []zap.Field{
		zap.String("method", metrics.Method),
		zap.String("path", metrics.Path),
		zap.Int("status", metrics.Status),
		zap.Duration("duration", metrics.Duration),
		zap.Int64("request_size", metrics.RequestSize),
		zap.Int64("response_size", metrics.ResponseSize),
		zap.String("user_agent", metrics.UserAgent),
		zap.String("client_ip", metrics.ClientIP),
		zap.String("request_id", metrics.RequestID),
		zap.Time("start_time", metrics.StartTime),
		zap.Time("end_time", metrics.EndTime),
		zap.Int("database_queries", metrics.DatabaseQueries),
		zap.Int("cache_hits", metrics.CacheHits),
		zap.Int("cache_misses", metrics.CacheMisses),
		zap.Int("error_count", metrics.ErrorCount),
		zap.String("type", "performance_metrics"),
	}

	// Log with appropriate level
	switch logLevel {
	case zapcore.ErrorLevel:
		log.Error(logMessage, fields...)
	case zapcore.WarnLevel:
		log.Warn(logMessage, fields...)
	case zapcore.InfoLevel:
		log.Info(logMessage, fields...)
	default:
		log.Debug(logMessage, fields...)
	}

	// Log to performance-specific logger if configured
	logPerformanceToFile(metrics)
}

// addPerformanceHeaders adds performance-related headers to response
func addPerformanceHeaders(c *gin.Context, metrics *PerformanceMetrics) {
	c.Header("X-Response-Time", metrics.Duration.String())
	c.Header("X-Request-ID", metrics.RequestID)
	c.Header("X-Request-Size", strconv.FormatInt(metrics.RequestSize, 10))
	c.Header("X-Response-Size", strconv.FormatInt(metrics.ResponseSize, 10))
	c.Header("X-Database-Queries", strconv.Itoa(metrics.DatabaseQueries))
	c.Header("X-Cache-Hits", strconv.Itoa(metrics.CacheHits))
	c.Header("X-Cache-Misses", strconv.Itoa(metrics.CacheMisses))
	c.Header("X-Error-Count", strconv.Itoa(metrics.ErrorCount))
}

// logPerformanceToFile logs performance data to a separate file (if configured)
func logPerformanceToFile(metrics *PerformanceMetrics) {
	// This could be implemented to log to a performance-specific file
	// or send to external monitoring systems like Prometheus, DataDog, etc.

	// Example: Log to performance.log
	// performanceLogger := logger.GetPerformanceLogger()
	// if performanceLogger != nil {
	//     performanceLogger.Info("performance_metric", zap.Any("metrics", metrics))
	// }
}

// Performance monitoring helper functions for other middleware/handlers

// IncrementDatabaseQueries increments the database query counter
func IncrementDatabaseQueries(c *gin.Context) {
	if metrics, exists := c.Get("performance_metrics"); exists {
		if perfMetrics, ok := metrics.(*PerformanceMetrics); ok {
			perfMetrics.DatabaseQueries++
		}
	}
}

// IncrementCacheHits increments the cache hit counter
func IncrementCacheHits(c *gin.Context) {
	if metrics, exists := c.Get("performance_metrics"); exists {
		if perfMetrics, ok := metrics.(*PerformanceMetrics); ok {
			perfMetrics.CacheHits++
		}
	}
}

// IncrementCacheMisses increments the cache miss counter
func IncrementCacheMisses(c *gin.Context) {
	if metrics, exists := c.Get("performance_metrics"); exists {
		if perfMetrics, ok := metrics.(*PerformanceMetrics); ok {
			perfMetrics.CacheMisses++
		}
	}
}

// IncrementErrorCount increments the error counter
func IncrementErrorCount(c *gin.Context) {
	if metrics, exists := c.Get("performance_metrics"); exists {
		if perfMetrics, ok := metrics.(*PerformanceMetrics); ok {
			perfMetrics.ErrorCount++
		}
	}
}

// GetPerformanceMetrics retrieves performance metrics from context
func GetPerformanceMetrics(c *gin.Context) *PerformanceMetrics {
	if metrics, exists := c.Get("performance_metrics"); exists {
		if perfMetrics, ok := metrics.(*PerformanceMetrics); ok {
			return perfMetrics
		}
	}
	return nil
}

// PerformanceThresholds defines performance thresholds for alerts
type PerformanceThresholds struct {
	SlowRequestThreshold     time.Duration
	VerySlowRequestThreshold time.Duration
	LargeRequestSize         int64
	LargeResponseSize        int64
	MaxDatabaseQueries       int
	MaxErrorCount            int
}

// DefaultPerformanceThresholds returns default performance thresholds
func DefaultPerformanceThresholds() *PerformanceThresholds {
	return &PerformanceThresholds{
		SlowRequestThreshold:     1 * time.Second,
		VerySlowRequestThreshold: 5 * time.Second,
		LargeRequestSize:         1024 * 1024, // 1MB
		LargeResponseSize:        1024 * 1024, // 1MB
		MaxDatabaseQueries:       50,
		MaxErrorCount:            5,
	}
}

// PerformanceMonitorWithThresholds creates middleware with custom thresholds
func PerformanceMonitorWithThresholds(thresholds *PerformanceThresholds) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set thresholds in context
		c.Set("performance_thresholds", thresholds)

		// Call the main performance monitor
		PerformanceMonitorMiddleware()(c)
	}
}
