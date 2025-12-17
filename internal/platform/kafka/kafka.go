package kafka

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

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

// ConsumerConfig holds retry and backoff settings.
type ConsumerConfig struct {
	MaxRetries  int
	BaseBackoff time.Duration
	MaxBackoff  time.Duration
}

// Consumer wraps the Kafka consumer client with retry support.
type Consumer struct {
	client         *kgo.Client
	handlers       map[string]Handler
	config         ConsumerConfig
	retryCounts    map[string]int
	pendingRetries []*kgo.Record
	currentBackoff time.Duration
	mu             sync.Mutex
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

// NewConsumer creates a new Kafka consumer with retry support.
// handlers maps topic names to their handler functions.
func NewConsumer(cfg *config.KafkaConfig, handlers map[string]Handler) (*Consumer, error) {
	topics := make([]string, 0, len(handlers))
	for topic := range handlers {
		topics = append(topics, topic)
	}

	consumerCfg := ConsumerConfig{
		MaxRetries:  cfg.MaxRetries,
		BaseBackoff: time.Duration(cfg.RetryBackoffSecs) * time.Second,
		MaxBackoff:  30 * time.Second,
	}
	if consumerCfg.MaxRetries <= 0 {
		consumerCfg.MaxRetries = 10
	}
	if consumerCfg.BaseBackoff <= 0 {
		consumerCfg.BaseBackoff = time.Second
	}

	slog.Info("creating kafka consumer",
		"brokers", cfg.Brokers,
		"group", cfg.ConsumerGroup,
		"topics", topics,
		"maxRetries", consumerCfg.MaxRetries,
		"baseBackoff", consumerCfg.BaseBackoff,
	)

	client, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ConsumerGroup(cfg.ConsumerGroup),
		kgo.ConsumeTopics(topics...),
		kgo.DisableAutoCommit(),
		kgo.BlockRebalanceOnPoll(),
	)
	if err != nil {
		return nil, fmt.Errorf("create kafka client: %w", err)
	}

	return &Consumer{
		client:      client,
		handlers:    handlers,
		config:      consumerCfg,
		retryCounts: make(map[string]int),
	}, nil
}

// Run starts the consumer loop with retry support and manual offset commits.
func (c *Consumer) Run(ctx context.Context) error {
	slog.Info("kafka consumer started")
	for {
		// First, process any pending retries
		if len(c.pendingRetries) > 0 {
			c.processPendingRetries(ctx)
			if len(c.pendingRetries) > 0 {
				// Still have failures, backoff and retry
				c.backoff(ctx)
				continue
			}
			c.resetBackoff()
		}

		// Poll for new messages
		fetches := c.client.PollFetches(ctx)
		if err := ctx.Err(); err != nil {
			slog.Info("kafka consumer shutting down")
			return nil
		}

		if errs := fetches.Errors(); len(errs) > 0 {
			slog.Error("kafka fetch errors", "errors", errs)
			return fmt.Errorf("kafka fetch errors: %v", errs)
		}

		// Process fetched records
		// Only commit contiguous prefix of successes - committing offset N marks all prior offsets consumed
		var toCommit []*kgo.Record
		var hitFailure bool

		iter := fetches.RecordIter()
		for !iter.Done() {
			record := iter.Next()

			if hitFailure {
				c.pendingRetries = append(c.pendingRetries, record)
				continue
			}

			if c.processRecord(ctx, record) {
				toCommit = append(toCommit, record)
			} else {
				hitFailure = true
				c.pendingRetries = append(c.pendingRetries, record)
			}
		}

		// Commit successful records
		if len(toCommit) > 0 {
			if err := c.client.CommitRecords(ctx, toCommit...); err != nil {
				slog.Error("commit failed", "error", err)
			}
		}
		c.client.AllowRebalance()

		// If we have pending retries, backoff before next iteration
		if len(c.pendingRetries) > 0 {
			c.backoff(ctx)
		}
	}
}

func (c *Consumer) processPendingRetries(ctx context.Context) {
	var stillPending []*kgo.Record
	var toCommit []*kgo.Record

	for _, record := range c.pendingRetries {
		if c.processRecord(ctx, record) {
			toCommit = append(toCommit, record)
		} else {
			stillPending = append(stillPending, record)
		}
	}

	c.pendingRetries = stillPending

	if len(toCommit) > 0 {
		if err := c.client.CommitRecords(ctx, toCommit...); err != nil {
			slog.Error("commit failed", "error", err)
		}
	}
}

func (c *Consumer) processRecord(ctx context.Context, record *kgo.Record) bool {
	handler, ok := c.handlers[record.Topic]
	if !ok {
		slog.Warn("no handler for topic", "topic", record.Topic)
		return true
	}

	key := c.messageKey(record)

	if err := handler(ctx, record.Value); err != nil {
		c.mu.Lock()
		c.retryCounts[key]++
		count := c.retryCounts[key]
		c.mu.Unlock()

		if count >= c.config.MaxRetries {
			slog.Warn("max retries exceeded, skipping message",
				"key", key,
				"topic", record.Topic,
				"attempts", count,
			)
			c.mu.Lock()
			delete(c.retryCounts, key)
			c.mu.Unlock()
			return true
		}

		slog.Error("handler failed, will retry",
			"key", key,
			"topic", record.Topic,
			"attempt", count,
			"error", err,
		)
		return false
	}

	c.mu.Lock()
	delete(c.retryCounts, key)
	c.mu.Unlock()
	return true
}

func (c *Consumer) messageKey(record *kgo.Record) string {
	if len(record.Key) > 0 {
		return string(record.Key)
	}
	return fmt.Sprintf("%s:%d:%d", record.Topic, record.Partition, record.Offset)
}

func (c *Consumer) backoff(ctx context.Context) {
	c.mu.Lock()
	if c.currentBackoff == 0 {
		c.currentBackoff = c.config.BaseBackoff
	}
	backoff := c.currentBackoff
	c.currentBackoff = min(c.currentBackoff*2, c.config.MaxBackoff)
	c.mu.Unlock()

	slog.Info("backing off before retry", "duration", backoff)
	select {
	case <-time.After(backoff):
	case <-ctx.Done():
	}
}

func (c *Consumer) resetBackoff() {
	c.mu.Lock()
	c.currentBackoff = 0
	c.mu.Unlock()
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
