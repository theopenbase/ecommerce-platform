package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type Client = redis.Client

func NewClient(opts *redis.Options) *Client {
	return redis.NewClient(opts)
}

func Ping(ctx context.Context, c *Client) error {
	return c.Ping(ctx).Err()
}
