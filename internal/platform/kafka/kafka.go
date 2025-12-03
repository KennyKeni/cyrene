package kafka

import (
	"cyrene/internal/platform/config"
)

// KafkaProducer wraps the Kafka producer client.
//
// To implement:
// 1. go get github.com/twmb/franz-go/pkg/kgo
// 2. Create client: kgo.NewClient(kgo.SeedBrokers(brokers...))
type KafkaProducer struct {
	// client *kgo.Client
}

// KafkaConsumer wraps the Kafka consumer client.
//
// To implement:
// 1. go get github.com/twmb/franz-go/pkg/kgo
// 2. Create client with consumer group: kgo.NewClient(kgo.SeedBrokers(...), kgo.ConsumerGroup(...), kgo.ConsumeTopics(...))
type KafkaConsumer struct {
	// client *kgo.Client
}

// NewProducer creates a new Kafka producer.
//
// Example implementation with franz-go:
//
//	import "github.com/twmb/franz-go/pkg/kgo"
//
//	func NewProducer(cfg *config.KafkaConfig) (*KafkaProducer, error) {
//	    brokers := strings.Split(cfg.Brokers, ",")
//	    client, err := kgo.NewClient(kgo.SeedBrokers(brokers...))
//	    if err != nil {
//	        return nil, err
//	    }
//	    return &KafkaProducer{client: client}, nil
//	}
//
//	func (p *KafkaProducer) Publish(ctx context.Context, topic string, key, value []byte) error {
//	    record := &kgo.Record{Topic: topic, Key: key, Value: value}
//	    return p.client.ProduceSync(ctx, record).FirstErr()
//	}
func NewProducer(cfg *config.KafkaConfig) *KafkaProducer {
	panic("kafka: producer not implemented - see comments for implementation guide")
}

// NewConsumer creates a new Kafka consumer.
//
// Example implementation with franz-go:
//
//	import "github.com/twmb/franz-go/pkg/kgo"
//
//	func NewConsumer(cfg *config.KafkaConfig, topics []string) (*KafkaConsumer, error) {
//	    brokers := strings.Split(cfg.Brokers, ",")
//	    client, err := kgo.NewClient(
//	        kgo.SeedBrokers(brokers...),
//	        kgo.ConsumerGroup(cfg.ConsumerGroup),
//	        kgo.ConsumeTopics(topics...),
//	    )
//	    if err != nil {
//	        return nil, err
//	    }
//	    return &KafkaConsumer{client: client}, nil
//	}
//
//	func (c *KafkaConsumer) Subscribe(ctx context.Context, topics []string, handler func(key, value []byte) error) error {
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
func NewConsumer(cfg *config.KafkaConfig) *KafkaConsumer {
	panic("kafka: consumer not implemented - see comments for implementation guide")
}
