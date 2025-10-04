package redis

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithAddrs(t *testing.T) {
	client := &Client{
		opts: &redis.UniversalOptions{},
	}
	addrs := []string{"localhost:6379", "localhost:6380"}

	opt := WithAddrs(addrs)
	opt(client)

	assert.Len(t, client.opts.Addrs, 2, "Expected 2 addresses")
	assert.Equal(t, "localhost:6379", client.opts.Addrs[0], "Expected first addr 'localhost:6379'")
}

func TestWithUsername(t *testing.T) {
	client := &Client{
		opts: &redis.UniversalOptions{},
	}
	username := "testuser"

	opt := WithUsername(username)
	opt(client)

	assert.Equal(t, username, client.opts.Username, "Expected correct username")
}

func TestWithPassword(t *testing.T) {
	client := &Client{
		opts: &redis.UniversalOptions{},
	}
	password := "testpass"

	opt := WithPassword(password)
	opt(client)

	assert.Equal(t, password, client.opts.Password, "Expected correct password")
}

func TestWithDB(t *testing.T) {
	client := &Client{
		opts: &redis.UniversalOptions{},
	}
	db := 5

	opt := WithDB(db)
	opt(client)

	assert.Equal(t, db, client.opts.DB, "Expected correct DB")
}

func TestWithDialTimeout(t *testing.T) {
	client := &Client{
		opts: &redis.UniversalOptions{},
	}
	timeout := 10 * time.Second

	opt := WithDialTimeout(timeout)
	opt(client)

	assert.Equal(t, timeout, client.opts.DialTimeout, "Expected correct dial timeout")
}

func TestWithReadTimeout(t *testing.T) {
	client := &Client{
		opts: &redis.UniversalOptions{},
	}
	timeout := 5 * time.Second

	opt := WithReadTimeout(timeout)
	opt(client)

	assert.Equal(t, timeout, client.opts.ReadTimeout, "Expected correct read timeout")
}

func TestWithWriteTimeout(t *testing.T) {
	client := &Client{
		opts: &redis.UniversalOptions{},
	}
	timeout := 5 * time.Second

	opt := WithWriteTimeout(timeout)
	opt(client)

	assert.Equal(t, timeout, client.opts.WriteTimeout, "Expected correct write timeout")
}

func TestWithPoolSize(t *testing.T) {
	client := &Client{
		opts: &redis.UniversalOptions{},
	}
	poolSize := 20

	opt := WithPoolSize(poolSize)
	opt(client)

	assert.Equal(t, poolSize, client.opts.PoolSize, "Expected correct pool size")
}

func TestConfig(t *testing.T) {
	config := Config{
		Addrs:    []string{"localhost:6379"},
		Username: "user",
		Password: "pass",
		DB:       1,
	}

	assert.Len(t, config.Addrs, 1, "Expected 1 addr")
	assert.Equal(t, "user", config.Username, "Expected correct username")
	assert.Equal(t, 1, config.DB, "Expected correct DB")
	assert.Equal(t, "pass", config.Password, "Expected correct password")
}

func setupMockRedis() (RedisClient, redismock.ClientMock) {
	db, mock := redismock.NewClientMock()
	client := &Client{
		opts:   &redis.UniversalOptions{},
		client: db,
	}
	return client, mock
}

func TestClient_Set_Get(t *testing.T) {
	client, mock := setupMockRedis()
	ctx := context.Background()

	key := "test_key"
	value := "test_value"

	mock.ExpectSet(key, value, 0).SetVal("OK")
	mock.ExpectGet(key).SetVal(value)

	err := client.Set(ctx, key, value, 0)
	require.NoError(t, err, "Set() should not fail")

	result, err := client.Get(ctx, key)
	require.NoError(t, err, "Get() should not fail")

	assert.Equal(t, value, result, "Expected correct value")

	require.NoError(t, mock.ExpectationsWereMet(), "Redis expectations should be met")
}

func TestClient_Del(t *testing.T) {
	client, mock := setupMockRedis()
	ctx := context.Background()

	key := "test_key"
	value := "test_value"

	mock.ExpectSet(key, value, 0).SetVal("OK")
	mock.ExpectDel(key).SetVal(1)

	err := client.Set(ctx, key, value, 0)
	require.NoError(t, err, "Set() should not fail")

	err = client.Del(ctx, key)
	require.NoError(t, err, "Del() should not fail")

	require.NoError(t, mock.ExpectationsWereMet(), "Redis expectations should be met")
}

func TestClient_Exists(t *testing.T) {
	client, mock := setupMockRedis()
	ctx := context.Background()

	key := "test_key"

	// Test key doesn't exist
	mock.ExpectExists(key).SetVal(0)
	exists, err := client.Exists(ctx, key)
	require.NoError(t, err, "Exists() should not fail")
	assert.False(t, exists, "Key should not exist initially")

	// Test key exists
	mock.ExpectExists(key).SetVal(1)
	exists, err = client.Exists(ctx, key)
	require.NoError(t, err, "Exists() should not fail")
	assert.True(t, exists, "Key should exist")

	require.NoError(t, mock.ExpectationsWereMet(), "Redis expectations should be met")
}

