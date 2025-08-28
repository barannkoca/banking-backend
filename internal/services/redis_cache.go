package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/barannkoca/banking-backend/internal/interfaces"
	"github.com/barannkoca/banking-backend/internal/models"
	"github.com/barannkoca/banking-backend/pkg/logger"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisCacheService implements CacheService interface using Redis
type RedisCacheService struct {
	client *redis.Client
	logger *zap.Logger
	stats  *CacheStats
	mu     sync.RWMutex
}

// CacheStats holds cache statistics
type CacheStats struct {
	Hits    int64 `json:"hits"`
	Misses  int64 `json:"misses"`
	Sets    int64 `json:"sets"`
	Deletes int64 `json:"deletes"`
	Errors  int64 `json:"errors"`
}

// NewRedisCacheService creates a new Redis cache service
func NewRedisCacheService(addr, password string, db int) (interfaces.CacheService, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
		PoolSize: 10,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.GetLogger().Info("Redis cache service initialized",
		zap.String("addr", addr),
		zap.Int("db", db))

	return &RedisCacheService{
		client: client,
		logger: logger.GetLogger(),
		stats:  &CacheStats{},
	}, nil
}

// Basic operations

func (r *RedisCacheService) Get(ctx context.Context, key string) (interface{}, error) {
	result, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			r.incrementMisses()
			return nil, nil // Cache miss, not an error
		}
		r.incrementErrors()
		return nil, fmt.Errorf("failed to get key %s: %w", key, err)
	}

	r.incrementHits()
	return result, nil
}

func (r *RedisCacheService) Set(ctx context.Context, key string, value interface{}, ttl int) error {
	var stringValue string
	switch v := value.(type) {
	case string:
		stringValue = v
	case []byte:
		stringValue = string(v)
	default:
		// Try to marshal to JSON
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			r.incrementErrors()
			return fmt.Errorf("failed to marshal value for key %s: %w", key, err)
		}
		stringValue = string(jsonBytes)
	}

	expiration := time.Duration(ttl) * time.Second
	err := r.client.Set(ctx, key, stringValue, expiration).Err()
	if err != nil {
		r.incrementErrors()
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}

	r.incrementSets()
	return nil
}

func (r *RedisCacheService) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		r.incrementErrors()
		return fmt.Errorf("failed to delete key %s: %w", key, err)
	}

	r.incrementDeletes()
	return nil
}

func (r *RedisCacheService) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		r.incrementErrors()
		return false, fmt.Errorf("failed to check existence of key %s: %w", key, err)
	}

	return result > 0, nil
}

func (r *RedisCacheService) Clear(ctx context.Context) error {
	err := r.client.FlushDB(ctx).Err()
	if err != nil {
		r.incrementErrors()
		return fmt.Errorf("failed to clear cache: %w", err)
	}
	return nil
}

// User-specific caching

func (r *RedisCacheService) CacheUser(ctx context.Context, user *models.User, ttl int) error {
	key := fmt.Sprintf("user:%s", user.ID.String())
	return r.Set(ctx, key, user, ttl)
}

func (r *RedisCacheService) GetCachedUser(ctx context.Context, userID uuid.UUID) (*models.User, error) {
	key := fmt.Sprintf("user:%s", userID.String())
	result, err := r.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil // Cache miss
	}

	var user models.User
	if err := json.Unmarshal([]byte(result.(string)), &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return &user, nil
}

func (r *RedisCacheService) InvalidateUserCache(ctx context.Context, userID uuid.UUID) error {
	key := fmt.Sprintf("user:%s", userID.String())
	return r.Delete(ctx, key)
}

// Transaction-specific caching

func (r *RedisCacheService) CacheTransaction(ctx context.Context, transaction *models.Transaction, ttl int) error {
	key := fmt.Sprintf("transaction:%s", transaction.ID.String())
	return r.Set(ctx, key, transaction, ttl)
}

func (r *RedisCacheService) GetCachedTransaction(ctx context.Context, transactionID uuid.UUID) (*models.Transaction, error) {
	key := fmt.Sprintf("transaction:%s", transactionID.String())
	result, err := r.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil // Cache miss
	}

	var transaction models.Transaction
	if err := json.Unmarshal([]byte(result.(string)), &transaction); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction: %w", err)
	}

	return &transaction, nil
}

func (r *RedisCacheService) InvalidateTransactionCache(ctx context.Context, transactionID uuid.UUID) error {
	key := fmt.Sprintf("transaction:%s", transactionID.String())
	return r.Delete(ctx, key)
}

// Transaction list operations

func (r *RedisCacheService) GetUserTransactions(ctx context.Context, userID string, limit, offset int) ([]interface{}, error) {
	key := fmt.Sprintf("user_transactions:%s:%d:%d", userID, limit, offset)
	result, err := r.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil // Cache miss
	}

	var transactions []interface{}
	if err := json.Unmarshal([]byte(result.(string)), &transactions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal transactions: %w", err)
	}

	return transactions, nil
}

