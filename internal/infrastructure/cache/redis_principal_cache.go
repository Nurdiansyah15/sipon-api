package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"sipon-api/internal/app/service/principal"
)

type RedisPrincipalCache struct {
	client *redis.Client
}

func NewRedisPrincipalCache(client *redis.Client) *RedisPrincipalCache {
	return &RedisPrincipalCache{client: client}
}

func (c *RedisPrincipalCache) key(userID string) string {
	return fmt.Sprintf("principal:%s", userID)
}

func (c *RedisPrincipalCache) Get(ctx context.Context, userID string) (*principal.Principal, error) {
	raw, err := c.client.Get(ctx, c.key(userID)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, err
	}

	var p principal.Principal
	if err := json.Unmarshal(raw, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *RedisPrincipalCache) Set(ctx context.Context, userID string, p *principal.Principal, ttl time.Duration) error {
	raw, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, c.key(userID), raw, ttl).Err()
}

func (c *RedisPrincipalCache) Delete(ctx context.Context, userID string) error {
	return c.client.Del(ctx, c.key(userID)).Err()
}
