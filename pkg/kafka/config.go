package kafka

import (
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/sasl"
)

// Config holds Kafka configuration
type Config struct {
	Brokers                []string
	ConsumerGroup          string
	ClientID               string
	MaxConcurrentFetches   int
	AllowAutoTopicCreation bool
	MetadataMaxAge         time.Duration
	RequestRetries         int
	DialTimeout            time.Duration
	RetryTimeout           time.Duration
	ConnIdleTimeout        time.Duration
	SASLMechanism          sasl.Mechanism
}

// NewWithConfig creates a new Kafka client from a config struct
func NewWithConfig(config Config) (KafkaClient, error) {
	opts := []kgo.Opt{
		WithBrokers(config.Brokers...),
	}

	if config.ConsumerGroup != "" {
		opts = append(opts, WithConsumerGroup(config.ConsumerGroup))
	}

	if config.ClientID != "" {
		opts = append(opts, WithClientID(config.ClientID))
	}

	if config.MaxConcurrentFetches > 0 {
		opts = append(opts, WithMaxConcurrentFetches(config.MaxConcurrentFetches))
	}

	if config.AllowAutoTopicCreation {
		opts = append(opts, WithAllowAutoTopicCreation())
	}

	if config.MetadataMaxAge > 0 {
		opts = append(opts, WithMetadataMaxAge(config.MetadataMaxAge))
	}

	if config.RequestRetries > 0 {
		opts = append(opts, WithRequestRetries(config.RequestRetries))
	}

	if config.DialTimeout > 0 {
		opts = append(opts, WithDialTimeout(config.DialTimeout))
	}

	if config.RetryTimeout > 0 {
		opts = append(opts, WithRetryTimeout(config.RetryTimeout))
	}

	if config.ConnIdleTimeout > 0 {
		opts = append(opts, WithConnIdleTimeout(config.ConnIdleTimeout))
	}

	if config.SASLMechanism != nil {
		opts = append(opts, WithSASL(config.SASLMechanism))
	}

	return New(opts...)
}
