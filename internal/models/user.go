package models

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
)

// User represents a user in the banking system
type User struct {
	ID           uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Username     string    `json:"username" gorm:"uniqueIndex;not null;size:50"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null;size:100"`
	PasswordHash string    `json:"-" gorm:"not null;size:255"`
	Role         UserRole  `json:"role" gorm:"not null;default:'customer'"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	// Relationships
	Balance      *Balance      `json:"balance,omitempty" gorm:"foreignKey:UserID"`
	Transactions []Transaction `json:"transactions,omitempty" gorm:"foreignKey:FromUserID"`
	ReceivedTx   []Transaction `json:"received_transactions,omitempty" gorm:"foreignKey:ToUserID"`
}

// UserRole defines the role of a user
type UserRole string

const (
	RoleCustomer UserRole = "customer"
	RoleAdmin    UserRole = "admin"
	RoleTeller   UserRole = "teller"
)

// TableName returns the table name for User model
func (User) TableName() string {
	return "users"
}

// UserCreateRequest represents the request to create a new user
type UserCreateRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role,omitempty"`
}

// UserResponse represents the response for user data (without sensitive info)
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Role      UserRole  `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// Validation methods for User

// ValidateUsername validates the username field
func (u *User) ValidateUsername() error {
	if len(u.Username) < 3 {
		return errors.New("kullanıcı adı en az 3 karakter olmalıdır")
	}
	if len(u.Username) > 50 {
		return errors.New("kullanıcı adı en fazla 50 karakter olabilir")
	}

	// Check for valid characters (alphanumeric and underscore only)
	validUsername := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !validUsername.MatchString(u.Username) {
		return errors.New("kullanıcı adı sadece harf, rakam ve alt çizgi içerebilir")
	}

	return nil
}

// ValidateEmail validates the email field
func (u *User) ValidateEmail() error {
	if u.Email == "" {
		return errors.New("e-posta adresi boş olamaz")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(u.Email) {
		return errors.New("geçersiz e-posta adresi formatı")
	}

	return nil
}

// ValidatePassword validates the password strength
func ValidatePassword(password string) error {
	if len(password) < 6 {
		return errors.New("şifre en az 6 karakter olmalıdır")
	}

	if len(password) > 128 {
		return errors.New("şifre en fazla 128 karakter olabilir")
	}

	var (
		hasUpper   = false
		hasLower   = false
		hasNumber  = false
		hasSpecial = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return errors.New("şifre en az bir büyük harf içermelidir")
	}
	if !hasLower {
		return errors.New("şifre en az bir küçük harf içermelidir")
	}
	if !hasNumber {
		return errors.New("şifre en az bir rakam içermelidir")
	}
	if !hasSpecial {
		return errors.New("şifre en az bir özel karakter içermelidir")
	}

	return nil
}

// ValidateRole validates the user role
func (u *User) ValidateRole() error {
	switch u.Role {
	case RoleCustomer, RoleAdmin, RoleTeller:
		return nil
	default:
		return errors.New("geçersiz kullanıcı rolü")
	}
}

// Validate validates all user fields
func (u *User) Validate() error {
	if err := u.ValidateUsername(); err != nil {
		return err
	}

	if err := u.ValidateEmail(); err != nil {
		return err
	}

	if err := u.ValidateRole(); err != nil {
		return err
	}

	return nil
}

// IsActive checks if user account is active
func (u *User) IsActive() bool {
	// For now, all users are considered active
	// This can be extended with an IsActive field in the future
	return true
}

// CanPerformTransaction checks if user can perform transactions
func (u *User) CanPerformTransaction() bool {
	return u.IsActive() && (u.Role == RoleCustomer || u.Role == RoleTeller)
}

// CanAccessAdminFeatures checks if user has admin privileges
func (u *User) CanAccessAdminFeatures() bool {
	return u.Role == RoleAdmin
}

// SanitizeInput sanitizes user input by trimming spaces
func (u *User) SanitizeInput() {
	u.Username = strings.TrimSpace(u.Username)
	u.Email = strings.TrimSpace(strings.ToLower(u.Email))
}

// JSON Marshaling/Unmarshaling methods

// MarshalJSON custom JSON marshaling for User
func (u User) MarshalJSON() ([]byte, error) {
	type Alias User
	return json.Marshal(&struct {
		*Alias
		Role string `json:"role"`
	}{
		Alias: (*Alias)(&u),
		Role:  string(u.Role),
	})
}

// UnmarshalJSON custom JSON unmarshaling for User
func (u *User) UnmarshalJSON(data []byte) error {
	type Alias User
	aux := &struct {
		*Alias
		Role string `json:"role"`
	}{
		Alias: (*Alias)(u),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	u.Role = UserRole(aux.Role)
	return nil
}

// ToJSON converts user to JSON string
func (u *User) ToJSON() (string, error) {
	jsonData, err := json.Marshal(u.ToResponse())
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// FromJSON creates user from JSON string
func (u *User) FromJSON(jsonStr string) error {
	return json.Unmarshal([]byte(jsonStr), u)
}
