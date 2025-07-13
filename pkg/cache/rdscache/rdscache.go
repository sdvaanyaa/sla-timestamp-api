package rdscache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/sdvaanyaa/sla-timestamp-api/pkg/cache"
	"log/slog"
	"time"
)

type Client struct {
	client *redis.Client
	log    *slog.Logger
}

func New(client *redis.Client, log *slog.Logger) (*Client, error) {
	if log == nil {
		log = slog.Default()
	}

	if err := client.Ping(context.Background()).Err(); err != nil {
		log.Error("redis ping failed", slog.Any("error", err))
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	log.Info("redis connection established")

	return &Client{
		client: client,
		log:    log,
	}, nil
}

func (c *Client) Get(ctx context.Context, key string, dest any) error {
	start := time.Now()
	val, err := c.client.Get(ctx, key).Result()
	duration := time.Since(start)

	c.logOp("GET", key, duration, err)

	if errors.Is(err, redis.Nil) {
		return cache.ErrCacheMiss
	}

	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(val), dest)
}

func (c *Client) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	start := time.Now()
	err = c.client.Set(ctx, key, data, ttl).Err()
	duration := time.Since(start)

	c.logOp("SET", key, duration, err)

	return err
}

func (c *Client) Delete(ctx context.Context, key string) error {
	start := time.Now()
	err := c.client.Del(ctx, key).Err()
	duration := time.Since(start)

	c.logOp("DEL", key, duration, err)

	return err
}

func (c *Client) logOp(op, key string, duration time.Duration, err error) {
	attrs := []slog.Attr{
		slog.String("operation", op),
		slog.String("key", key),
		slog.Duration("duration", duration),
	}
	if err != nil {
		attrs = append(attrs, slog.Any("error", err))
		c.log.LogAttrs(context.Background(), slog.LevelError, "cache operation failed", attrs...)

		return
	}
	c.log.LogAttrs(context.Background(), slog.LevelDebug, "cache operation succeeded", attrs...)
}
