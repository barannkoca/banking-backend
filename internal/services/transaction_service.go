package services

import (
	"context"
	"fmt"
	"time"

	"github.com/barannkoca/banking-backend/internal/database"
	"github.com/barannkoca/banking-backend/internal/interfaces"
	"github.com/barannkoca/banking-backend/internal/models"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type TransactionService struct {
	transactionRepo interfaces.TransactionRepository
	balanceRepo     interfaces.BalanceRepository
	auditService    interfaces.AuditService
	logger          *zap.Logger
}

func NewTransactionService(
	transactionRepo interfaces.TransactionRepository,
	balanceRepo interfaces.BalanceRepository,
	auditService interfaces.AuditService,
	logger *zap.Logger,
) *TransactionService {
	return &TransactionService{
		transactionRepo: transactionRepo,
		balanceRepo:     balanceRepo,
		auditService:    auditService,
		logger:          logger,
	}
}

// Credit adds money to account with database transaction and rollback support
func (ts *TransactionService) Credit(ctx context.Context, accountID uuid.UUID, amount float64) error {
	// Validate credit amount
	if amount <= 0 {
		return fmt.Errorf("credit amount must be positive")
	}

	// Create transaction record
	transaction := &models.Transaction{
		ID:        uuid.New(),
		ToUserID:  &accountID,
		Amount:    amount,
		Type:      models.TransactionTypeDeposit,
		Status:    models.TransactionStatusPending,
		Reference: "Credit transaction",
		CreatedAt: time.Now(),
	}

	// Execute within database transaction
	err := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Save transaction record
		if err := tx.Create(transaction).Error; err != nil {
			return fmt.Errorf("failed to create transaction record: %w", err)
		}

		// 2. Get current balance
		var balance models.Balance
		if err := tx.Where("user_id = ?", accountID).First(&balance).Error; err != nil {
			return fmt.Errorf("failed to get current balance: %w", err)
		}

		// 3. Update balance
		newBalance := balance.Amount + amount
		if err := tx.Model(&models.Balance{}).
			Where("user_id = ?", accountID).
			Update("amount", newBalance).Error; err != nil {
			return fmt.Errorf("failed to update balance: %w", err)
		}

		// 4. Update transaction status to completed
		if err := tx.Model(&models.Transaction{}).
			Where("id = ?", transaction.ID).
			Update("status", models.TransactionStatusCompleted).Error; err != nil {
			return fmt.Errorf("failed to update transaction status: %w", err)
		}

		return nil
	})

	if err != nil {
		// Log failed transaction
		transaction.Status = models.TransactionStatusFailed
		if ts.auditService != nil {
			ts.auditService.LogTransactionActivity(ctx, transaction, "CREDIT_FAILED", err.Error())
		}
		ts.logger.Error("Credit transaction failed",
			zap.String("transaction_id", transaction.ID.String()),
			zap.String("account_id", accountID.String()),
			zap.Float64("amount", amount),
			zap.Error(err))
		return err
	}

	// Log successful transaction
	if ts.auditService != nil {
		ts.auditService.LogTransactionActivity(ctx, transaction, "CREDIT_COMPLETED", "Credit successful")
	}

	ts.logger.Info("Credit completed",
		zap.String("transaction_id", transaction.ID.String()),
		zap.String("account_id", accountID.String()),
		zap.Float64("amount", amount))

	return nil
}

// Debit removes money from account with database transaction and rollback support
func (ts *TransactionService) Debit(ctx context.Context, accountID uuid.UUID, amount float64) error {
	// Validate debit amount
	if amount <= 0 {
		return fmt.Errorf("debit amount must be positive")
	}

	// Create transaction record
	transaction := &models.Transaction{
		ID:         uuid.New(),
		FromUserID: &accountID,
		Amount:     amount,
		Type:       models.TransactionTypeWithdraw,
		Status:     models.TransactionStatusPending,
		Reference:  "Debit transaction",
		CreatedAt:  time.Now(),
	}

	// Execute within database transaction
	err := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Save transaction record
		if err := tx.Create(transaction).Error; err != nil {
			return fmt.Errorf("failed to create transaction record: %w", err)
		}

		// 2. Get current balance and check sufficient funds
		var balance models.Balance
		if err := tx.Where("user_id = ?", accountID).First(&balance).Error; err != nil {
			return fmt.Errorf("failed to get current balance: %w", err)
		}

		if balance.Amount < amount {
			return fmt.Errorf("insufficient balance: current=%f, required=%f", balance.Amount, amount)
		}

		// 3. Update balance
		newBalance := balance.Amount - amount
		if err := tx.Model(&models.Balance{}).
			Where("user_id = ?", accountID).
			Update("amount", newBalance).Error; err != nil {
			return fmt.Errorf("failed to update balance: %w", err)
		}

		// 4. Update transaction status to completed
		if err := tx.Model(&models.Transaction{}).
			Where("id = ?", transaction.ID).
			Update("status", models.TransactionStatusCompleted).Error; err != nil {
			return fmt.Errorf("failed to update transaction status: %w", err)
		}

		return nil
	})

	if err != nil {
		// Log failed transaction
		transaction.Status = models.TransactionStatusFailed
		if ts.auditService != nil {
			ts.auditService.LogTransactionActivity(ctx, transaction, "DEBIT_FAILED", err.Error())
		}
		ts.logger.Error("Debit transaction failed",
			zap.String("transaction_id", transaction.ID.String()),
			zap.String("account_id", accountID.String()),
			zap.Float64("amount", amount),
			zap.Error(err))
		return err
	}

	// Log successful transaction
	if ts.auditService != nil {
		ts.auditService.LogTransactionActivity(ctx, transaction, "DEBIT_COMPLETED", "Debit successful")
	}

	ts.logger.Info("Debit completed",
		zap.String("transaction_id", transaction.ID.String()),
		zap.String("account_id", accountID.String()),
		zap.Float64("amount", amount))

	return nil
}

