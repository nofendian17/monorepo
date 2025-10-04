package jwt

import (
	"context"
	"fmt"
	"time"

	"monorepo/pkg/redis"
)

// RedisStore implements RefreshTokenStore interface using the existing pkg/redis package
type RedisStore struct {
	client redis.RedisClient
	ctx    context.Context
}

// NewRedisStore creates a new Redis store using the existing pkg/redis client
func NewRedisStore(redisClient redis.RedisClient) *RedisStore {
	return &RedisStore{
		client: redisClient,
		ctx:    context.Background(),
	}
}

// Save stores a refresh token with its expiry time in Redis
func (s *RedisStore) Save(userID, tokenID, token string, expiry time.Time) error {
	key := fmt.Sprintf("refresh_token:%s:%s", userID, tokenID)

	// Calculate the duration until expiry
	duration := time.Until(expiry)

	// Store the token with expiry using the existing Redis client
	err := s.client.Set(s.ctx, key, token, duration)
	if err != nil {
		return fmt.Errorf("failed to save refresh token to Redis: %w", err)
	}

	return nil
}

// Get retrieves a stored refresh token from Redis
func (s *RedisStore) Get(userID, tokenID string) (string, error) {
	key := fmt.Sprintf("refresh_token:%s:%s", userID, tokenID)

	token, err := s.client.Get(s.ctx, key)
	if err != nil {
		return "", fmt.Errorf("refresh token not found for user %s, token ID %s: %w", userID, tokenID, err)
	}

	return token, nil
}

// Delete removes a refresh token from Redis storage
func (s *RedisStore) Delete(userID, tokenID string) error {
	key := fmt.Sprintf("refresh_token:%s:%s", userID, tokenID)

	err := s.client.Del(s.ctx, key)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token from Redis: %w", err)
	}

	return nil
}

// DeleteAll removes all refresh tokens for a user from Redis
func (s *RedisStore) DeleteAll(userID string) error {
	// Find all keys matching the pattern for this user
	pattern := fmt.Sprintf("refresh_token:%s:*", userID)

	// Get all matching keys using the underlying Redis client
	underlyingClient := s.client.GetClient()
	keys, err := underlyingClient.Keys(s.ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to find refresh tokens for user %s: %w", userID, err)
	}

	if len(keys) == 0 {
		return nil // No tokens to delete
	}

	// Delete all matching keys using the underlying client
	deleted, err := s.client.GetClient().Del(s.ctx, keys...).Result()
	if err != nil {
		return fmt.Errorf("failed to delete refresh tokens for user %s: %w", userID, err)
	}

	if deleted == 0 {
		return fmt.Errorf("no refresh tokens were deleted for user %s", userID)
	}

	return nil
}

// Cleanup removes expired tokens from Redis (this is handled automatically by Redis TTL)
func (s *RedisStore) Cleanup() error {
	// Redis automatically removes keys with expired TTLs
	// This method exists to satisfy the interface
	return nil
}

// Close closes the Redis connection
func (s *RedisStore) Close() error {
	return s.client.Close()
}
