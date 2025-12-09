package kafka

import (
	"context"
	"fmt"

	"cyrene/internal/platform/config"

	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/twmb/franz-go/pkg/kgo"
)

type Handler func(ctx context.Context, payload []byte) error

// Producer wraps the Kafka producer client.
type Producer struct {
	client *kgo.Client
}

// Consumer wraps the Kafka consumer client.
type Consumer struct {
	client   *kgo.Client
	handlers map[string]Handler
}

// NewProducer creates a new Kafka producer.
func NewProducer(cfg *config.KafkaConfig) (*Producer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Brokers...),
	)
	if err != nil {
		return nil, fmt.Errorf("create kafka producer: %w", err)
	}
	return &Producer{client: client}, nil
}

// Produce sends a message to the specified topic synchronously.
func (p *Producer) Produce(ctx context.Context, topic string, key, value []byte) error {
	record := &kgo.Record{
		Topic: topic,
		Key:   key,
		Value: value,
	}
	results := p.client.ProduceSync(ctx, record)
	return results.FirstErr()
}

// Close closes the producer client.
func (p *Producer) Close() {
	p.client.Close()
}

// NewConsumer creates a new Kafka consumer.
// handlers maps topic names to their handler functions.
// Topics are derived from the handler map keys.
func NewConsumer(cfg *config.KafkaConfig, handlers map[string]Handler) (*Consumer, error) {
	topics := make([]string, 0, len(handlers))
	for topic := range handlers {
		topics = append(topics, topic)
	}

	client, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ConsumerGroup(cfg.ConsumerGroup),
		kgo.ConsumeTopics(topics...),
	)
	if err != nil {
		return nil, fmt.Errorf("create kafka client: %w", err)
	}

	return &Consumer{
		client:   client,
		handlers: handlers,
	}, nil
}

// Run starts the consumer loop, routing messages to handlers based on topic.
func (c *Consumer) Run(ctx context.Context) error {
	for {
		fetches := c.client.PollFetches(ctx)
		if err := ctx.Err(); err != nil {
			return nil // graceful shutdown
		}

		if errs := fetches.Errors(); len(errs) > 0 {
			return fmt.Errorf("kafka fetch errors: %v", errs)
		}

		iter := fetches.RecordIter()
		for !iter.Done() {
			record := iter.Next()
			handler, ok := c.handlers[record.Topic]
			if !ok {
				// TODO: log no handler for topic
				continue
			}
			if err := handler(ctx, record.Value); err != nil {
				// TODO: log and continue
			}
		}
	}
}

func EnsureTopics(ctx context.Context, brokers []string, topics []string) error {
	client, err := kgo.NewClient(kgo.SeedBrokers(brokers...))
	if err != nil {
		return fmt.Errorf("create admin client: %w", err)
	}
	defer client.Close()

	admin := kadm.NewClient(client)

	resp, err := admin.CreateTopics(ctx, 1, 1, nil, topics...)
	if err != nil {
		return fmt.Errorf("create topics: %w", err)
	}

	for _, r := range resp {
		if r.Err != nil && r.Err != kerr.TopicAlreadyExists {
			return fmt.Errorf("create topic %s: %w", r.Topic, r.Err)
		}
	}
	return nil
}