// Transfer transfers money between two accounts with database transaction and rollback support
func (ts *TransactionService) Transfer(ctx context.Context, fromAccountID, toAccountID uuid.UUID, amount float64) error {
	// Validate transfer
	if fromAccountID == toAccountID {
		return fmt.Errorf("cannot transfer to same account")
	}
	if amount <= 0 {
		return fmt.Errorf("transfer amount must be positive")
	}

	// Create transaction record
	transaction := &models.Transaction{
		ID:         uuid.New(),
		FromUserID: &fromAccountID,
		ToUserID:   &toAccountID,
		Amount:     amount,
		Type:       models.TransactionTypeTransfer,
		Status:     models.TransactionStatusPending,
		Reference:  "Transfer transaction",
		CreatedAt:  time.Now(),
	}

	// Execute within database transaction
	err := database.GetDB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Save transaction record
		if err := tx.Create(transaction).Error; err != nil {
			return fmt.Errorf("failed to create transaction record: %w", err)
		}

		// 2. Get current balances
		var fromBalance, toBalance models.Balance

		// Get from account balance
		if err := tx.Where("user_id = ?", fromAccountID).First(&fromBalance).Error; err != nil {
			return fmt.Errorf("failed to get from account balance: %w", err)
		}

		// Check sufficient balance
		if fromBalance.Amount < amount {
			return fmt.Errorf("insufficient balance in from account: current=%f, required=%f", fromBalance.Amount, amount)
		}

		// Get to account balance
		if err := tx.Where("user_id = ?", toAccountID).First(&toBalance).Error; err != nil {
			return fmt.Errorf("failed to get to account balance: %w", err)
		}

		// 3. Update balances atomically
		// Subtract from sender
		if err := tx.Model(&models.Balance{}).
			Where("user_id = ?", fromAccountID).
			Update("amount", gorm.Expr("amount - ?", amount)).Error; err != nil {
			return fmt.Errorf("failed to subtract from sender balance: %w", err)
		}

		// Add to receiver
		if err := tx.Model(&models.Balance{}).
			Where("user_id = ?", toAccountID).
			Update("amount", gorm.Expr("amount + ?", amount)).Error; err != nil {
			return fmt.Errorf("failed to add to receiver balance: %w", err)
		}

		// 4. Update transaction status to completed
		if err := tx.Model(&models.Transaction{}).
			Where("id = ?", transaction.ID).
			Update("status", models.TransactionStatusCompleted).Error; err != nil {
			return fmt.Errorf("failed to update transaction status: %w", err)
		}

		return nil
	})

	if err != nil {
		// Log failed transaction
		transaction.Status = models.TransactionStatusFailed
		if ts.auditService != nil {
			ts.auditService.LogTransactionActivity(ctx, transaction, "TRANSFER_FAILED", err.Error())
		}
		ts.logger.Error("Transfer transaction failed",
			zap.String("transaction_id", transaction.ID.String()),
			zap.String("from_account", fromAccountID.String()),
			zap.String("to_account", toAccountID.String()),
			zap.Float64("amount", amount),
			zap.Error(err))
		return err
	}

	// Log successful transaction
	if ts.auditService != nil {
		ts.auditService.LogTransactionActivity(ctx, transaction, "TRANSFER_COMPLETED", "Transfer successful")
	}

	ts.logger.Info("Transfer completed",
		zap.String("transaction_id", transaction.ID.String()),
		zap.String("from_account", fromAccountID.String()),
		zap.String("to_account", toAccountID.String()),
		zap.Float64("amount", amount))

	return nil
}

// Helper methods
func (ts *TransactionService) CanPerformTransaction(ctx context.Context, accountID uuid.UUID, amount float64) (bool, error) {
	balance, err := ts.balanceRepo.GetBalance(ctx, accountID)
	if err != nil {
		return false, err
	}
	return balance >= amount, nil
}
