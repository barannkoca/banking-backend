package interfaces

import (
	"context"
	"time"

	"github.com/barannkoca/banking-backend/internal/models"
	"github.com/google/uuid"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// Create operations
	Create(ctx context.Context, user *models.User) error

	// Read operations
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByUsername(ctx context.Context, username string) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetAll(ctx context.Context, limit, offset int) ([]*models.User, error)

	// Update operations
	Update(ctx context.Context, user *models.User) error
	UpdateRole(ctx context.Context, userID uuid.UUID, role models.UserRole) error

	// Delete operations
	Delete(ctx context.Context, id uuid.UUID) error

	// Validation operations
	EmailExists(ctx context.Context, email string) (bool, error)
	UsernameExists(ctx context.Context, username string) (bool, error)

	// Count operations
	Count(ctx context.Context) (int64, error)
}

// TransactionRepository defines the interface for transaction data operations
type TransactionRepository interface {
	Save(ctx context.Context, tx *models.Transaction) error
	FindByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error)
	FindByAccount(ctx context.Context, accountID uuid.UUID) ([]*models.Transaction, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status models.TransactionStatus) error
}

// BalanceRepository defines the interface for balance data operations
type BalanceRepository interface {
	GetBalance(ctx context.Context, accountID uuid.UUID) (float64, error)
	UpdateBalance(ctx context.Context, accountID uuid.UUID, amount float64) error
	SaveBalanceHistory(ctx context.Context, accountID uuid.UUID, amount float64, timestamp time.Time) error
}

// AuditLogRepository defines the interface for audit log data operations
type AuditLogRepository interface {
	// Create operations
	Create(ctx context.Context, auditLog *models.AuditLog) error

	// Read operations
	GetByID(ctx context.Context, id uuid.UUID) (*models.AuditLog, error)
	GetByEntityID(ctx context.Context, entityID string, limit, offset int) ([]*models.AuditLog, error)
	GetByEntityType(ctx context.Context, entityType string, limit, offset int) ([]*models.AuditLog, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.AuditLog, error)
	GetByAction(ctx context.Context, action string, limit, offset int) ([]*models.AuditLog, error)
	GetAll(ctx context.Context, limit, offset int) ([]*models.AuditLog, error)

	// Delete operations
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteOldEntries(ctx context.Context, olderThan int) error // Delete entries older than X days

	// Count operations
	Count(ctx context.Context) (int64, error)
	CountByAction(ctx context.Context, action string) (int64, error)
}
