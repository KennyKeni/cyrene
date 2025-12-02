package redis

import (
	"context"
	"time"

	"cyrene/internal/config"
)

// Service defines the Redis client interface.
//
// To implement:
// 1. go get github.com/redis/go-redis/v9
// 2. Create client: redis.NewClient(&redis.Options{...})
// 3. Implement methods below
type Service interface {
	// Health returns a map of health status information.
	Health() map[string]string

	// Get retrieves a value by key.
	Get(ctx context.Context, key string) (string, error)

	// Set stores a value with an optional TTL.
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Del removes one or more keys.
	Del(ctx context.Context, keys ...string) error

	// Close terminates the Redis connection.
	Close() error
}

// New creates a new Redis service.
//
// Example implementation with go-redis:
//
//	import "github.com/redis/go-redis/v9"
//
//	type service struct {
//	    client *redis.Client
//	}
//
//	func New(cfg *config.RedisConfig) Service {
//	    client := redis.NewClient(&redis.Options{
//	        Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
//	        Password: cfg.Password,
//	        DB:       cfg.DB,
//	    })
//	    return &service{client: client}
//	}
func New(cfg *config.RedisConfig) Service {
	panic("redis: not implemented - see comments for implementation guide")
}
