package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/barannkoca/banking-backend/internal/interfaces"
	"github.com/barannkoca/banking-backend/internal/models"
	"github.com/google/uuid"
)

// BalanceService implements the BalanceService interface
type BalanceService struct {
	balanceRepo  interfaces.BalanceRepository
	auditService interfaces.AuditService
	cache        interfaces.CacheService

	// Thread-safe balance locks for SafeUpdateBalance
	balanceLocks map[uuid.UUID]*sync.RWMutex
	locksMutex   sync.RWMutex
}

// NewBalanceService creates a new BalanceService instance
func NewBalanceService(
	balanceRepo interfaces.BalanceRepository,
	auditService interfaces.AuditService,
	cache interfaces.CacheService,
) interfaces.BalanceService {
	return &BalanceService{
		balanceRepo:  balanceRepo,
		auditService: auditService,
		cache:        cache,
		balanceLocks: make(map[uuid.UUID]*sync.RWMutex),
	}
}

// GetBalance retrieves the current balance for a given account ID
func (bs *BalanceService) GetBalance(ctx context.Context, accountID uuid.UUID) (float64, error) {
	// Try to get from cache first
	if bs.cache != nil {
		if cachedBalance, err := bs.cache.Get(ctx, fmt.Sprintf("balance:%s", accountID)); err == nil {
			if balance, ok := cachedBalance.(float64); ok {
				return balance, nil
			}
		}
	}

	// Get from repository
	balance, err := bs.balanceRepo.GetBalance(ctx, accountID)
	if err != nil {
		return 0, fmt.Errorf("bakiye alınamadı: %w", err)
	}

	// Cache the result
	if bs.cache != nil {
		bs.cache.Set(ctx, fmt.Sprintf("balance:%s", accountID), balance, 300) // 5 minutes TTL
	}

	return balance, nil
}

// UpdateBalance updates the balance for a given account ID
func (bs *BalanceService) UpdateBalance(ctx context.Context, accountID uuid.UUID, amount float64) error {
	// Validate amount
	if amount < 0 {
		return fmt.Errorf("bakiye negatif olamaz")
	}

	// Update in repository
	err := bs.balanceRepo.UpdateBalance(ctx, accountID, amount)
	if err != nil {
		return fmt.Errorf("bakiye güncellenemedi: %w", err)
	}

	// Invalidate cache
	if bs.cache != nil {
		bs.cache.Delete(ctx, fmt.Sprintf("balance:%s", accountID))
	}

	// Log audit
	if bs.auditService != nil {
		bs.auditService.LogUserActivity(ctx, accountID, "BALANCE_UPDATE", "balance", accountID.String(),
			fmt.Sprintf("Bakiye %f olarak güncellendi", amount))
	}

	return nil
}

// SafeUpdateBalance performs a thread-safe balance update using RWMutex
func (bs *BalanceService) SafeUpdateBalance(ctx context.Context, accountID uuid.UUID, amount float64) error {
	// Get or create lock for this account
	lock := bs.getOrCreateLock(accountID)

	// Acquire write lock
	lock.Lock()
	defer lock.Unlock()

	// Get current balance
	currentBalance, err := bs.balanceRepo.GetBalance(ctx, accountID)
	if err != nil {
		return fmt.Errorf("mevcut bakiye alınamadı: %w", err)
	}

	// Calculate new balance
	newBalance := currentBalance + amount
	if newBalance < 0 {
		return fmt.Errorf("yetersiz bakiye: mevcut %f, çıkarılacak %f", currentBalance, -amount)
	}

	// Update balance
	err = bs.balanceRepo.UpdateBalance(ctx, accountID, newBalance)
	if err != nil {
		return fmt.Errorf("bakiye güncellenemedi: %w", err)
	}

	// Save balance history
	if err := bs.balanceRepo.SaveBalanceHistory(ctx, accountID, newBalance, time.Now()); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Bakiye geçmişi kaydedilemedi: %v\n", err)
	}

	// Invalidate cache
	if bs.cache != nil {
		bs.cache.Delete(ctx, fmt.Sprintf("balance:%s", accountID))
	}

	// Log audit
	if bs.auditService != nil {
		action := "BALANCE_CREDIT"
		if amount < 0 {
			action = "BALANCE_DEBIT"
		}
		bs.auditService.LogUserActivity(ctx, accountID, action, "balance", accountID.String(),
			fmt.Sprintf("Bakiye %f değişti (mevcut: %f, yeni: %f)", amount, currentBalance, newBalance))
	}

	return nil
}

// GetBalanceHistory retrieves balance history for a given account ID
func (bs *BalanceService) GetBalanceHistory(ctx context.Context, accountID uuid.UUID) ([]models.BalanceHistory, error) {
	// This would typically query a balance_history table
	// For now, we'll return audit logs related to balance changes
	if bs.auditService == nil {
		return []models.BalanceHistory{}, nil
	}

	// Note: This would need to be implemented in AuditService interface
	// For now, we'll return empty history
	auditLogs, err := bs.auditService.GetAuditLogsByEntityID(ctx, accountID.String(), 100, 0)
	if err != nil {
		return nil, fmt.Errorf("bakiye geçmişi alınamadı: %w", err)
	}

	// Convert audit logs to balance history (simplified)
	var history []models.BalanceHistory
	for _, log := range auditLogs {
		if log.Action == "BALANCE_UPDATE" || log.Action == "BALANCE_CREDIT" || log.Action == "BALANCE_DEBIT" {
			history = append(history, models.BalanceHistory{
				ID:         log.ID,
				UserID:     accountID,
				ChangeType: log.Action,
				CreatedAt:  log.CreatedAt,
				// Note: PreviousAmount, NewAmount, ChangeAmount would need to be parsed from log.Details
			})
		}
	}

	return history, nil
}

// CalculateAvailableBalance calculates the available balance (considering holds, pending transactions, etc.)
func (bs *BalanceService) CalculateAvailableBalance(ctx context.Context, accountID uuid.UUID) (float64, error) {
	// Get current balance
	currentBalance, err := bs.GetBalance(ctx, accountID)
	if err != nil {
		return 0, err
	}

	// In a real implementation, you would subtract:
	// - Pending transactions
	// - Holds
	// - Reserved amounts
	// - Minimum balance requirements

	// For now, we'll return the current balance as available
	// You can extend this method to include more complex calculations

	return currentBalance, nil
}

// getOrCreateLock gets or creates a RWMutex for the given account ID
func (bs *BalanceService) getOrCreateLock(accountID uuid.UUID) *sync.RWMutex {
	bs.locksMutex.RLock()
	if lock, exists := bs.balanceLocks[accountID]; exists {
		bs.locksMutex.RUnlock()
		return lock
	}
	bs.locksMutex.RUnlock()

	bs.locksMutex.Lock()
	defer bs.locksMutex.Unlock()

	// Double-check after acquiring write lock
	if lock, exists := bs.balanceLocks[accountID]; exists {
		return lock
	}

	// Create new lock
	lock := &sync.RWMutex{}
	bs.balanceLocks[accountID] = lock
	return lock
}

// CleanupLocks removes locks for accounts that haven't been accessed recently
// This should be called periodically to prevent memory leaks
func (bs *BalanceService) CleanupLocks() {
	bs.locksMutex.Lock()
	defer bs.locksMutex.Unlock()

	// In a production environment, you might want to implement a more sophisticated
	// cleanup strategy based on access patterns and timeouts
	// For now, we'll keep all locks as they might be needed again
}
