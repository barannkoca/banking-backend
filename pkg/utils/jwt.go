package utils

import (
	"errors"
	"fmt"
	"time"

	"github.com/barannkoca/banking-backend/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// JWTClaims represents the claims in a JWT token
type JWTClaims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey          string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
}

// DefaultJWTConfig returns default JWT configuration
func DefaultJWTConfig() *JWTConfig {
	return &JWTConfig{
		SecretKey:          "your-secret-key-change-in-production", // TODO: Load from environment
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour, // 7 days
	}
}

// GenerateAccessToken generates a new access token for a user
func GenerateAccessToken(user *models.User, config *JWTConfig) (string, error) {
	if config == nil {
		config = DefaultJWTConfig()
	}

	now := time.Now()
	claims := JWTClaims{
		UserID:   user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
		Role:     string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(config.AccessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "banking-backend",
			Subject:   user.ID.String(),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.SecretKey))
}

// GenerateRefreshToken generates a new refresh token for a user
func GenerateRefreshToken(user *models.User, config *JWTConfig) (string, error) {
	if config == nil {
		config = DefaultJWTConfig()
	}

	now := time.Now()
	claims := JWTClaims{
		UserID:   user.ID.String(),
		Username: user.Username,
		Email:    user.Email,
		Role:     string(user.Role),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(config.RefreshTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "banking-backend",
			Subject:   user.ID.String(),
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.SecretKey))
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString string, config *JWTConfig) (*JWTClaims, error) {
	if config == nil {
		config = DefaultJWTConfig()
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(config.SecretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("token parsing failed: %w", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshTokenPair generates new access and refresh tokens from a refresh token
func RefreshTokenPair(refreshToken string, config *JWTConfig) (string, string, error) {
	// Validate refresh token
	claims, err := ValidateToken(refreshToken, config)
	if err != nil {
		return "", "", fmt.Errorf("invalid refresh token: %w", err)
	}

	// Create user object from claims
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return "", "", fmt.Errorf("invalid user ID in token: %w", err)
	}

	user := &models.User{
		ID:       userID,
		Username: claims.Username,
		Email:    claims.Email,
		Role:     models.UserRole(claims.Role),
	}

	// Generate new token pair
	accessToken, err := GenerateAccessToken(user, config)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := GenerateRefreshToken(user, config)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, newRefreshToken, nil
}

// ExtractUserFromToken extracts user information from a JWT token
func ExtractUserFromToken(tokenString string, config *JWTConfig) (*models.User, error) {
	claims, err := ValidateToken(tokenString, config)
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID in token: %w", err)
	}

	user := &models.User{
		ID:       userID,
		Username: claims.Username,
		Email:    claims.Email,
		Role:     models.UserRole(claims.Role),
	}

	return user, nil
}

