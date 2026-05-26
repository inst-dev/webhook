package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/inst-dev/webhook/internal/config"
	goredis "github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

// Client wraps go-redis client
type Client struct {
	*goredis.Client
}

// NewClient creates a new Redis client
func NewClient(cfg config.RedisConfig) (*Client, error) {
	rdb := goredis.NewClient(&goredis.Options{
		Addr:         cfg.Addr(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     50,
		MinIdleConns: 10,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Info().
		Str("addr", cfg.Addr()).
		Int("db", cfg.DB).
		Msg("Redis client connected")

	return &Client{rdb}, nil
}

// Close closes the Redis client
func (c *Client) Close() error {
	return c.Client.Close()
}

// IncrCounter increments a counter with expiry
func (c *Client) IncrCounter(ctx context.Context, key string, expiry time.Duration) (int64, error) {
	pipe := c.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, expiry)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return incr.Val(), nil
}

// RateLimit checks if a key has exceeded the limit within the window
func (c *Client) RateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, int64, error) {
	pipe := c.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, err
	}

	count := incr.Val()
	return count > int64(limit), count, nil
}

// PublishEvent publishes an event to a Redis channel
func (c *Client) PublishEvent(ctx context.Context, channel string, message interface{}) error {
	return c.Publish(ctx, channel, message).Err()
}

// SubscribeChannel subscribes to a Redis channel
func (c *Client) SubscribeChannel(ctx context.Context, channels ...string) *goredis.PubSub {
	return c.Subscribe(ctx, channels...)
}