func TestClient_Expire_TTL(t *testing.T) {
	client, mock := setupMockRedis()
	ctx := context.Background()

	key := "test_key"
	value := "test_value"
	expiration := 2 * time.Second

	mock.ExpectSet(key, value, expiration).SetVal("OK")
	mock.ExpectExpire(key, 5*time.Second).SetVal(true)
	mock.ExpectTTL(key).SetVal(5 * time.Second)

	err := client.Set(ctx, key, value, expiration)
	require.NoError(t, err, "Set() should not fail")

	err = client.Expire(ctx, key, 5*time.Second)
	require.NoError(t, err, "Expire() should not fail")

	ttl, err := client.TTL(ctx, key)
	require.NoError(t, err, "TTL() should not fail")
	assert.True(t, ttl > 0, "TTL should be positive")

	require.NoError(t, mock.ExpectationsWereMet(), "Redis expectations should be met")
}

func TestClient_HSet_HGet(t *testing.T) {
	client, mock := setupMockRedis()
	ctx := context.Background()

	key := "test_hash"
	field := "test_field"
	value := "test_value"

	mock.ExpectHSet(key, field, value).SetVal(1)
	mock.ExpectHGet(key, field).SetVal(value)

	err := client.HSet(ctx, key, field, value)
	require.NoError(t, err, "HSet() should not fail")

	result, err := client.HGet(ctx, key, field)
	require.NoError(t, err, "HGet() should not fail")

	assert.Equal(t, value, result, "Expected correct value")

	require.NoError(t, mock.ExpectationsWereMet(), "Redis expectations should be met")
}

func TestClient_HMSet_HMGet(t *testing.T) {
	client, mock := setupMockRedis()
	ctx := context.Background()

	key := "test_hash"
	fields := map[string]interface{}{
		"field1": "value1",
		"field2": "value2",
	}

	mock.ExpectHMSet(key, fields).SetVal(true)
	mock.ExpectHMGet(key, "field1", "field2").SetVal([]interface{}{"value1", "value2"})

	err := client.HMSet(ctx, key, fields)
	require.NoError(t, err, "HMSet() should not fail")

	results, err := client.HMGet(ctx, key, "field1", "field2")
	require.NoError(t, err, "HMGet() should not fail")

	assert.Len(t, results, 2, "Expected 2 results")
	assert.Equal(t, "value1", results[0], "Expected correct value1")
	assert.Equal(t, "value2", results[1], "Expected correct value2")

	require.NoError(t, mock.ExpectationsWereMet(), "Redis expectations should be met")
}

func TestClient_SAdd_SMembers(t *testing.T) {
	client, mock := setupMockRedis()
	ctx := context.Background()

	key := "test_set"
	members := []interface{}{"member1", "member2", "member3"}

	mock.ExpectSAdd(key, members...).SetVal(3)
	mock.ExpectSMembers(key).SetVal([]string{"member1", "member2", "member3"})

	err := client.SAdd(ctx, key, members...)
	require.NoError(t, err, "SAdd() should not fail")

	results, err := client.SMembers(ctx, key)
	require.NoError(t, err, "SMembers() should not fail")

	assert.Len(t, results, 3, "Expected 3 members")

	// Check if all members are present
	memberMap := make(map[string]bool)
	for _, member := range results {
		memberMap[member] = true
	}

	for _, expected := range []string{"member1", "member2", "member3"} {
		assert.True(t, memberMap[expected], "Expected member %s not found in results", expected)
	}

	require.NoError(t, mock.ExpectationsWereMet(), "Redis expectations should be met")
}

func TestClient_LPush_RPop(t *testing.T) {
	client, mock := setupMockRedis()
	ctx := context.Background()

	key := "test_list"
	values := []interface{}{"value1", "value2", "value3"}

	mock.ExpectLPush(key, values...).SetVal(3)
	mock.ExpectRPop(key).SetVal("value1")

	err := client.LPush(ctx, key, values...)
	require.NoError(t, err, "LPush() should not fail")

	result, err := client.RPop(ctx, key)
	require.NoError(t, err, "RPop() should not fail")

	assert.Equal(t, "value1", result, "Expected correct value")

	require.NoError(t, mock.ExpectationsWereMet(), "Redis expectations should be met")
}

func TestNew(t *testing.T) {
	client, err := New(WithAddrs([]string{"localhost:6379"}))
	require.NoError(t, err, "New() should not fail")
	require.NotNil(t, client, "New() should return a client")
}

func TestNewWithConfig(t *testing.T) {
	config := Config{
		Addrs: []string{"localhost:6379"},
	}
	client, err := NewWithConfig(config)
	require.NoError(t, err, "NewWithConfig() should not fail")
	require.NotNil(t, client, "NewWithConfig() should return a client")
}

func TestClient_Getters(t *testing.T) {
	client, _ := setupMockRedis()

	// Test getter methods
	assert.NotNil(t, client.GetClient(), "GetClient() should return client")
	assert.Equal(t, 0, client.DB(), "DB() should return default DB")
	assert.Equal(t, 0, client.PoolSize(), "PoolSize() should return pool size")
}
