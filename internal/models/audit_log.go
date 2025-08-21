package models

import (
	"time"

	"github.com/google/uuid"
)

// AuditLog represents an audit log entry for tracking system activities
type AuditLog struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	EntityType string     `json:"entity_type" gorm:"not null;size:50;index"`
	EntityID   string     `json:"entity_id" gorm:"not null;size:100;index"`
	Action     string     `json:"action" gorm:"not null;size:50"`
	Details    string     `json:"details" gorm:"type:text"`
	UserID     *uuid.UUID `json:"user_id,omitempty" gorm:"type:uuid;index"`
	IPAddress  string     `json:"ip_address,omitempty" gorm:"size:45"`
	UserAgent  string     `json:"user_agent,omitempty" gorm:"size:500"`
	CreatedAt  time.Time  `json:"created_at" gorm:"autoCreateTime;index"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// AuditAction defines common audit actions
type AuditAction string

const (
	AuditActionCreate   AuditAction = "create"
	AuditActionUpdate   AuditAction = "update"
	AuditActionDelete   AuditAction = "delete"
	AuditActionLogin    AuditAction = "login"
	AuditActionLogout   AuditAction = "logout"
	AuditActionTransfer AuditAction = "transfer"
	AuditActionDeposit  AuditAction = "deposit"
	AuditActionWithdraw AuditAction = "withdraw"
)

// EntityType defines common entity types for auditing
type EntityType string

const (
	EntityTypeUser        EntityType = "user"
	EntityTypeTransaction EntityType = "transaction"
	EntityTypeBalance     EntityType = "balance"
	EntityTypeAuth        EntityType = "auth"
)

// TableName returns the table name for AuditLog model
func (AuditLog) TableName() string {
	return "audit_logs"
}

// AuditLogResponse represents the response for audit log data
type AuditLogResponse struct {
	ID         uuid.UUID  `json:"id"`
	EntityType string     `json:"entity_type"`
	EntityID   string     `json:"entity_id"`
	Action     string     `json:"action"`
	Details    string     `json:"details"`
	UserID     *uuid.UUID `json:"user_id,omitempty"`
	IPAddress  string     `json:"ip_address,omitempty"`
	UserAgent  string     `json:"user_agent,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// ToResponse converts AuditLog to AuditLogResponse
func (a *AuditLog) ToResponse() *AuditLogResponse {
	return &AuditLogResponse{
		ID:         a.ID,
		EntityType: a.EntityType,
		EntityID:   a.EntityID,
		Action:     a.Action,
		Details:    a.Details,
		UserID:     a.UserID,
		IPAddress:  a.IPAddress,
		UserAgent:  a.UserAgent,
		CreatedAt:  a.CreatedAt,
	}
}

// CreateAuditLogRequest represents a request to create an audit log entry
type CreateAuditLogRequest struct {
	EntityType string     `json:"entity_type" binding:"required"`
	EntityID   string     `json:"entity_id" binding:"required"`
	Action     string     `json:"action" binding:"required"`
	Details    string     `json:"details,omitempty"`
	UserID     *uuid.UUID `json:"user_id,omitempty"`
	IPAddress  string     `json:"ip_address,omitempty"`
	UserAgent  string     `json:"user_agent,omitempty"`
}
