package middleware

import (
	"log/slog"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"sipon-api/internal/app/apperror"
	"sipon-api/internal/app/port"
	"sipon-api/internal/interfaces/http/httperror"
)

// RateLimitByIP limits requests per client IP.
// Passes through (fail-open) when limiter is nil or on Redis error.
func RateLimitByIP(limiter port.RateLimiter, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if limiter == nil {
			c.Next()
			return
		}

		ip := c.ClientIP()
		key := "rate_limit:ip:" + ip

		result, err := limiter.Allow(c.Request.Context(), key, limit, window)
		if err != nil {
			slog.Warn("rate limiter error", "key", key, "err", err)
			c.Next()
			return
		}

		setRateLimitHeaders(c, limit, result)

		if !result.Allowed {
			setRetryAfterHeader(c, result.ResetAt)
			httperror.Handle(c, apperror.TooManyRequests(string(apperror.CodeTooManyRequests)))
			return
		}

		c.Next()
	}
}

// RateLimitByUser limits requests per authenticated user ID.
// Skips if user_id is not set in context (unauthenticated requests).
// Passes through (fail-open) when limiter is nil or on Redis error.
func RateLimitByUser(limiter port.RateLimiter, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if limiter == nil {
			c.Next()
			return
		}

		userID := c.GetString("user_id")
		if userID == "" {
			c.Next()
			return
		}

		key := "rate_limit:user:" + userID

		result, err := limiter.Allow(c.Request.Context(), key, limit, window)
		if err != nil {
			slog.Warn("rate limiter error", "key", key, "err", err)
			c.Next()
			return
		}

		setRateLimitHeaders(c, limit, result)

		if !result.Allowed {
			setRetryAfterHeader(c, result.ResetAt)
			httperror.Handle(c, apperror.TooManyRequests(string(apperror.CodeTooManyRequests)))
			return
		}

		c.Next()
	}
}

func setRateLimitHeaders(c *gin.Context, limit int, result port.RateLimitResult) {
	c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
	c.Header("X-RateLimit-Remaining", strconv.Itoa(result.Remaining))
	c.Header("X-RateLimit-Reset", strconv.FormatInt(result.ResetAt.Unix(), 10))
}

func setRetryAfterHeader(c *gin.Context, resetAt time.Time) {
	secs := int64(time.Until(resetAt).Seconds()) + 1
	if secs < 1 {
		secs = 1
	}
	c.Header("Retry-After", strconv.FormatInt(secs, 10))
}
