package models

import (
	"time"
)

// AuthRegisterRequest represents the request for user registration
type AuthRegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role,omitempty"`
}

// AuthLoginRequest represents the request for user login
type AuthLoginRequest struct {
	UsernameOrEmail string `json:"username_or_email" binding:"required"`
	Password        string `json:"password" binding:"required"`
}

// AuthRefreshRequest represents the request for token refresh
type AuthRefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// AuthResponse represents the response for authentication operations
type AuthResponse struct {
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	TokenType    string        `json:"token_type"`
	ExpiresIn    int64         `json:"expires_in"`
	User         *UserResponse `json:"user"`
}

// AuthRefreshResponse represents the response for token refresh
type AuthRefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// TokenInfo represents token information
type TokenInfo struct {
	Token     string    `json:"token"`
	TokenType string    `json:"token_type"`
	ExpiresAt time.Time `json:"expires_at"`
	IssuedAt  time.Time `json:"issued_at"`
}

// AuthError represents authentication error response
type AuthError struct {
	Error      string `json:"error"`
	Message    string `json:"message"`
	Code       string `json:"code,omitempty"`
	RetryAfter int64  `json:"retry_after,omitempty"`
}

// AuthSuccess represents successful authentication response
type AuthSuccess struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// UserUpdateRequest represents the request for user update
type UserUpdateRequest struct {
	Username string `json:"username,omitempty" binding:"omitempty,min=3,max=50"`
	Email    string `json:"email,omitempty" binding:"omitempty,email"`
	Role     string `json:"role,omitempty" binding:"omitempty,oneof=admin teller customer"`
}

// NewAuthResponse creates a new AuthResponse
func NewAuthResponse(accessToken, refreshToken string, expiresIn int64, user *User) *AuthResponse {
	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
		User:         user.ToResponse(),
	}
}

// NewAuthRefreshResponse creates a new AuthRefreshResponse
func NewAuthRefreshResponse(accessToken, refreshToken string, expiresIn int64) *AuthRefreshResponse {
	return &AuthRefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    expiresIn,
	}
}

// NewAuthError creates a new AuthError
func NewAuthError(errorType, message, code string) *AuthError {
	return &AuthError{
		Error:   errorType,
		Message: message,
		Code:    code,
	}
}

// NewAuthSuccess creates a new AuthSuccess
func NewAuthSuccess(message string, data interface{}) *AuthSuccess {
	return &AuthSuccess{
		Message: message,
		Data:    data,
	}
}