func (r *RedisCacheService) SetUserTransactions(ctx context.Context, userID string, transactions []interface{}, expiration time.Duration) error {
	key := fmt.Sprintf("user_transactions:%s", userID)
	jsonBytes, err := json.Marshal(transactions)
	if err != nil {
		return fmt.Errorf("failed to marshal transactions: %w", err)
	}

	return r.Set(ctx, key, string(jsonBytes), int(expiration.Seconds()))
}

func (r *RedisCacheService) InvalidateUserTransactions(ctx context.Context, userID string) error {
	// Delete all user transaction cache keys
	pattern := fmt.Sprintf("user_transactions:%s*", userID)
	keys, err := r.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get keys for pattern %s: %w", pattern, err)
	}

	if len(keys) > 0 {
		err = r.client.Del(ctx, keys...).Err()
		if err != nil {
			return fmt.Errorf("failed to delete user transaction keys: %w", err)
		}
	}

	return nil
}

// GetTransactions retrieves a list of transactions from cache
func (r *RedisCacheService) GetTransactions(ctx context.Context, cacheKey string) ([]*models.Transaction, error) {
	result, err := r.Get(ctx, cacheKey)
	if err != nil {
		r.incrementErrors()
		return nil, err
	}

	if result == nil {
		r.incrementMisses()
		return nil, nil // Cache miss
	}

	r.incrementHits()

	// Unmarshal the cached transactions
	var transactions []*models.Transaction
	if err := json.Unmarshal([]byte(result.(string)), &transactions); err != nil {
		r.incrementErrors()
		return nil, fmt.Errorf("failed to unmarshal transactions: %w", err)
	}

	return transactions, nil
}

// SetTransactions stores a list of transactions in cache
func (r *RedisCacheService) SetTransactions(ctx context.Context, cacheKey string, transactions []*models.Transaction, expiration time.Duration) error {
	data, err := json.Marshal(transactions)
	if err != nil {
		r.incrementErrors()
		return fmt.Errorf("failed to marshal transactions: %w", err)
	}

	err = r.Set(ctx, cacheKey, string(data), int(expiration.Seconds()))
	if err != nil {
		r.incrementErrors()
		return err
	}

	r.incrementSets()
	return nil
}

// Banking-specific operations

func (r *RedisCacheService) GetBalance(ctx context.Context, userID string) (float64, error) {
	key := fmt.Sprintf("balance:%s", userID)
	result, err := r.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	if result == nil {
		return 0, nil // Cache miss
	}

	balanceStr, ok := result.(string)
	if !ok {
		return 0, fmt.Errorf("invalid balance format for user %s", userID)
	}

	balance, err := strconv.ParseFloat(balanceStr, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse balance for user %s: %w", userID, err)
	}

	return balance, nil
}

func (r *RedisCacheService) SetBalance(ctx context.Context, userID string, balance float64, expiration time.Duration) error {
	key := fmt.Sprintf("balance:%s", userID)
	return r.Set(ctx, key, fmt.Sprintf("%.2f", balance), int(expiration.Seconds()))
}

func (r *RedisCacheService) InvalidateBalance(ctx context.Context, userID string) error {
	key := fmt.Sprintf("balance:%s", userID)
	return r.Delete(ctx, key)
}

func (r *RedisCacheService) GetTransaction(ctx context.Context, transactionID string) (interface{}, error) {
	key := fmt.Sprintf("transaction:%s", transactionID)
	result, err := r.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, nil // Cache miss
	}

	// Try to unmarshal as Transaction
	var transaction models.Transaction
	if err := json.Unmarshal([]byte(result.(string)), &transaction); err != nil {
		return result, nil // Return raw string if unmarshaling fails
	}

	return &transaction, nil
}

func (r *RedisCacheService) SetTransaction(ctx context.Context, transactionID string, transaction interface{}, expiration time.Duration) error {
	key := fmt.Sprintf("transaction:%s", transactionID)
	jsonBytes, err := json.Marshal(transaction)
	if err != nil {
		return fmt.Errorf("failed to marshal transaction: %w", err)
	}

	return r.Set(ctx, key, string(jsonBytes), int(expiration.Seconds()))
}

func (r *RedisCacheService) InvalidateTransaction(ctx context.Context, transactionID string) error {
	key := fmt.Sprintf("transaction:%s", transactionID)
	return r.Delete(ctx, key)
}

// Cache statistics

func (r *RedisCacheService) GetStats() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	total := r.stats.Hits + r.stats.Misses
	hitRate := 0.0
	if total > 0 {
		hitRate = float64(r.stats.Hits) / float64(total) * 100
	}

	return map[string]interface{}{
		"hits":     r.stats.Hits,
		"misses":   r.stats.Misses,
		"sets":     r.stats.Sets,
		"deletes":  r.stats.Deletes,
		"errors":   r.stats.Errors,
		"hit_rate": hitRate,
		"total":    total,
	}
}

// Helper methods for statistics

func (r *RedisCacheService) incrementHits() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stats.Hits++
}

func (r *RedisCacheService) incrementMisses() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stats.Misses++
}

func (r *RedisCacheService) incrementSets() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stats.Sets++
}

func (r *RedisCacheService) incrementDeletes() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stats.Deletes++
}

func (r *RedisCacheService) incrementErrors() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stats.Errors++
}
