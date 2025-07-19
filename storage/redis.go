package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"sync/atomic"
	"time"
	"tiny-url-service/models"
	"tiny-url-service/utils"

	"github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	client  *redis.Client
	baseURL string
	ctx     context.Context
	counter uint64 // Local counter, synced with Redis
}

func NewRedisStorage(baseURL, redisURL string) (*RedisStorage, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opts)
	ctx := context.Background()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	storage := &RedisStorage{
		client:  client,
		baseURL: baseURL,
		ctx:     ctx,
	}

	// Initialize counter from Redis
	if err := storage.initCounter(); err != nil {
		return nil, fmt.Errorf("failed to initialize counter: %w", err)
	}

	return storage, nil
}

func (r *RedisStorage) initCounter() error {
	// Get current counter value from Redis, or start at 0
	val, err := r.client.Get(r.ctx, "counter").Uint64()
	if err == redis.Nil {
		// Counter doesn't exist, start at 0
		atomic.StoreUint64(&r.counter, 0)
		return nil
	}
	if err != nil {
		return err
	}
	atomic.StoreUint64(&r.counter, val)
	return nil
}

// Store saves a URL mapping and returns the generated short code
func (r *RedisStorage) Store(mapping *models.URLMapping) (string, error) {
	// Generate unique ID using Redis INCR for atomicity across instances
	id, err := r.client.Incr(r.ctx, "counter").Result()
	if err != nil {
		return "", fmt.Errorf("failed to generate ID: %w", err)
	}

	// Generate short code using base62 encoding
	shortCode := utils.EncodeBase62(uint64(id))

	// Complete the mapping
	mapping.ID = uint64(id)
	mapping.ShortCode = shortCode
	mapping.CreatedAt = time.Now()

	// Serialize mapping to JSON
	data, err := json.Marshal(mapping)
	if err != nil {
		return "", fmt.Errorf("failed to marshal URL mapping: %w", err)
	}

	// Store in Redis
	if err := r.client.Set(r.ctx, "url:"+shortCode, data, 0).Err(); err != nil {
		return "", fmt.Errorf("failed to store URL mapping in Redis: %w", err)
	}

	// Update local counter
	atomic.StoreUint64(&r.counter, uint64(id))

	return shortCode, nil
}

// Get retrieves the URL mapping for a given short code
func (r *RedisStorage) Get(shortCode string) (*models.URLMapping, error) {
	data, err := r.client.Get(r.ctx, "url:"+shortCode).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("short code not found: %s", shortCode)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get URL mapping from Redis: %w", err)
	}

	var mapping models.URLMapping
	if err := json.Unmarshal([]byte(data), &mapping); err != nil {
		return nil, fmt.Errorf("failed to unmarshal URL mapping: %w", err)
	}

	// Check if expired
	if r.IsExpired(&mapping) {
		return nil, fmt.Errorf("URL has expired: %s", shortCode)
	}

	return &mapping, nil
}

// IsExpired checks if a URL mapping has expired
func (r *RedisStorage) IsExpired(mapping *models.URLMapping) bool {
	if mapping.ExpirationDate == nil {
		return false // No expiration set
	}
	return time.Now().After(*mapping.ExpirationDate)
}

// GetStats returns storage statistics
func (r *RedisStorage) GetStats() map[string]interface{} {
	// Get current counter
	currentCounter := atomic.LoadUint64(&r.counter)

	// Count total URLs (this is expensive for large datasets)
	totalUrls, err := r.client.Eval(r.ctx, `
		local keys = redis.call('KEYS', 'url:*')
		return #keys
	`, []string{}).Result()

	if err != nil {
		totalUrls = 0
	}

	return map[string]interface{}{
		"total_urls":      totalUrls,
		"current_counter": currentCounter,
		"storage_type":    "redis",
	}
}

// Close closes the Redis connection
func (r *RedisStorage) Close() error {
	return r.client.Close()
} 