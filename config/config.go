package config

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Database  DatabaseConfig
	Server    ServerConfig
	JWT       JWTConfig
	App       AppConfig
	RateLimit RateLimitConfig
	Security  SecurityConfig
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string
	Host string
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	Secret string
}

// AppConfig holds application configuration
type AppConfig struct {
	Environment string
	LogLevel    string
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	GlobalRequestsPerSecond  float64
	GlobalBurst              int
	AuthRequestsPerSecond    float64
	AuthBurst                int
	BankingRequestsPerSecond float64
	BankingBurst             int
}

// SecurityConfig holds security configuration
type SecurityConfig struct {
	AllowedOrigins []string
	TrustedProxies []string
	EnableHSTS     bool
	EnableCSP      bool
}

var cfg *Config

// Load loads configuration from environment variables and .env file
func Load() (*Config, error) {
	// Load .env file (ignore error in production where env vars might be set directly)
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	cfg = &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "banking_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		JWT: JWTConfig{
			Secret: getEnv("JWT_SECRET", "your-super-secure-jwt-secret-key-here"),
		},
		App: AppConfig{
			Environment: getEnv("ENVIRONMENT", "development"),
			LogLevel:    getEnv("LOG_LEVEL", "info"),
		},
		RateLimit: RateLimitConfig{
			GlobalRequestsPerSecond:  getEnvAsFloat("RATE_LIMIT_GLOBAL_RPS", 10.0),
			GlobalBurst:              getEnvAsInt("RATE_LIMIT_GLOBAL_BURST", 20),
			AuthRequestsPerSecond:    getEnvAsFloat("RATE_LIMIT_AUTH_RPS", 1.0),
			AuthBurst:                getEnvAsInt("RATE_LIMIT_AUTH_BURST", 3),
			BankingRequestsPerSecond: getEnvAsFloat("RATE_LIMIT_BANKING_RPS", 5.0),
			BankingBurst:             getEnvAsInt("RATE_LIMIT_BANKING_BURST", 10),
		},
		Security: SecurityConfig{
			AllowedOrigins: []string{"http://localhost:3000", "https://banking-frontend.com"},
			TrustedProxies: []string{"127.0.0.1", "::1"},
			EnableHSTS:     getEnvAsBool("ENABLE_HSTS", true),
			EnableCSP:      getEnvAsBool("ENABLE_CSP", true),
		},
	}

	// Validate required configurations
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// Get returns the loaded configuration
func Get() *Config {
	if cfg == nil {
		log.Fatal("Configuration not loaded. Call config.Load() first")
	}
	return cfg
}

// GetDatabaseDSN returns the database connection string
func (c *Config) GetDatabaseDSN() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=UTC",
		c.Database.Host,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.Port,
		c.Database.SSLMode,
	)
}

// GetServerAddress returns the server address
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

// validate validates the configuration
func (c *Config) validate() error {
	// Database validation
	if c.Database.User == "" {
		return fmt.Errorf("DB_USER is required")
	}
	if c.Database.DBName == "" {
		return fmt.Errorf("DB_NAME is required")
	}

	// Server validation
	if port, err := strconv.Atoi(c.Server.Port); err != nil || port <= 0 || port > 65535 {
		return fmt.Errorf("invalid SERVER_PORT: %s", c.Server.Port)
	}

	// JWT validation
	if len(c.JWT.Secret) < 32 {
		log.Println("Warning: JWT secret is shorter than 32 characters")
	}

	return nil
}

// getEnv gets environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvAsInt gets environment variable as integer with fallback
func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}

// getEnvAsBool gets environment variable as boolean with fallback
func getEnvAsBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return fallback
}

// getEnvAsFloat gets environment variable as float64 with fallback
func getEnvAsFloat(key string, fallback float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
			return floatVal
		}
	}
	return fallback
}
