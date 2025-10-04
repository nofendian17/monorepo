package kafka

import (
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl"
)

// WithBrokers sets the Kafka brokers
func WithBrokers(brokers ...string) kgo.Opt {
	return kgo.SeedBrokers(brokers...)
}

// WithConsumerGroup sets the consumer group for the client
func WithConsumerGroup(group string) kgo.Opt {
	return kgo.ConsumerGroup(group)
}

// WithClientID sets the client ID for the Kafka client
func WithClientID(clientID string) kgo.Opt {
	return kgo.ClientID(clientID)
}

// WithSASL sets SASL authentication
func WithSASL(mechanism sasl.Mechanism) kgo.Opt {
	return kgo.SASL(mechanism)
}

// WithMaxConcurrentFetches sets the maximum number of concurrent fetches
func WithMaxConcurrentFetches(max int) kgo.Opt {
	return kgo.MaxConcurrentFetches(max)
}

// WithAllowAutoTopicCreation enables automatic topic creation
func WithAllowAutoTopicCreation() kgo.Opt {
	return kgo.AllowAutoTopicCreation()
}

// WithMetadataMaxAge sets the maximum age of metadata before refresh
func WithMetadataMaxAge(age time.Duration) kgo.Opt {
	return kgo.MetadataMaxAge(age)
}

// WithRequestRetries sets the number of request retries
func WithRequestRetries(n int) kgo.Opt {
	return kgo.RequestRetries(n)
}

// WithDialTimeout sets the dial timeout
func WithDialTimeout(timeout time.Duration) kgo.Opt {
	return kgo.DialTimeout(timeout)
}

// WithRetryTimeout sets the retry timeout
func WithRetryTimeout(timeout time.Duration) kgo.Opt {
	return kgo.RetryTimeout(timeout)
}

// WithConnIdleTimeout sets the connection idle timeout
func WithConnIdleTimeout(timeout time.Duration) kgo.Opt {
	return kgo.ConnIdleTimeout(timeout)
}
