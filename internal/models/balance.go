package models

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Balance represents a user's account balance with thread-safe operations
type Balance struct {
	UserID        uuid.UUID `json:"user_id" gorm:"type:uuid;primary_key"`
	Amount        float64   `json:"amount" gorm:"not null;type:decimal(15,2);default:0"`
	LastUpdatedAt time.Time `json:"last_updated_at" gorm:"autoUpdateTime"`

	// Thread-safety
	mutex sync.RWMutex `json:"-" gorm:"-"`

	// Relationships
	User *User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

// TableName returns the table name for Balance model
func (*Balance) TableName() string {
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

// Thread-safe balance operations

// GetAmount returns the current balance amount (thread-safe read)
func (b *Balance) GetAmount() float64 {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.Amount
}

// HasSufficientBalance checks if the balance is sufficient for a transaction (thread-safe)
func (b *Balance) HasSufficientBalance(amount float64) bool {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.Amount >= amount
}

// AddAmount adds the specified amount to the balance (thread-safe)
func (b *Balance) AddAmount(amount float64) error {
	if amount <= 0 {
		return errors.New("eklenen tutar sıfırdan büyük olmalıdır")
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()

	// Check for overflow
	if b.Amount+amount < b.Amount {
		return errors.New("bakiye taşması hatası")
	}

	b.Amount += amount
	b.LastUpdatedAt = time.Now()
	return nil
}

// SubtractAmount subtracts the specified amount from the balance (thread-safe)
func (b *Balance) SubtractAmount(amount float64) error {
	if amount <= 0 {
		return errors.New("çıkarılan tutar sıfırdan büyük olmalıdır")
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()

	if b.Amount < amount {
		return errors.New("yetersiz bakiye")
	}

	b.Amount -= amount
	b.LastUpdatedAt = time.Now()
	return nil
}

// TransferTo transfers amount from this balance to another balance (thread-safe)
func (b *Balance) TransferTo(targetBalance *Balance, amount float64) error {
	if amount <= 0 {
		return errors.New("transfer tutarı sıfırdan büyük olmalıdır")
	}

	if b.UserID == targetBalance.UserID {
		return errors.New("aynı hesaba transfer yapılamaz")
	}

	// Lock balances in a consistent order to prevent deadlocks
	var firstMutex, secondMutex *sync.RWMutex
	if b.UserID.String() < targetBalance.UserID.String() {
		firstMutex = &b.mutex
		secondMutex = &targetBalance.mutex
	} else {
		firstMutex = &targetBalance.mutex
		secondMutex = &b.mutex
	}

	firstMutex.Lock()
	defer firstMutex.Unlock()
	secondMutex.Lock()
	defer secondMutex.Unlock()

	// Check sufficient balance
	if b.Amount < amount {
		return errors.New("yetersiz bakiye")
	}

	// Check for overflow in target balance
	if targetBalance.Amount+amount < targetBalance.Amount {
		return errors.New("hedef bakiye taşması hatası")
	}

	// Perform the transfer
	b.Amount -= amount
	b.LastUpdatedAt = time.Now()

	targetBalance.Amount += amount
	targetBalance.LastUpdatedAt = time.Now()

	return nil
}

// SetAmount sets the balance amount (thread-safe) - use with caution
func (b *Balance) SetAmount(amount float64) error {
	if amount < 0 {
		return errors.New("bakiye negatif olamaz")
	}

	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.Amount = amount
	b.LastUpdatedAt = time.Now()
	return nil
}

// Freeze freezes the balance for read operations only
func (b *Balance) Freeze() {
	b.mutex.Lock()
}

// Unfreeze unfreezes the balance
func (b *Balance) Unfreeze() {
	b.mutex.Unlock()
}

// IsNegative checks if the balance is negative (thread-safe)
func (b *Balance) IsNegative() bool {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.Amount < 0
}

// IsZero checks if the balance is zero (thread-safe)
func (b *Balance) IsZero() bool {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.Amount == 0
}

// Validate validates the balance
func (b *Balance) Validate() error {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	if b.Amount < 0 {
		return errors.New("bakiye negatif olamaz")
	}

	if b.UserID == uuid.Nil {
		return errors.New("geçersiz kullanıcı ID")
	}

	return nil
}

// GetBalanceHistory returns balance change history (placeholder for future implementation)
func (b *Balance) GetBalanceHistory() []BalanceHistory {
	// This would be implemented with a separate BalanceHistory model
	// For now, returning empty slice
	return []BalanceHistory{}
}

// BalanceHistory represents a historical balance change record
type BalanceHistory struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID         uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	PreviousAmount float64    `json:"previous_amount" gorm:"type:decimal(15,2)"`
	NewAmount      float64    `json:"new_amount" gorm:"type:decimal(15,2)"`
	ChangeAmount   float64    `json:"change_amount" gorm:"type:decimal(15,2)"`
	ChangeType     string     `json:"change_type" gorm:"size:50"`
	TransactionID  *uuid.UUID `json:"transaction_id,omitempty" gorm:"type:uuid"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

// JSON Marshaling/Unmarshaling methods for Balance

// MarshalJSON custom JSON marshaling for Balance (thread-safe)
func (b *Balance) MarshalJSON() ([]byte, error) {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	type Alias Balance
	aux := struct {
		UserID        uuid.UUID `json:"user_id"`
		Amount        float64   `json:"amount"`
		LastUpdatedAt time.Time `json:"last_updated_at"`
	}{
		UserID:        b.UserID,
		Amount:        b.Amount,
		LastUpdatedAt: b.LastUpdatedAt,
	}

	return json.Marshal(&aux)
}

// UnmarshalJSON custom JSON unmarshaling for Balance
func (b *Balance) UnmarshalJSON(data []byte) error {
	type Alias Balance
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(b),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	// Initialize mutex after unmarshaling
	b.mutex = sync.RWMutex{}
	return nil
}

// ToJSON converts balance to JSON string (thread-safe)
func (b *Balance) ToJSON() (string, error) {
	jsonData, err := json.Marshal(b.ToResponse())
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// FromJSON creates balance from JSON string
func (b *Balance) FromJSON(jsonStr string) error {
	if err := json.Unmarshal([]byte(jsonStr), b); err != nil {
		return err
	}

	// Ensure mutex is initialized
	b.mutex = sync.RWMutex{}
	return nil
}

// ToJSON converts balance history to JSON string
func (bh *BalanceHistory) ToJSON() (string, error) {
	jsonData, err := json.Marshal(bh)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// FromJSON creates balance history from JSON string
func (bh *BalanceHistory) FromJSON(jsonStr string) error {
	return json.Unmarshal([]byte(jsonStr), bh)
}
