package interfaces

import (
	"context"

	"github.com/barannkoca/banking-backend/internal/models"
	"github.com/google/uuid"
)

// AuthService defines the interface for JWT token operations
type AuthService interface {
	// Token management
	ValidateToken(ctx context.Context, token string) (*models.User, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, error)
	GenerateToken(ctx context.Context, user *models.User) (string, error)
	Logout(ctx context.Context, userID uuid.UUID) error

	// Account management
	DeactivateAccount(ctx context.Context, userID uuid.UUID) error
	ActivateAccount(ctx context.Context, userID uuid.UUID) error
}

// UserService defines the interface for user management and authentication operations
type UserService interface {
	// User management
	CreateUser(ctx context.Context, req *models.UserCreateRequest) (*models.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error
	DeleteUser(ctx context.Context, id uuid.UUID) error

	// User listing and searching
	GetAllUsers(ctx context.Context, limit, offset int) ([]*models.User, error)
	SearchUsers(ctx context.Context, query string, limit, offset int) ([]*models.User, error)

	// Role management
	UpdateUserRole(ctx context.Context, userID uuid.UUID, role models.UserRole) error
	GetUsersByRole(ctx context.Context, role models.UserRole, limit, offset int) ([]*models.User, error)

	// Validation
	ValidateUserData(ctx context.Context, user *models.User) error
	CheckEmailAvailability(ctx context.Context, email string) (bool, error)
	CheckUsernameAvailability(ctx context.Context, username string) (bool, error)

	// Authentication
	AuthenticateUser(ctx context.Context, usernameOrEmail, password string) (*models.User, error)
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error
}

// TransactionService defines the interface for transaction operations
type TransactionService interface {
	// Core transaction operations
	Credit(ctx context.Context, accountID uuid.UUID, amount float64) error
	Debit(ctx context.Context, accountID uuid.UUID, amount float64) error
	Transfer(ctx context.Context, fromAccountID, toAccountID uuid.UUID, amount float64) error
}

// BalanceService defines the interface for balance management operations
type BalanceService interface {
	GetBalance(ctx context.Context, accountID uuid.UUID) (float64, error)
	UpdateBalance(ctx context.Context, accountID uuid.UUID, amount float64) error

	// Thread-safe update (örn. RWMutex veya DB lock ile)
	SafeUpdateBalance(ctx context.Context, accountID uuid.UUID, amount float64) error

	// History takibi (audit log dışında daha "balance-centric" tracking)
	GetBalanceHistory(ctx context.Context, accountID uuid.UUID) ([]models.BalanceHistory, error)

	// Optimizasyon (cache, pre-computation vb.)
	CalculateAvailableBalance(ctx context.Context, accountID uuid.UUID) (float64, error)
}

// AuditService defines the interface for audit log operations
type AuditService interface {
	// Audit logging
	LogUserActivity(ctx context.Context, userID uuid.UUID, action, entityType, entityID, details string) error
	LogTransactionActivity(ctx context.Context, transaction *models.Transaction, action, details string) error
	LogSystemActivity(ctx context.Context, action, details string) error

	// Audit queries
	GetAuditLogsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.AuditLog, error)
	GetAuditLogsByEntityID(ctx context.Context, entityID string, limit, offset int) ([]*models.AuditLog, error)
	GetAuditLogsByAction(ctx context.Context, action string, limit, offset int) ([]*models.AuditLog, error)
	GetAllAuditLogs(ctx context.Context, limit, offset int) ([]*models.AuditLog, error)

	// Audit management
	DeleteOldAuditLogs(ctx context.Context, olderThanDays int) error
	GetAuditStatistics(ctx context.Context) (map[string]int64, error)
}

// NotificationService defines the interface for notification operations
type NotificationService interface {
	// Email notifications
	SendWelcomeEmail(ctx context.Context, user *models.User) error
	SendTransactionNotification(ctx context.Context, user *models.User, transaction *models.Transaction) error
	SendPasswordResetEmail(ctx context.Context, user *models.User, resetToken string) error
	SendBalanceAlert(ctx context.Context, user *models.User, balance *models.Balance) error

	// SMS notifications (if implemented)
	SendSMSNotification(ctx context.Context, phoneNumber, message string) error

	// Push notifications (if implemented)
	SendPushNotification(ctx context.Context, userID uuid.UUID, title, message string) error
}

// ReportService defines the interface for reporting operations
type ReportService interface {
	// Financial reports
	GenerateTransactionReport(ctx context.Context, startDate, endDate string) (interface{}, error)
	GenerateBalanceReport(ctx context.Context) (interface{}, error)
	GenerateUserActivityReport(ctx context.Context, userID uuid.UUID, startDate, endDate string) (interface{}, error)

	// System reports
	GenerateSystemHealthReport(ctx context.Context) (interface{}, error)
	GenerateAuditReport(ctx context.Context, startDate, endDate string) (interface{}, error)

	// Export functionality
	ExportTransactionsToCSV(ctx context.Context, startDate, endDate string) ([]byte, error)
	ExportUsersToCSV(ctx context.Context) ([]byte, error)
}

// CacheService defines the interface for caching operations
type CacheService interface {
	// Basic cache operations
	Set(ctx context.Context, key string, value interface{}, ttl int) error
	Get(ctx context.Context, key string) (interface{}, error)
	Delete(ctx context.Context, key string) error

	// Cache management
	Clear(ctx context.Context) error
	Exists(ctx context.Context, key string) (bool, error)

	// User-specific caching
	CacheUser(ctx context.Context, user *models.User, ttl int) error
	GetCachedUser(ctx context.Context, userID uuid.UUID) (*models.User, error)
	InvalidateUserCache(ctx context.Context, userID uuid.UUID) error

	// Transaction-specific caching
	CacheTransaction(ctx context.Context, transaction *models.Transaction, ttl int) error
	GetCachedTransaction(ctx context.Context, transactionID uuid.UUID) (*models.Transaction, error)
	InvalidateTransactionCache(ctx context.Context, transactionID uuid.UUID) error
}

// ValidationService defines the interface for data validation operations
type ValidationService interface {
	// User validation
	ValidateUser(ctx context.Context, user *models.User) error
	ValidateUserCreation(ctx context.Context, req *models.UserCreateRequest) error

	// Transaction validation
	ValidateTransaction(ctx context.Context, transaction *models.Transaction) error
	ValidateTransferRequest(ctx context.Context, req *models.TransferRequest) error
	ValidateDepositRequest(ctx context.Context, req *models.DepositRequest) error
	ValidateWithdrawRequest(ctx context.Context, req *models.WithdrawRequest) error

	// Balance validation
	ValidateBalance(ctx context.Context, balance *models.Balance) error

	// Security validation
	ValidatePassword(password string) error
	ValidateEmail(email string) error
	ValidateAmount(amount float64) error
}
