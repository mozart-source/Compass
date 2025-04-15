package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// RateLimiter defines an interface for rate limiting functionality
type RateLimiter interface {
	// Allow checks if the request should be allowed based on the key
	Allow(ctx context.Context, key string) (bool, int, time.Time, error)
	// Reset resets the counter for a specific key
	Reset(ctx context.Context, key string) error
	// WithLimit creates a new rate limiter with the specified limit
	WithLimit(maxAttempts int64, window time.Duration) RateLimiter
}

// RedisRateLimiter implements rate limiting using Redis
type RedisRateLimiter struct {
	client      *redis.Client
	prefix      string
	window      time.Duration
	maxAttempts int64
}

// NewRedisRateLimiter creates a new rate limiter using Redis
func NewRedisRateLimiter(client *redis.Client, window time.Duration, maxAttempts int64) *RedisRateLimiter {
	return &RedisRateLimiter{
		client:      client,
		prefix:      "ratelimit:",
		window:      window,
		maxAttempts: maxAttempts,
	}
}

// WithLimit creates a new rate limiter with the specified limit
func (rl *RedisRateLimiter) WithLimit(maxAttempts int64, window time.Duration) RateLimiter {
	return &RedisRateLimiter{
		client:      rl.client,
		prefix:      rl.prefix,
		window:      window,
		maxAttempts: maxAttempts,
	}
}

// Allow checks if the request should be allowed based on the key
func (rl *RedisRateLimiter) Allow(ctx context.Context, key string) (bool, int, time.Time, error) {
	redisKey := fmt.Sprintf("%s%s", rl.prefix, key)
	now := time.Now()
	windowStart := now.Truncate(rl.window)

	pipe := rl.client.Pipeline()
	incr := pipe.Incr(ctx, redisKey)
	pipe.ExpireAt(ctx, redisKey, windowStart.Add(rl.window))

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, time.Time{}, fmt.Errorf("rate limiter error: %w", err)
	}

	count := incr.Val()
	remaining := rl.maxAttempts - count
	if remaining < 0 {
		remaining = 0
	}

	resetTime := windowStart.Add(rl.window)
	allowed := count <= rl.maxAttempts

	return allowed, int(remaining), resetTime, nil
}

// Reset resets the counter for a specific key
func (rl *RedisRateLimiter) Reset(ctx context.Context, key string) error {
	redisKey := fmt.Sprintf("%s%s", rl.prefix, key)
	return rl.client.Del(ctx, redisKey).Err()
}

// GetWindow returns the rate limit window duration
func (rl *RedisRateLimiter) GetWindow() time.Duration {
	return rl.window
}

// GetMaxAttempts returns the maximum number of attempts allowed
func (rl *RedisRateLimiter) GetMaxAttempts() int64 {
	return rl.maxAttempts
}
