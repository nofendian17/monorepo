package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient defines the interface for Redis operations
type RedisClient interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)
	HSet(ctx context.Context, key string, field string, value any) error
	HGet(ctx context.Context, key string, field string) (string, error)
	HMSet(ctx context.Context, key string, fields map[string]interface{}) error
	HMGet(ctx context.Context, key string, fields ...string) ([]interface{}, error)
	SAdd(ctx context.Context, key string, members ...interface{}) error
	SMembers(ctx context.Context, key string) ([]string, error)
	LPush(ctx context.Context, key string, values ...interface{}) error
	RPop(ctx context.Context, key string) (string, error)
	Close() error
	GetClient() redis.UniversalClient
	Addrs() []string
	Username() string
	DB() int
	DialTimeout() time.Duration
	ReadTimeout() time.Duration
	WriteTimeout() time.Duration
	PoolSize() int
}

// Option is a function that configures a Client
type Option func(*Client)

// Client represents a Redis client wrapper
type Client struct {
	opts   *redis.UniversalOptions
	client redis.UniversalClient
}

// New creates a new Redis client with the provided options
func New(opts ...Option) (RedisClient, error) {
	client := &Client{
		opts: &redis.UniversalOptions{
			Addrs:        []string{"localhost:6379"},
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			PoolSize:     10,
		},
	}

	// Apply options
	for _, opt := range opts {
		opt(client)
	}

	// Create the actual Redis client with the configured options
	client.client = redis.NewUniversalClient(client.opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}

// NewWithConfig creates a new Redis client from a config struct
func NewWithConfig(config Config) (RedisClient, error) {
	opts := []Option{
		WithAddrs(config.Addrs),
		WithUsername(config.Username),
		WithPassword(config.Password),
		WithDB(config.DB),
		WithDialTimeout(config.DialTimeout),
		WithReadTimeout(config.ReadTimeout),
		WithWriteTimeout(config.WriteTimeout),
		WithPoolSize(config.PoolSize),
	}

	return New(opts...)
}

// Set sets a key-value pair with expiration
func (r *Client) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.client.Set(ctx, key, value, expiration).Err()
}

// Get gets a value by key
func (r *Client) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Del deletes a key
func (r *Client) Del(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Exists checks if a key exists
func (r *Client) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Expire sets expiration for a key
func (r *Client) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.client.Expire(ctx, key, expiration).Err()
}

// TTL returns time to live for a key
func (r *Client) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

// HSet sets a hash field to value
func (r *Client) HSet(ctx context.Context, key string, field string, value any) error {
	return r.client.HSet(ctx, key, field, value).Err()
}

// HGet gets a hash field value
func (r *Client) HGet(ctx context.Context, key string, field string) (string, error) {
	return r.client.HGet(ctx, key, field).Result()
}

// HMSet sets multiple hash fields to values
func (r *Client) HMSet(ctx context.Context, key string, fields map[string]interface{}) error {
	return r.client.HMSet(ctx, key, fields).Err()
}

// HMGet gets multiple hash field values
func (r *Client) HMGet(ctx context.Context, key string, fields ...string) ([]interface{}, error) {
	return r.client.HMGet(ctx, key, fields...).Result()
}

// SAdd adds members to a set
func (r *Client) SAdd(ctx context.Context, key string, members ...interface{}) error {
	return r.client.SAdd(ctx, key, members...).Err()
}

// SMembers returns all members of a set
func (r *Client) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.client.SMembers(ctx, key).Result()
}

// LPush prepends values to a list
func (r *Client) LPush(ctx context.Context, key string, values ...interface{}) error {
	return r.client.LPush(ctx, key, values...).Err()
}

// RPop removes and returns the last element of a list
func (r *Client) RPop(ctx context.Context, key string) (string, error) {
	return r.client.RPop(ctx, key).Result()
}

// Close closes the Redis client
func (r *Client) Close() error {
	return r.client.Close()
}

// GetClient returns the underlying Redis client for advanced operations
func (r *Client) GetClient() redis.UniversalClient {
	return r.client
}

// Addrs returns the Redis server addresses
func (r *Client) Addrs() []string {
	return r.opts.Addrs
}

// Username returns the Redis username
func (r *Client) Username() string {
	return r.opts.Username
}

// DB returns the Redis database number
func (r *Client) DB() int {
	return r.opts.DB
}

// DialTimeout returns the dial timeout setting
func (r *Client) DialTimeout() time.Duration {
	return r.opts.DialTimeout
}

// ReadTimeout returns the read timeout setting
func (r *Client) ReadTimeout() time.Duration {
	return r.opts.ReadTimeout
}

// WriteTimeout returns the write timeout setting
func (r *Client) WriteTimeout() time.Duration {
	return r.opts.WriteTimeout
}

// PoolSize returns the connection pool size
func (r *Client) PoolSize() int {
	return r.opts.PoolSize
}
