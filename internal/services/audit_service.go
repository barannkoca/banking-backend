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
)

type AuditService struct {
	logger *zap.Logger
}

func NewAuditService(logger *zap.Logger) interfaces.AuditService {
	return &AuditService{
		logger: logger,
	}
}

// LogUserActivity logs user-related activities
func (as *AuditService) LogUserActivity(ctx context.Context, userID uuid.UUID, action, entityType, entityID, details string) error {
	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		UserID:     &userID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Details:    details,
		IPAddress:  "", // TODO: Extract from context
		UserAgent:  "", // TODO: Extract from context
		CreatedAt:  time.Now(),
	}

	if err := database.GetDB().WithContext(ctx).Create(auditLog).Error; err != nil {
		as.logger.Error("Failed to log user activity",
			zap.String("user_id", userID.String()),
			zap.String("action", action),
			zap.Error(err))
		return fmt.Errorf("failed to log user activity: %w", err)
	}

	as.logger.Info("User activity logged",
		zap.String("user_id", userID.String()),
		zap.String("action", action),
		zap.String("entity_type", entityType),
		zap.String("entity_id", entityID))

	return nil
}

// LogTransactionActivity logs transaction-related activities
func (as *AuditService) LogTransactionActivity(ctx context.Context, transaction *models.Transaction, action, details string) error {
	var userID uuid.UUID
	if transaction.FromUserID != nil {
		userID = *transaction.FromUserID
	} else if transaction.ToUserID != nil {
		userID = *transaction.ToUserID
	}

	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		UserID:     &userID,
		Action:     action,
		EntityType: "transaction",
		EntityID:   transaction.ID.String(),
		Details: fmt.Sprintf("Transaction %s: %s (Amount: %f, Type: %s, Status: %s)",
			transaction.ID.String(), details, transaction.Amount, transaction.Type, transaction.Status),
		IPAddress: "", // TODO: Extract from context
		UserAgent: "", // TODO: Extract from context
		CreatedAt: time.Now(),
	}

	if err := database.GetDB().WithContext(ctx).Create(auditLog).Error; err != nil {
		as.logger.Error("Failed to log transaction activity",
			zap.String("transaction_id", transaction.ID.String()),
			zap.String("action", action),
			zap.Error(err))
		return fmt.Errorf("failed to log transaction activity: %w", err)
	}

	as.logger.Info("Transaction activity logged",
		zap.String("transaction_id", transaction.ID.String()),
		zap.String("action", action),
		zap.Float64("amount", transaction.Amount))

	return nil
}

// LogSystemActivity logs system-level activities
func (as *AuditService) LogSystemActivity(ctx context.Context, action, details string) error {
	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		UserID:     nil, // System activity, no specific user
		Action:     action,
		EntityType: "system",
		EntityID:   "system",
		Details:    details,
		IPAddress:  "", // TODO: Extract from context
		UserAgent:  "", // TODO: Extract from context
		CreatedAt:  time.Now(),
	}

	if err := database.GetDB().WithContext(ctx).Create(auditLog).Error; err != nil {
		as.logger.Error("Failed to log system activity",
			zap.String("action", action),
			zap.Error(err))
		return fmt.Errorf("failed to log system activity: %w", err)
	}

	as.logger.Info("System activity logged",
		zap.String("action", action),
		zap.String("details", details))

	return nil
}

// GetAuditLogsByUserID retrieves audit logs for a specific user
func (as *AuditService) GetAuditLogsByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*models.AuditLog, error) {
	var auditLogs []*models.AuditLog

	query := database.GetDB().WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset)

	if err := query.Find(&auditLogs).Error; err != nil {
		return nil, fmt.Errorf("failed to get audit logs by user ID: %w", err)
	}

	return auditLogs, nil
}

// GetAuditLogsByEntityID retrieves audit logs for a specific entity
func (as *AuditService) GetAuditLogsByEntityID(ctx context.Context, entityID string, limit, offset int) ([]*models.AuditLog, error) {
	var auditLogs []*models.AuditLog

	query := database.GetDB().WithContext(ctx).
		Where("entity_id = ?", entityID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset)

	if err := query.Find(&auditLogs).Error; err != nil {
		return nil, fmt.Errorf("failed to get audit logs by entity ID: %w", err)
	}

	return auditLogs, nil
}

// GetAuditLogsByAction retrieves audit logs for a specific action
func (as *AuditService) GetAuditLogsByAction(ctx context.Context, action string, limit, offset int) ([]*models.AuditLog, error) {
	var auditLogs []*models.AuditLog

	query := database.GetDB().WithContext(ctx).
		Where("action = ?", action).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset)

	if err := query.Find(&auditLogs).Error; err != nil {
		return nil, fmt.Errorf("failed to get audit logs by action: %w", err)
	}

	return auditLogs, nil
}

// GetAllAuditLogs retrieves all audit logs with pagination
func (as *AuditService) GetAllAuditLogs(ctx context.Context, limit, offset int) ([]*models.AuditLog, error) {
	var auditLogs []*models.AuditLog

	query := database.GetDB().WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset)

	if err := query.Find(&auditLogs).Error; err != nil {
		return nil, fmt.Errorf("failed to get all audit logs: %w", err)
	}

	return auditLogs, nil
}

// DeleteOldAuditLogs deletes audit logs older than specified days
func (as *AuditService) DeleteOldAuditLogs(ctx context.Context, olderThanDays int) error {
	cutoffDate := time.Now().AddDate(0, 0, -olderThanDays)

	result := database.GetDB().WithContext(ctx).
		Where("created_at < ?", cutoffDate).
		Delete(&models.AuditLog{})

	if result.Error != nil {
		return fmt.Errorf("failed to delete old audit logs: %w", result.Error)
	}

	as.logger.Info("Old audit logs deleted",
		zap.Int("older_than_days", olderThanDays),
		zap.Int64("deleted_count", result.RowsAffected))

	return nil
}

// GetAuditStatistics returns audit statistics
func (as *AuditService) GetAuditStatistics(ctx context.Context) (map[string]int64, error) {
	stats := make(map[string]int64)

	// Get total count
	var totalCount int64
	if err := database.GetDB().WithContext(ctx).Model(&models.AuditLog{}).Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to get total audit log count: %w", err)
	}
	stats["total"] = totalCount

	// Get counts by action type
	var actionStats []struct {
		Action string `json:"action"`
		Count  int64  `json:"count"`
	}

	if err := database.GetDB().WithContext(ctx).
		Model(&models.AuditLog{}).
		Select("action, count(*) as count").
		Group("action").
		Find(&actionStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get action statistics: %w", err)
	}

	for _, stat := range actionStats {
		stats[fmt.Sprintf("action_%s", stat.Action)] = stat.Count
	}

	// Get counts by entity type
	var entityStats []struct {
		EntityType string `json:"entity_type"`
		Count      int64  `json:"count"`
	}

	if err := database.GetDB().WithContext(ctx).
		Model(&models.AuditLog{}).
		Select("entity_type, count(*) as count").
		Group("entity_type").
		Find(&entityStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get entity type statistics: %w", err)
	}

	for _, stat := range entityStats {
		stats[fmt.Sprintf("entity_%s", stat.EntityType)] = stat.Count
	}

	return stats, nil
}
