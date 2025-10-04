package kafka

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"
)

// KafkaClient defines the interface for Kafka operations
type KafkaClient interface {
	Produce(ctx context.Context, topic string, value []byte) error
	ProduceAsync(ctx context.Context, topic string, value []byte)
	Consume(topics ...string) <-chan *kgo.Record
	Close() error
	GetClient() *kgo.Client
}

// Client represents a Kafka client wrapper that handles both producing and consuming
type Client struct {
	opts   []kgo.Opt
	client *kgo.Client
}

// New creates a new Kafka client with the provided options
func New(opts ...kgo.Opt) (KafkaClient, error) {
	// Create the actual Kafka client with the provided options
	kafkaClient, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	client := &Client{
		opts:   opts,
		client: kafkaClient,
	}

	return client, nil
}

// Produce sends a message to a Kafka topic
func (k *Client) Produce(ctx context.Context, topic string, value []byte) error {
	record := &kgo.Record{
		Topic: topic,
		Value: value,
	}

	return k.client.ProduceSync(ctx, record).FirstErr()
}

// ProduceAsync sends a message to a Kafka topic asynchronously
func (k *Client) ProduceAsync(ctx context.Context, topic string, value []byte) {
	record := &kgo.Record{
		Topic: topic,
		Value: value,
	}

	k.client.Produce(ctx, record, func(record *kgo.Record, err error) {
		if err != nil {
			// In a real application, you might want to log this error
			// or handle it according to your error handling strategy
		}
	})
}

// Consume starts consuming messages from the specified topics
// It returns a channel that will receive Kafka records
func (k *Client) Consume(topics ...string) <-chan *kgo.Record {
	// Add consume topics to the client
	k.client.AddConsumeTopics(topics...)

	recordsChan := make(chan *kgo.Record, 100) // buffered channel
	go func() {
		defer close(recordsChan)
		for {
			// Poll for fetches
			fetches := k.client.PollFetches(context.Background())
			if fetches.IsClientClosed() {
				return
			}

			iter := fetches.RecordIter()
			for !iter.Done() {
				record := iter.Next()
				select {
				case recordsChan <- record:
				case <-context.Background().Done():
					return
				}
			}
		}
	}()

	return recordsChan
}

// Close closes the Kafka client
func (k *Client) Close() error {
	if k.client != nil {
		k.client.Close()
	}
	return nil
}

// GetClient returns the underlying Kafka client for advanced operations
func (k *Client) GetClient() *kgo.Client {
	return k.client
}
