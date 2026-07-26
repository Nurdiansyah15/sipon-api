package cache

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"sipon-api/internal/app/port"
)

type RedisRateLimiter struct {
	client *redis.Client
}

func NewRedisRateLimiter(client *redis.Client) *RedisRateLimiter {
	return &RedisRateLimiter{client: client}
}

// Allow implements sliding window counter approximation.
// Uses two INCR keys (current + previous bucket) in one pipeline roundtrip.
// Fail-open: returns Allowed=true on Redis error (caller logs the error).
func (r *RedisRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (port.RateLimitResult, error) {
	now := time.Now()
	windowSecs := int64(window.Seconds())
	if windowSecs <= 0 {
		windowSecs = 60
	}

	currentBucket := now.Unix() / windowSecs
	prevBucket := currentBucket - 1

	currKey := fmt.Sprintf("%s:%d", key, currentBucket)
	prevKey := fmt.Sprintf("%s:%d", key, prevBucket)

	pipe := r.client.Pipeline()
	incrCmd := pipe.Incr(ctx, currKey)
	getCmd := pipe.Get(ctx, prevKey)
	pipe.Expire(ctx, currKey, window*2)
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		// fail-open
		return port.RateLimitResult{Allowed: true, Remaining: limit}, err
	}

	currCount, _ := incrCmd.Result()

	var prevCount int64
	if v := getCmd.Val(); v != "" {
		prevCount, _ = strconv.ParseInt(v, 10, 64)
	}

	// Sliding window approximation
	elapsed := float64(now.Unix()%windowSecs) / float64(windowSecs)
	rate := float64(prevCount)*(1-elapsed) + float64(currCount)

	resetAt := time.Unix((currentBucket+1)*windowSecs, 0)
	remaining := int(math.Max(0, float64(limit)-math.Ceil(rate)))

	if rate > float64(limit) {
		return port.RateLimitResult{Allowed: false, Remaining: 0, ResetAt: resetAt}, nil
	}

	return port.RateLimitResult{Allowed: true, Remaining: remaining, ResetAt: resetAt}, nil
}
