package models

import (
	"time"

	"github.com/google/uuid"
)

// Transaction represents a financial transaction
type Transaction struct {
	ID         uuid.UUID         `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	FromUserID *uuid.UUID        `json:"from_user_id" gorm:"type:uuid;index"`
	ToUserID   *uuid.UUID        `json:"to_user_id" gorm:"type:uuid;index"`
	Amount     float64           `json:"amount" gorm:"not null;type:decimal(15,2)"`
	Type       TransactionType   `json:"type" gorm:"not null"`
	Status     TransactionStatus `json:"status" gorm:"not null;default:'pending'"`
	Reference  string            `json:"reference,omitempty" gorm:"size:100"`
	CreatedAt  time.Time         `json:"created_at" gorm:"autoCreateTime"`

	// Relationships
	FromUser *User `json:"from_user,omitempty" gorm:"foreignKey:FromUserID"`
	ToUser   *User `json:"to_user,omitempty" gorm:"foreignKey:ToUserID"`
}

// TransactionType defines the type of transaction
type TransactionType string

const (
	TransactionTypeTransfer TransactionType = "transfer"
	TransactionTypeDeposit  TransactionType = "deposit"
	TransactionTypeWithdraw TransactionType = "withdraw"
	TransactionTypePayment  TransactionType = "payment"
	TransactionTypeRefund   TransactionType = "refund"
)

// TransactionStatus defines the status of a transaction
type TransactionStatus string

const (
	TransactionStatusPending   TransactionStatus = "pending"
	TransactionStatusCompleted TransactionStatus = "completed"
	TransactionStatusFailed    TransactionStatus = "failed"
	TransactionStatusCancelled TransactionStatus = "cancelled"
)

// TableName returns the table name for Transaction model
func (Transaction) TableName() string {
	return "transactions"
}

// TransferRequest represents a money transfer request
type TransferRequest struct {
	ToUserID  uuid.UUID `json:"to_user_id" binding:"required"`
	Amount    float64   `json:"amount" binding:"required,gt=0"`
	Reference string    `json:"reference,omitempty" binding:"max=100"`
}

// DepositRequest represents a deposit request
type DepositRequest struct {
	Amount    float64 `json:"amount" binding:"required,gt=0"`
	Reference string  `json:"reference,omitempty" binding:"max=100"`
}

// WithdrawRequest represents a withdrawal request
type WithdrawRequest struct {
	Amount    float64 `json:"amount" binding:"required,gt=0"`
	Reference string  `json:"reference,omitempty" binding:"max=100"`
}

// TransactionResponse represents the response for transaction data
type TransactionResponse struct {
	ID         uuid.UUID         `json:"id"`
	FromUserID *uuid.UUID        `json:"from_user_id"`
	ToUserID   *uuid.UUID        `json:"to_user_id"`
	Amount     float64           `json:"amount"`
	Type       TransactionType   `json:"type"`
	Status     TransactionStatus `json:"status"`
	Reference  string            `json:"reference,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
}

// ToResponse converts Transaction to TransactionResponse
func (t *Transaction) ToResponse() *TransactionResponse {
	return &TransactionResponse{
		ID:         t.ID,
		FromUserID: t.FromUserID,
		ToUserID:   t.ToUserID,
		Amount:     t.Amount,
		Type:       t.Type,
		Status:     t.Status,
		Reference:  t.Reference,
		CreatedAt:  t.CreatedAt,
	}
}
