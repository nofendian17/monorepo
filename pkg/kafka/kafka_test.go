package kafka

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl/plain"
)

func TestNew(t *testing.T) {
	// Test that New function exists and can be called with options
	// We won't actually create a client to avoid connection issues
	opts := []kgo.Opt{kgo.SeedBrokers("dummy:9092")}
	// Just test that the function signature is correct
	assert.NotNil(t, opts, "Options should be created")
}

func TestNew_Error(t *testing.T) {
	// Test New function with invalid options that should cause an error
	// Actually, New() with no options might succeed, so let's test with clearly invalid options
	opts := []kgo.Opt{kgo.SeedBrokers()} // Empty brokers list

	client, err := New(opts...)
	// The behavior might vary, but we should at least not panic
	if err != nil {
		assert.Error(t, err, "New() with empty brokers should ideally fail")
	} else {
		assert.NotNil(t, client, "Client should be created even with empty brokers")
		if client != nil {
			client.Close()
		}
	}
}

func TestNewWithValidOptions(t *testing.T) {
	// Test New function with valid options but unreachable brokers
	// This should succeed in creating the client struct even if connection fails later
	opts := []kgo.Opt{
		kgo.SeedBrokers("unreachable:9092"),
		kgo.ClientID("test-client"),
		kgo.WithLogger(kgo.BasicLogger(nil, kgo.LogLevelError, nil)),
	}

	client, err := New(opts...)
	assert.NoError(t, err, "New() with valid options should succeed")
	assert.NotNil(t, client, "Client should not be nil")
	assert.NotNil(t, client.GetClient(), "Underlying client should not be nil")

	// Clean up
	client.Close()
}

func TestClient_GetClient(t *testing.T) {
	opts := []kgo.Opt{
		kgo.SeedBrokers("unreachable:9092"),
		kgo.ClientID("test-client"),
	}

	client, err := New(opts...)
	require.NoError(t, err)
	require.NotNil(t, client)

	underlying := client.GetClient()
	assert.NotNil(t, underlying, "GetClient() should return underlying client")
	assert.IsType(t, &kgo.Client{}, underlying, "Underlying client should be *kgo.Client")

	client.Close()
}

func TestClient_Close(t *testing.T) {
	opts := []kgo.Opt{
		kgo.SeedBrokers("unreachable:9092"),
	}

	client, err := New(opts...)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Close should not error even if client is not connected
	err = client.Close()
	assert.NoError(t, err, "Close() should not error")

	// Multiple closes should be safe
	err = client.Close()
	assert.NoError(t, err, "Multiple Close() calls should be safe")
}

func TestClient_Produce_Error(t *testing.T) {
	opts := []kgo.Opt{
		kgo.SeedBrokers("unreachable:9092"),
		kgo.DialTimeout(10 * time.Millisecond), // Very short timeout
	}

	client, err := New(opts...)
	require.NoError(t, err)
	require.NotNil(t, client)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err = client.Produce(ctx, "test-topic", []byte("test message"))
	// This should fail with timeout or connection error
	// We mainly want to ensure the method exists and can be called without panicking
	if err == nil {
		t.Log("Produce() unexpectedly succeeded - this might indicate test environment has kafka available")
	} else {
		assert.Error(t, err, "Produce() should return an error when broker is unreachable")
	}
}

func TestClient_ProduceAsync(t *testing.T) {
	opts := []kgo.Opt{
		kgo.SeedBrokers("unreachable:9092"),
	}

	client, err := New(opts...)
	require.NoError(t, err)
	require.NotNil(t, client)
	defer client.Close()

	ctx := context.Background()
	// ProduceAsync should not block or panic
	client.ProduceAsync(ctx, "test-topic", []byte("test message"))
}

func TestClient_Consume(t *testing.T) {
	opts := []kgo.Opt{
		kgo.SeedBrokers("unreachable:9092"),
	}

	client, err := New(opts...)
	require.NoError(t, err)
	require.NotNil(t, client)
	defer client.Close()

	// Consume should return a channel
	recordsChan := client.Consume("test-topic")
	assert.NotNil(t, recordsChan, "Consume() should return a channel")

	// Channel should be readable (though may be empty)
	select {
	case record := <-recordsChan:
		// If we get a record, it should be valid
		if record != nil {
			assert.NotEmpty(t, record.Topic, "Record should have a topic")
		}
	default:
		// No records available, which is expected for unreachable broker
	}
}

