package port

import (
	"context"
	"time"
)

type RateLimitResult struct {
	Allowed   bool
	Remaining int
	ResetAt   time.Time
}

type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (RateLimitResult, error)
}
