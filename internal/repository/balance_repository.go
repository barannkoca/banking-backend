package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/barannkoca/banking-backend/internal/interfaces"
	"github.com/barannkoca/banking-backend/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BalanceRepository struct {
	db *gorm.DB
}

func NewBalanceRepository(db *gorm.DB) interfaces.BalanceRepository {
	return &BalanceRepository{
		db: db,
	}
}

// GetBalance gets balance by account ID
func (br *BalanceRepository) GetBalance(ctx context.Context, accountID uuid.UUID) (float64, error) {
	var balance models.Balance
	err := br.db.WithContext(ctx).Where("user_id = ?", accountID).First(&balance).Error
	if err != nil {
		return 0, err
	}
	return balance.Amount, nil
}

// UpdateBalance updates balance amount
func (br *BalanceRepository) UpdateBalance(ctx context.Context, accountID uuid.UUID, amount float64) error {
	return br.db.WithContext(ctx).Model(&models.Balance{}).
		Where("user_id = ?", accountID).
		Update("amount", amount).Error
}

// SaveBalanceHistory saves balance history
func (br *BalanceRepository) SaveBalanceHistory(ctx context.Context, accountID uuid.UUID, amount float64, timestamp time.Time) error {
	// This would typically save to a balance_history table
	// For now, we'll just log it or save to audit log
	return nil
}

// Additional helper methods (not in interface but useful for internal use)

// Create creates a new balance
func (br *BalanceRepository) Create(ctx context.Context, balance *models.Balance) error {
	return br.db.WithContext(ctx).Create(balance).Error
}

// GetByUserID gets balance by user ID
func (br *BalanceRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*models.Balance, error) {
	var balance models.Balance
	err := br.db.WithContext(ctx).Where("user_id = ?", userID).First(&balance).Error
	if err != nil {
		return nil, err
	}
	return &balance, nil
}

// GetAll gets all balances with pagination
func (br *BalanceRepository) GetAll(ctx context.Context, limit, offset int) ([]*models.Balance, error) {
	var balances []*models.Balance
	err := br.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&balances).Error
	return balances, err
}

// Update updates a balance
func (br *BalanceRepository) Update(ctx context.Context, balance *models.Balance) error {
	return br.db.WithContext(ctx).Save(balance).Error
}

// UpdateAmount updates balance amount
func (br *BalanceRepository) UpdateAmount(ctx context.Context, userID uuid.UUID, amount float64) error {
	return br.db.WithContext(ctx).Model(&models.Balance{}).
		Where("user_id = ?", userID).
		Update("amount", amount).Error
}

// Delete deletes a balance
func (br *BalanceRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	return br.db.WithContext(ctx).Where("user_id = ?", userID).Delete(&models.Balance{}).Error
}

// TransferBetweenUsers transfers money between users (ACID protected)
func (br *BalanceRepository) TransferBetweenUsers(ctx context.Context, fromUserID, toUserID uuid.UUID, amount float64) error {
	return br.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Check sender balance
		var fromBalance models.Balance
		if err := tx.Where("user_id = ?", fromUserID).First(&fromBalance).Error; err != nil {
			return fmt.Errorf("sender balance not found: %w", err)
		}
		if fromBalance.Amount < amount {
			return fmt.Errorf("insufficient balance")
		}

		// 2. Subtract from sender
		if err := tx.Model(&models.Balance{}).
			Where("user_id = ?", fromUserID).
			Update("amount", gorm.Expr("amount - ?", amount)).Error; err != nil {
			return fmt.Errorf("failed to subtract from sender: %w", err)
		}

		// 3. Add to receiver
		if err := tx.Model(&models.Balance{}).
			Where("user_id = ?", toUserID).
			Update("amount", gorm.Expr("amount + ?", amount)).Error; err != nil {
			return fmt.Errorf("failed to add to receiver: %w", err)
		}

		return nil
	})
}

// AddToBalance adds amount to user balance (ACID protected)
func (br *BalanceRepository) AddToBalance(ctx context.Context, userID uuid.UUID, amount float64) error {
	return br.db.WithContext(ctx).Model(&models.Balance{}).
		Where("user_id = ?", userID).
		Update("amount", gorm.Expr("amount + ?", amount)).Error
}

// SubtractFromBalance subtracts amount from user balance (ACID protected)
func (br *BalanceRepository) SubtractFromBalance(ctx context.Context, userID uuid.UUID, amount float64) error {
	return br.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Check sufficient balance
		var balance models.Balance
		if err := tx.Where("user_id = ?", userID).First(&balance).Error; err != nil {
			return fmt.Errorf("balance not found: %w", err)
		}
		if balance.Amount < amount {
			return fmt.Errorf("insufficient balance")
		}

		// Subtract amount
		return tx.Model(&models.Balance{}).
			Where("user_id = ?", userID).
			Update("amount", gorm.Expr("amount - ?", amount)).Error
	})
}

// GetTotalBalance gets total system balance
func (br *BalanceRepository) GetTotalBalance(ctx context.Context) (float64, error) {
	var total float64
	err := br.db.WithContext(ctx).Model(&models.Balance{}).Select("COALESCE(SUM(amount), 0)").Scan(&total).Error
	return total, err
}

// Count counts total balances
func (br *BalanceRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := br.db.WithContext(ctx).Model(&models.Balance{}).Count(&count).Error
	return count, err
}
