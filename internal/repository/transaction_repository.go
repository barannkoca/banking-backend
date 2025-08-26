package repository

import (
	"context"
	"time"

	"github.com/barannkoca/banking-backend/internal/interfaces"
	"github.com/barannkoca/banking-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TransactionRepository struct {
	db *gorm.DB
}

func NewTransactionRepository(db *gorm.DB) interfaces.TransactionRepository {
	return &TransactionRepository{
		db: db,
	}
}

// Save saves a transaction (create or update)
func (tr *TransactionRepository) Save(ctx context.Context, transaction *models.Transaction) error {
	return tr.db.WithContext(ctx).Save(transaction).Error
}

// FindByID finds transaction by ID
func (tr *TransactionRepository) FindByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error) {
	var transaction models.Transaction
	err := tr.db.WithContext(ctx).Where("id = ?", id).First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

// FindByAccount finds transactions by account ID
func (tr *TransactionRepository) FindByAccount(ctx context.Context, accountID uuid.UUID) ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	err := tr.db.WithContext(ctx).
		Where("from_user_id = ? OR to_user_id = ?", accountID, accountID).
		Order("created_at DESC").
		Find(&transactions).Error
	return transactions, err
}

// UpdateStatus updates transaction status
func (tr *TransactionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status models.TransactionStatus) error {
	return tr.db.WithContext(ctx).Model(&models.Transaction{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// Additional helper methods (not in interface but useful for internal use)

// Create creates a new transaction
func (tr *TransactionRepository) Create(ctx context.Context, transaction *models.Transaction) error {
	return tr.db.WithContext(ctx).Create(transaction).Error
}

// GetByID gets transaction by ID
func (tr *TransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error) {
	return tr.FindByID(ctx, id)
}

// GetByUserID gets transactions by user ID
func (tr *TransactionRepository) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	err := tr.db.WithContext(ctx).
		Where("from_user_id = ? OR to_user_id = ?", userID, userID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&transactions).Error
	return transactions, err
}

// GetByStatus gets transactions by status
func (tr *TransactionRepository) GetByStatus(ctx context.Context, status models.TransactionStatus, limit, offset int) ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	err := tr.db.WithContext(ctx).
		Where("status = ?", status).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&transactions).Error
	return transactions, err
}

// GetByType gets transactions by type
func (tr *TransactionRepository) GetByType(ctx context.Context, transactionType models.TransactionType, limit, offset int) ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	err := tr.db.WithContext(ctx).
		Where("type = ?", transactionType).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&transactions).Error
	return transactions, err
}

// GetByReference gets transaction by reference
func (tr *TransactionRepository) GetByReference(ctx context.Context, reference string) (*models.Transaction, error) {
	var transaction models.Transaction
	err := tr.db.WithContext(ctx).Where("reference = ?", reference).First(&transaction).Error
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

// GetAll gets all transactions with pagination
func (tr *TransactionRepository) GetAll(ctx context.Context, limit, offset int) ([]*models.Transaction, error) {
	var transactions []*models.Transaction
	err := tr.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&transactions).Error
	return transactions, err
}

// Update updates a transaction
func (tr *TransactionRepository) Update(ctx context.Context, transaction *models.Transaction) error {
	return tr.db.WithContext(ctx).Save(transaction).Error
}

// Delete deletes a transaction
func (tr *TransactionRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return tr.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Transaction{}).Error
}

// GetPendingTransactions gets pending transactions
func (tr *TransactionRepository) GetPendingTransactions(ctx context.Context, limit, offset int) ([]*models.Transaction, error) {
	return tr.GetByStatus(ctx, models.TransactionStatusPending, limit, offset)
}

// GetTransactionHistory gets transaction history for a user
func (tr *TransactionRepository) GetTransactionHistory(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Transaction, error) {
	return tr.GetByUserID(ctx, userID, limit, offset)
}

// GetRollbackableTransactions gets transactions that can be rolled back
func (tr *TransactionRepository) GetRollbackableTransactions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.Transaction, error) {
	// Get completed transactions within rollback window (e.g., 24 hours)
	rollbackWindow := time.Now().Add(-24 * time.Hour)

	var transactions []*models.Transaction
	err := tr.db.WithContext(ctx).
		Where("(from_user_id = ? OR to_user_id = ?) AND status = ? AND created_at > ?",
			userID, userID, models.TransactionStatusCompleted, rollbackWindow).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&transactions).Error
	return transactions, err
}

// Count counts total transactions
func (tr *TransactionRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := tr.db.WithContext(ctx).Model(&models.Transaction{}).Count(&count).Error
	return count, err
}

// CountByStatus counts transactions by status
func (tr *TransactionRepository) CountByStatus(ctx context.Context, status models.TransactionStatus) (int64, error) {
	var count int64
	err := tr.db.WithContext(ctx).Model(&models.Transaction{}).
		Where("status = ?", status).Count(&count).Error
	return count, err
}
