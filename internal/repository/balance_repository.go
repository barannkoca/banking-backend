package repository

import (
	"context"
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

// GetBalance retrieves the current balance for a given account ID
func (br *BalanceRepository) GetBalance(ctx context.Context, accountID uuid.UUID) (float64, error) {
	var balance models.Balance
	err := br.db.WithContext(ctx).Where("user_id = ?", accountID).First(&balance).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create a new balance record with zero amount
			newBalance := &models.Balance{
				UserID:        accountID,
				Amount:        0,
				LastUpdatedAt: time.Now(),
			}
			if err := br.db.WithContext(ctx).Create(newBalance).Error; err != nil {
				return 0, err
			}
			return 0, nil
		}
		return 0, err
	}
	return balance.Amount, nil
}

// UpdateBalance updates the balance for a given account ID
func (br *BalanceRepository) UpdateBalance(ctx context.Context, accountID uuid.UUID, amount float64) error {
	return br.db.WithContext(ctx).Model(&models.Balance{}).
		Where("user_id = ?", accountID).
		Updates(map[string]interface{}{
			"amount":          amount,
			"last_updated_at": time.Now(),
		}).Error
}

// SaveBalanceHistory saves a balance history record
func (br *BalanceRepository) SaveBalanceHistory(ctx context.Context, accountID uuid.UUID, amount float64, timestamp time.Time) error {
	history := &models.BalanceHistory{
		UserID:       accountID,
		NewAmount:    amount,
		ChangeAmount: 0, // This would be calculated based on previous amount
		ChangeType:   "BALANCE_UPDATE",
		CreatedAt:    timestamp,
	}
	return br.db.WithContext(ctx).Create(history).Error
}

// Additional helper methods

// GetBalanceModel gets the full balance model for an account
func (br *BalanceRepository) GetBalanceModel(ctx context.Context, accountID uuid.UUID) (*models.Balance, error) {
	var balance models.Balance
	err := br.db.WithContext(ctx).Where("user_id = ?", accountID).First(&balance).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create a new balance record with zero amount
			newBalance := &models.Balance{
				UserID:        accountID,
				Amount:        0,
				LastUpdatedAt: time.Now(),
			}
			if err := br.db.WithContext(ctx).Create(newBalance).Error; err != nil {
				return nil, err
			}
			return newBalance, nil
		}
		return nil, err
	}
	return &balance, nil
}

// GetBalanceHistory gets balance history for an account
func (br *BalanceRepository) GetBalanceHistory(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]models.BalanceHistory, error) {
	var history []models.BalanceHistory
	err := br.db.WithContext(ctx).
		Where("user_id = ?", accountID).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&history).Error
	return history, err
}

// CreateBalance creates a new balance record
func (br *BalanceRepository) CreateBalance(ctx context.Context, balance *models.Balance) error {
	return br.db.WithContext(ctx).Create(balance).Error
}

// DeleteBalance deletes a balance record
func (br *BalanceRepository) DeleteBalance(ctx context.Context, accountID uuid.UUID) error {
	return br.db.WithContext(ctx).Where("user_id = ?", accountID).Delete(&models.Balance{}).Error
}

// GetAllBalances gets all balance records
func (br *BalanceRepository) GetAllBalances(ctx context.Context, limit, offset int) ([]*models.Balance, error) {
	var balances []*models.Balance
	err := br.db.WithContext(ctx).
		Order("last_updated_at DESC").
		Limit(limit).Offset(offset).
		Find(&balances).Error
	return balances, err
}

// CountBalances counts total balance records
func (br *BalanceRepository) CountBalances(ctx context.Context) (int64, error) {
	var count int64
	err := br.db.WithContext(ctx).Model(&models.Balance{}).Count(&count).Error
	return count, err
}
