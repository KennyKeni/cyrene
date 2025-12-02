package kafka

import (
	"context"

	"cyrene/internal/config"
)

// Producer defines the Kafka producer interface.
//
// To implement:
// 1. go get github.com/twmb/franz-go/pkg/kgo
// 2. Create client: kgo.NewClient(kgo.SeedBrokers(brokers...))
type Producer interface {
	// Publish sends a message to the specified topic.
	Publish(ctx context.Context, topic string, key, value []byte) error

	// Close flushes and closes the producer.
	Close() error
}

// Consumer defines the Kafka consumer interface.
//
// To implement:
// 1. go get github.com/twmb/franz-go/pkg/kgo
// 2. Create client with consumer group: kgo.NewClient(kgo.SeedBrokers(...), kgo.ConsumerGroup(...), kgo.ConsumeTopics(...))
type Consumer interface {
	// Subscribe starts consuming messages from the specified topics.
	// The handler is called for each message received.
	Subscribe(ctx context.Context, topics []string, handler func(key, value []byte) error) error

	// Close stops the consumer.
	Close() error
}

// Message represents a Kafka message.
type Message struct {
	Topic string
	Key   []byte
	Value []byte
}

// NewProducer creates a new Kafka producer.
//
// Example implementation with franz-go:
//
//	import "github.com/twmb/franz-go/pkg/kgo"
//
//	type producer struct {
//	    client *kgo.Client
//	}
//
//	func NewProducer(cfg *config.KafkaConfig) (Producer, error) {
//	    brokers := strings.Split(cfg.Brokers, ",")
//	    client, err := kgo.NewClient(kgo.SeedBrokers(brokers...))
//	    if err != nil {
//	        return nil, err
//	    }
//	    return &producer{client: client}, nil
//	}
//
//	func (p *producer) Publish(ctx context.Context, topic string, key, value []byte) error {
//	    record := &kgo.Record{Topic: topic, Key: key, Value: value}
//	    return p.client.ProduceSync(ctx, record).FirstErr()
//	}
func NewProducer(cfg *config.KafkaConfig) Producer {
	panic("kafka: producer not implemented - see comments for implementation guide")
}

// NewConsumer creates a new Kafka consumer.
//
// Example implementation with franz-go:
//
//	import "github.com/twmb/franz-go/pkg/kgo"
//
//	type consumer struct {
//	    client *kgo.Client
//	}
//
//	func NewConsumer(cfg *config.KafkaConfig, topics []string) (Consumer, error) {
//	    brokers := strings.Split(cfg.Brokers, ",")
//	    client, err := kgo.NewClient(
//	        kgo.SeedBrokers(brokers...),
//	        kgo.ConsumerGroup(cfg.ConsumerGroup),
//	        kgo.ConsumeTopics(topics...),
//	    )
//	    if err != nil {
//	        return nil, err
//	    }
//	    return &consumer{client: client}, nil
//	}
//
//	func (c *consumer) Subscribe(ctx context.Context, topics []string, handler func(key, value []byte) error) error {
//	    for {
//	        fetches := c.client.PollFetches(ctx)
//	        if err := ctx.Err(); err != nil {
//	            return err
//	        }
//	        fetches.EachRecord(func(r *kgo.Record) {
//	            handler(r.Key, r.Value)
//	        })
//	    }
//	}
func NewConsumer(cfg *config.KafkaConfig) Consumer {
	panic("kafka: consumer not implemented - see comments for implementation guide")
}
