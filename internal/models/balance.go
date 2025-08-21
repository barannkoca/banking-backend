package models

import (
	"time"

	"github.com/google/uuid"
)

// Balance represents a user's account balance
type Balance struct {
	UserID        uuid.UUID `json:"user_id" gorm:"type:uuid;primary_key"`
	Amount        float64   `json:"amount" gorm:"not null;type:decimal(15,2);default:0"`
	LastUpdatedAt time.Time `json:"last_updated_at" gorm:"autoUpdateTime"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName returns the table name for Balance model
func (Balance) TableName() string {
	return "balances"
}

// BalanceResponse represents the response for balance data
type BalanceResponse struct {
	UserID        uuid.UUID `json:"user_id"`
	Amount        float64   `json:"amount"`
	LastUpdatedAt time.Time `json:"last_updated_at"`
}

// ToResponse converts Balance to BalanceResponse
func (b *Balance) ToResponse() *BalanceResponse {
	return &BalanceResponse{
		UserID:        b.UserID,
		Amount:        b.Amount,
		LastUpdatedAt: b.LastUpdatedAt,
	}
}

// HasSufficientBalance checks if the balance is sufficient for a transaction
func (b *Balance) HasSufficientBalance(amount float64) bool {
	return b.Amount >= amount
}

// AddAmount adds the specified amount to the balance
func (b *Balance) AddAmount(amount float64) {
	b.Amount += amount
	b.LastUpdatedAt = time.Now()
}

// SubtractAmount subtracts the specified amount from the balance
func (b *Balance) SubtractAmount(amount float64) bool {
	if !b.HasSufficientBalance(amount) {
		return false
	}
	b.Amount -= amount
	b.LastUpdatedAt = time.Now()
	return true
}
