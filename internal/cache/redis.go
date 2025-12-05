package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	RedisClient *redis.Client
	ctx         = context.Background()
)

// InitRedis initializes Redis connection if REDIS_URL is set
func InitRedis() error {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		fmt.Println("Redis not configured, using database for webhook tests")
		return nil
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	RedisClient = redis.NewClient(opt)

	// Test connection
	if err := RedisClient.Ping(ctx).Err(); err != nil {
		RedisClient = nil
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	fmt.Println("Redis connected successfully")
	return nil
}

// IsRedisAvailable checks if Redis is available
func IsRedisAvailable() bool {
	return RedisClient != nil
}

// WebhookTestData represents webhook test data for Redis
type WebhookTestData struct {
	Token     string    `json:"token"`
	UserID    uint      `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

// WebhookRequestData represents webhook request data for Redis
type WebhookRequestData struct {
	ID          string    `json:"id"`
	Method      string    `json:"method"`
	URL         string    `json:"url"`
	Headers     string    `json:"headers"`
	Body        string    `json:"body"`
	QueryParams string    `json:"query_params"`
	ClientIP    string    `json:"client_ip"`
	CreatedAt   time.Time `json:"created_at"`
}

// SaveWebhookTest saves webhook test to Redis
func SaveWebhookTest(token string, data WebhookTestData) error {
	if !IsRedisAvailable() {
		return fmt.Errorf("Redis not available")
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("webhook_test:%s", token)
	ttl := time.Until(data.ExpiresAt)
	
	return RedisClient.Set(ctx, key, jsonData, ttl).Err()
}

// GetWebhookTest retrieves webhook test from Redis
func GetWebhookTest(token string) (*WebhookTestData, error) {
	if !IsRedisAvailable() {
		return nil, fmt.Errorf("Redis not available")
	}

	key := fmt.Sprintf("webhook_test:%s", token)
	val, err := RedisClient.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var data WebhookTestData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil, err
	}

	return &data, nil
}

// SaveWebhookRequest saves webhook request to Redis
func SaveWebhookRequest(token string, request WebhookRequestData) error {
	if !IsRedisAvailable() {
		return fmt.Errorf("Redis not available")
	}

	jsonData, err := json.Marshal(request)
	if err != nil {
		return err
	}

	// Add to list (LPUSH for newest first)
	listKey := fmt.Sprintf("webhook_requests:%s", token)
	if err := RedisClient.LPush(ctx, listKey, jsonData).Err(); err != nil {
		return err
	}

	// Trim to keep only last 100 requests
	if err := RedisClient.LTrim(ctx, listKey, 0, 99).Err(); err != nil {
		return err
	}

	// Set expiration (24 hours)
	return RedisClient.Expire(ctx, listKey, 24*time.Hour).Err()
}

// GetWebhookRequests retrieves webhook requests from Redis
func GetWebhookRequests(token string, limit int) ([]WebhookRequestData, error) {
	if !IsRedisAvailable() {
		return nil, fmt.Errorf("Redis not available")
	}

	if limit <= 0 {
		limit = 100
	}

	listKey := fmt.Sprintf("webhook_requests:%s", token)
	vals, err := RedisClient.LRange(ctx, listKey, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}

	requests := make([]WebhookRequestData, 0, len(vals))
	for _, val := range vals {
		var request WebhookRequestData
		if err := json.Unmarshal([]byte(val), &request); err != nil {
			continue
		}
		requests = append(requests, request)
	}

	return requests, nil
}

// DeleteWebhookTest deletes webhook test and its requests from Redis
func DeleteWebhookTest(token string) error {
	if !IsRedisAvailable() {
		return fmt.Errorf("Redis not available")
	}

	testKey := fmt.Sprintf("webhook_test:%s", token)
	requestsKey := fmt.Sprintf("webhook_requests:%s", token)

	pipe := RedisClient.Pipeline()
	pipe.Del(ctx, testKey)
	pipe.Del(ctx, requestsKey)
	_, err := pipe.Exec(ctx)

	return err
}

// GetUserWebhookTests retrieves all webhook tests for a user from Redis
func GetUserWebhookTests(userID uint) ([]WebhookTestData, error) {
	if !IsRedisAvailable() {
		return nil, fmt.Errorf("Redis not available")
	}

	// Scan for all webhook_test:* keys
	var cursor uint64
	var webhooks []WebhookTestData

	for {
		keys, nextCursor, err := RedisClient.Scan(ctx, cursor, "webhook_test:*", 100).Result()
		if err != nil {
			return nil, err
		}

		for _, key := range keys {
			val, err := RedisClient.Get(ctx, key).Result()
			if err != nil {
				continue
			}

			var data WebhookTestData
			if err := json.Unmarshal([]byte(val), &data); err != nil {
				continue
			}

			if data.UserID == userID {
				webhooks = append(webhooks, data)
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return webhooks, nil
}

// CleanupExpiredWebhooks is not needed for Redis (TTL handles it automatically)
func CleanupExpiredWebhooks() {
	// Redis automatically removes expired keys
	fmt.Println("Redis: Automatic TTL cleanup (no action needed)")
}
