package chatstore

import (
	"context"
	platformredis "cyrene/internal/platform/redis"
	"encoding/json"
	"fmt"
	"time"

	"github.com/firebase/genkit/go/ai"
)

type ChatStore struct {
	client     *platformredis.Client
	maxMessage int
	ttl        time.Duration
}

func NewChatStore(client *platformredis.Client, maxMessage int, ttl time.Duration) *ChatStore {
	return &ChatStore{
		client:     client,
		maxMessage: maxMessage,
		ttl:        ttl,
	}
}

func (c *ChatStore) key(user string) string {
	return fmt.Sprintf("chat:history:%s", user)
}

// Technically, I should make my own message type but i cba

func (c *ChatStore) Get(ctx context.Context, user string) ([]*ai.Message, error) {
	data, err := c.client.Client.LRange(ctx, c.key(user), 0, -1).Result()
	if err != nil {
		return nil, err
	}

	messages := make([]*ai.Message, 0, len(data))
	for _, d := range data {
		var msg ai.Message
		if err := json.Unmarshal([]byte(d), &msg); err != nil {
			continue
		}
		messages = append(messages, &msg)
	}
	return messages, nil
}

func (c *ChatStore) Append(ctx context.Context, user string, messages ...*ai.Message) error {
	key := c.key(user)
	pipe := c.client.Client.Pipeline()

	for _, msg := range messages {
		data, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		pipe.RPush(ctx, key, string(data))
	}

	pipe.LTrim(ctx, key, int64(-c.maxMessage), -1)
	pipe.Expire(ctx, key, c.ttl)

	_, err := pipe.Exec(ctx)
	return err
}

func (c *ChatStore) Clear(ctx context.Context, user string) error {
	return c.client.Client.Del(ctx, c.key(user)).Err()
}
