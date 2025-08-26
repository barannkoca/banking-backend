package models

import (
	"encoding/json"
	"errors"
	"fmt"
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
	TransactionStatusRefund    TransactionStatus = "refund"
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

// State management methods for Transaction

// CanTransitionTo checks if the transaction can transition to the target status
func (t *Transaction) CanTransitionTo(targetStatus TransactionStatus) bool {
	switch t.Status {
	case TransactionStatusPending:
		return targetStatus == TransactionStatusCompleted ||
			targetStatus == TransactionStatusFailed ||
			targetStatus == TransactionStatusCancelled
	case TransactionStatusCompleted:
		return targetStatus == TransactionStatusRefund
	case TransactionStatusFailed:
		return false // Failed transactions cannot transition
	case TransactionStatusCancelled:
		return false // Cancelled transactions cannot transition
	default:
		return false
	}
}

// TransitionTo changes the transaction status if the transition is valid
func (t *Transaction) TransitionTo(targetStatus TransactionStatus) error {
	if !t.CanTransitionTo(targetStatus) {
		return fmt.Errorf("işlem durumu %s'den %s'e geçiş yapılamaz", t.Status, targetStatus)
	}

	t.Status = targetStatus
	return nil
}

// MarkAsCompleted marks the transaction as completed
func (t *Transaction) MarkAsCompleted() error {
	return t.TransitionTo(TransactionStatusCompleted)
}

// MarkAsFailed marks the transaction as failed
func (t *Transaction) MarkAsFailed() error {
	return t.TransitionTo(TransactionStatusFailed)
}

// MarkAsCancelled marks the transaction as cancelled
func (t *Transaction) MarkAsCancelled() error {
	return t.TransitionTo(TransactionStatusCancelled)
}

// IsCompleted checks if the transaction is completed
func (t *Transaction) IsCompleted() bool {
	return t.Status == TransactionStatusCompleted
}

// IsPending checks if the transaction is pending
func (t *Transaction) IsPending() bool {
	return t.Status == TransactionStatusPending
}

// IsFailed checks if the transaction is failed
func (t *Transaction) IsFailed() bool {
	return t.Status == TransactionStatusFailed
}

// IsCancelled checks if the transaction is cancelled
func (t *Transaction) IsCancelled() bool {
	return t.Status == TransactionStatusCancelled
}

// IsFinalized checks if the transaction is in a final state (cannot be changed)
func (t *Transaction) IsFinalized() bool {
	return t.Status == TransactionStatusCompleted ||
		t.Status == TransactionStatusFailed ||
		t.Status == TransactionStatusCancelled
}

// Validate validates the transaction fields
func (t *Transaction) Validate() error {
	if t.Amount <= 0 {
		return errors.New("işlem tutarı sıfırdan büyük olmalıdır")
	}

	if t.Amount > 1000000 { // Max transaction limit
		return errors.New("işlem tutarı çok yüksek (maksimum: 1,000,000)")
	}

	// Validate transaction type
	switch t.Type {
	case TransactionTypeTransfer, TransactionTypeDeposit, TransactionTypeWithdraw,
		TransactionTypePayment, TransactionTypeRefund:
		// Valid types
	default:
		return errors.New("geçersiz işlem türü")
	}

	// Validate status
	switch t.Status {
	case TransactionStatusPending, TransactionStatusCompleted, TransactionStatusFailed,
		TransactionStatusCancelled:
		// Valid statuses
	default:
		return errors.New("geçersiz işlem durumu")
	}

	// Validate user IDs based on transaction type
	if t.Type == TransactionTypeTransfer {
		if t.FromUserID == nil || t.ToUserID == nil {
			return errors.New("transfer işlemi için gönderen ve alıcı kullanıcılar gereklidir")
		}
		if *t.FromUserID == *t.ToUserID {
			return errors.New("kullanıcı kendisine transfer yapamaz")
		}
	}

	if t.Type == TransactionTypeDeposit {
		if t.ToUserID == nil {
			return errors.New("para yatırma işlemi için alıcı kullanıcı gereklidir")
		}
		if t.FromUserID != nil {
			return errors.New("para yatırma işleminde gönderen kullanıcı belirtilmemelidir")
		}
	}

	if t.Type == TransactionTypeWithdraw {
		if t.FromUserID == nil {
			return errors.New("para çekme işlemi için gönderen kullanıcı gereklidir")
		}
		if t.ToUserID != nil {
			return errors.New("para çekme işleminde alıcı kullanıcı belirtilmemelidir")
		}
	}

	return nil
}

// GetDescription returns a human-readable description of the transaction
func (t *Transaction) GetDescription() string {
	switch t.Type {
	case TransactionTypeTransfer:
		return fmt.Sprintf("%.2f TL transfer işlemi", t.Amount)
	case TransactionTypeDeposit:
		return fmt.Sprintf("%.2f TL para yatırma", t.Amount)
	case TransactionTypeWithdraw:
		return fmt.Sprintf("%.2f TL para çekme", t.Amount)
	case TransactionTypePayment:
		return fmt.Sprintf("%.2f TL ödeme", t.Amount)
	case TransactionTypeRefund:
		return fmt.Sprintf("%.2f TL iade", t.Amount)
	default:
		return fmt.Sprintf("%.2f TL işlem", t.Amount)
	}
}

// GetStatusDescription returns a human-readable status description
func (t *Transaction) GetStatusDescription() string {
	switch t.Status {
	case TransactionStatusPending:
		return "Beklemede"
	case TransactionStatusCompleted:
		return "Tamamlandı"
	case TransactionStatusFailed:
		return "Başarısız"
	case TransactionStatusCancelled:
		return "İptal Edildi"
	default:
		return "Bilinmeyen"
	}
}

// JSON Marshaling/Unmarshaling methods

// MarshalJSON custom JSON marshaling for Transaction
func (t Transaction) MarshalJSON() ([]byte, error) {
	type Alias Transaction
	return json.Marshal(&struct {
		*Alias
		Type   string `json:"type"`
		Status string `json:"status"`
	}{
		Alias:  (*Alias)(&t),
		Type:   string(t.Type),
		Status: string(t.Status),
	})
}

// UnmarshalJSON custom JSON unmarshaling for Transaction
func (t *Transaction) UnmarshalJSON(data []byte) error {
	type Alias Transaction
	aux := &struct {
		*Alias
		Type   string `json:"type"`
		Status string `json:"status"`
	}{
		Alias: (*Alias)(t),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	t.Type = TransactionType(aux.Type)
	t.Status = TransactionStatus(aux.Status)
	return nil
}

// ToJSON converts transaction to JSON string
func (t *Transaction) ToJSON() (string, error) {
	jsonData, err := json.Marshal(t.ToResponse())
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// FromJSON creates transaction from JSON string
func (t *Transaction) FromJSON(jsonStr string) error {
	return json.Unmarshal([]byte(jsonStr), t)
}