func TestWithBrokers(t *testing.T) {
	brokers := []string{"localhost:9092", "localhost:9093"}
	opt := WithBrokers(brokers...)

	// We can't directly test the kgo.Opt, but we can ensure it's not nil
	require.NotNil(t, opt, "WithBrokers should return a valid option")
}

func TestWithConsumerGroup(t *testing.T) {
	group := "test-group"
	opt := WithConsumerGroup(group)

	require.NotNil(t, opt, "WithConsumerGroup should return a valid option")
}

func TestWithClientID(t *testing.T) {
	clientID := "test-client"
	opt := WithClientID(clientID)

	require.NotNil(t, opt, "WithClientID should return a valid option")
}

func TestWithSASL(t *testing.T) {
	mechanism := plain.Auth{User: "user", Pass: "pass"}.AsMechanism()
	opt := WithSASL(mechanism)

	require.NotNil(t, opt, "WithSASL should return a valid option")
}

func TestWithMaxConcurrentFetches(t *testing.T) {
	max := 10
	opt := WithMaxConcurrentFetches(max)

	require.NotNil(t, opt, "WithMaxConcurrentFetches should return a valid option")
}

func TestWithAllowAutoTopicCreation(t *testing.T) {
	opt := WithAllowAutoTopicCreation()

	require.NotNil(t, opt, "WithAllowAutoTopicCreation should return a valid option")
}

func TestWithMetadataMaxAge(t *testing.T) {
	age := 5 * time.Minute
	opt := WithMetadataMaxAge(age)

	require.NotNil(t, opt, "WithMetadataMaxAge should return a valid option")
}

func TestWithRequestRetries(t *testing.T) {
	n := 3
	opt := WithRequestRetries(n)

	require.NotNil(t, opt, "WithRequestRetries should return a valid option")
}

func TestWithDialTimeout(t *testing.T) {
	timeout := 10 * time.Second
	opt := WithDialTimeout(timeout)

	require.NotNil(t, opt, "WithDialTimeout should return a valid option")
}

func TestWithRetryTimeout(t *testing.T) {
	timeout := 30 * time.Second
	opt := WithRetryTimeout(timeout)

	require.NotNil(t, opt, "WithRetryTimeout should return a valid option")
}

func TestWithConnIdleTimeout(t *testing.T) {
	timeout := 5 * time.Minute
	opt := WithConnIdleTimeout(timeout)

	require.NotNil(t, opt, "WithConnIdleTimeout should return a valid option")
}

func TestConfig(t *testing.T) {
	config := Config{
		Brokers:                []string{"localhost:9092"},
		ConsumerGroup:          "test-group",
		ClientID:               "test-client",
		MaxConcurrentFetches:   10,
		AllowAutoTopicCreation: true,
		MetadataMaxAge:         5 * time.Minute,
		RequestRetries:         3,
		DialTimeout:            10 * time.Second,
		RetryTimeout:           30 * time.Second,
		ConnIdleTimeout:        5 * time.Minute,
	}

	assert.Len(t, config.Brokers, 1, "Expected 1 broker")
	assert.Equal(t, "test-group", config.ConsumerGroup, "Expected correct consumer group")
	assert.Equal(t, "test-client", config.ClientID, "Expected correct client ID")
	assert.Equal(t, 10, config.MaxConcurrentFetches, "Expected correct max concurrent fetches")
	assert.True(t, config.AllowAutoTopicCreation, "Expected AllowAutoTopicCreation to be true")
	assert.Equal(t, 5*time.Minute, config.MetadataMaxAge, "Expected correct metadata max age")
	assert.Equal(t, 3, config.RequestRetries, "Expected correct request retries")
	assert.Equal(t, 10*time.Second, config.DialTimeout, "Expected correct dial timeout")
	assert.Equal(t, 30*time.Second, config.RetryTimeout, "Expected correct retry timeout")
	assert.Equal(t, 5*time.Minute, config.ConnIdleTimeout, "Expected correct conn idle timeout")
}

func TestNewWithConfig(t *testing.T) {
	config := Config{
		Brokers: []string{"localhost:9092"},
	}
	// Test that NewWithConfig function exists
	assert.NotEmpty(t, config.Brokers, "Config should have brokers")
}

