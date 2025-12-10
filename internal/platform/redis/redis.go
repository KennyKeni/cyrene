package redis

import (
	"context"
	"cyrene/internal/platform/config"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// RedisClient wraps the Redis SDK connection.
//
// To implement:
// 1. go get github.com/redis/go-redis/v9
// 2. Create client: redis.NewClient(&redis.Options{...})
type Client struct {
	Client *redis.Client
}

func New(cfg *config.RedisConfig) (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("redis ping: %w", err)
	}

	return &Client{Client: client}, nil
}
