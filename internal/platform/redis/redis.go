package redis

import (
	"cyrene/internal/platform/config"
)

// RedisClient wraps the Redis SDK connection.
//
// To implement:
// 1. go get github.com/redis/go-redis/v9
// 2. Create client: redis.NewClient(&redis.Options{...})
type RedisClient struct {
	// client *redis.Client
}

// New creates a new Redis client.
//
// Example implementation with go-redis:
//
//	import goredis "github.com/redis/go-redis/v9"
//
//	func New(cfg *config.RedisConfig) *RedisClient {
//	    client := goredis.NewClient(&goredis.Options{
//	        Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
//	        Password: cfg.Password,
//	        DB:       cfg.DB,
//	    })
//	    return &RedisClient{client: client}
//	}
func New(cfg *config.RedisConfig) *RedisClient {
	panic("redis: not implemented - see comments for implementation guide")
}