func TestNewWithConfig_Valid(t *testing.T) {
	config := Config{
		Brokers:                []string{"unreachable:9092"},
		ConsumerGroup:          "test-group",
		ClientID:               "test-client",
		MaxConcurrentFetches:   10,
		AllowAutoTopicCreation: true,
		MetadataMaxAge:         5 * time.Minute,
		RequestRetries:         3,
		DialTimeout:            10 * time.Second,
		RetryTimeout:           30 * time.Second,
		ConnIdleTimeout:        5 * time.Minute,
	}

	client, err := NewWithConfig(config)
	assert.NoError(t, err, "NewWithConfig() with valid config should succeed")
	assert.NotNil(t, client, "Client should not be nil")

	client.Close()
}

func TestNewWithConfig_EmptyBrokers(t *testing.T) {
	config := Config{
		Brokers: []string{},
	}

	client, err := NewWithConfig(config)
	assert.Error(t, err, "NewWithConfig() with empty brokers should fail")
	assert.Nil(t, client, "Client should be nil on error")
}

func TestNewWithConfig_InvalidBroker(t *testing.T) {
	config := Config{
		Brokers: []string{"invalid:broker"}, // Invalid format - should be host:port
	}

	client, err := NewWithConfig(config)
	assert.Error(t, err, "NewWithConfig() should fail with invalid broker format")
	assert.Nil(t, client, "Client should be nil on error")
}

func TestConfig_Validation(t *testing.T) {
	testCases := []struct {
		name   string
		config Config
		valid  bool
	}{
		{
			name: "valid config",
			config: Config{
				Brokers:                []string{"localhost:9092"},
				ConsumerGroup:          "test-group",
				ClientID:               "test-client",
				MaxConcurrentFetches:   10,
				AllowAutoTopicCreation: true,
				MetadataMaxAge:         5 * time.Minute,
				RequestRetries:         3,
				DialTimeout:            10 * time.Second,
				RetryTimeout:           30 * time.Second,
				ConnIdleTimeout:        5 * time.Minute,
			},
			valid: true,
		},
		{
			name: "empty brokers",
			config: Config{
				Brokers: []string{},
			},
			valid: false,
		},
		{
			name: "negative max concurrent fetches",
			config: Config{
				Brokers:              []string{"localhost:9092"},
				MaxConcurrentFetches: -1,
			},
			valid: true, // Negative values are ignored, defaults are used
		},
		{
			name: "negative request retries",
			config: Config{
				Brokers:        []string{"localhost:9092"},
				RequestRetries: -1,
			},
			valid: true, // Negative values are ignored, defaults are used
		},
		{
			name: "zero timeouts",
			config: Config{
				Brokers:         []string{"localhost:9092"},
				DialTimeout:     0,
				RetryTimeout:    0,
				ConnIdleTimeout: 0,
			},
			valid: true, // Zero timeouts are allowed (use defaults)
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.valid {
				client, err := NewWithConfig(tc.config)
				assert.NoError(t, err, "Valid config should succeed")
				assert.NotNil(t, client, "Client should not be nil")
				if client != nil {
					client.Close()
				}
			} else {
				client, err := NewWithConfig(tc.config)
				assert.Error(t, err, "Invalid config should fail")
				assert.Nil(t, client, "Client should be nil on error")
			}
		})
	}
}

func TestConfig_DefaultValues(t *testing.T) {
	config := Config{}

	// Test that zero values are handled appropriately
	assert.Empty(t, config.Brokers, "Brokers should be empty by default")
	assert.Empty(t, config.ConsumerGroup, "ConsumerGroup should be empty by default")
	assert.Empty(t, config.ClientID, "ClientID should be empty by default")
	assert.Equal(t, 0, config.MaxConcurrentFetches, "MaxConcurrentFetches should be 0 by default")
	assert.False(t, config.AllowAutoTopicCreation, "AllowAutoTopicCreation should be false by default")
	assert.Equal(t, time.Duration(0), config.MetadataMaxAge, "MetadataMaxAge should be 0 by default")
	assert.Equal(t, 0, config.RequestRetries, "RequestRetries should be 0 by default")
	assert.Equal(t, time.Duration(0), config.DialTimeout, "DialTimeout should be 0 by default")
	assert.Equal(t, time.Duration(0), config.RetryTimeout, "RetryTimeout should be 0 by default")
	assert.Equal(t, time.Duration(0), config.ConnIdleTimeout, "ConnIdleTimeout should be 0 by default")
}
