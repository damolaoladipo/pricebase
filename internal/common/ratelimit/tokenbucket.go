package ratelimit

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type tokenBucketMemoryEntry struct {
	Tokens     float64
	LastRefill time.Time
}

type TokenBucketLimiter struct {
	Redis *redis.Client

	Memory map[string]*tokenBucketMemoryEntry
	Mutex  sync.Mutex
}

func NewTokenBucketLimiter(redisClient *redis.Client) *TokenBucketLimiter {

	return &TokenBucketLimiter{
		Redis: redisClient,
		Memory: make(
			map[string]*tokenBucketMemoryEntry,
		),
	}
}

var tokenBucketScript = redis.NewScript(`
local key = KEYS[1]

local rate = tonumber(ARGV[1])
local burst = tonumber(ARGV[2])
local now_ms = tonumber(ARGV[3])

local tokens = redis.call("HGET", key, "tokens")
local last_refill = redis.call("HGET", key, "last_refill")

if not tokens then
    tokens = burst
else
    tokens = tonumber(tokens)
end

if not last_refill then
    last_refill = now_ms
else
    last_refill = tonumber(last_refill)
end

local elapsed_ms = now_ms - last_refill

if elapsed_ms < 0 then
    elapsed_ms = 0
end

local refill = (elapsed_ms / 1000) * rate

tokens = math.min(
    burst,
    tokens + refill
)

local allowed = 0

if tokens >= 1 then
    tokens = tokens - 1
    allowed = 1
end

redis.call(
    "HSET",
    key,
    "tokens",
    tokens,
    "last_refill",
    now_ms
)

local ttl_ms = math.ceil(
    (burst / rate) * 1000
) * 2

redis.call(
    "PEXPIRE",
    key,
    ttl_ms
)

return {
    allowed,
    tokens
}
`)

func (l *TokenBucketLimiter) CheckRateLimit(ctx context.Context, key string, rate float64, burst int) (bool, error) {

	// Check burst and rate
	if burst <= 0 {
		return false, nil
	}

	if rate <= 0 {
		return false, nil
	}

	// Check Redis for rate limit
	if l.Redis != nil {

		result, err := tokenBucketScript.Run(
			ctx,
			l.Redis,
			[]string{key},
			strconv.FormatFloat(
				rate,
				'f',
				6,
				64,
			),
			burst,
			time.Now().UnixMilli(),
		).Slice()

		if err == nil && len(result) >= 2 {

			allowed, ok := result[0].(int64)

			if !ok {
				return false, fmt.Errorf(
					"invalid token bucket response",
				)
			}

			return allowed == 1, nil
		}

		if err != nil {
			// Fall through to memory limiter.
		}
	}

	// Check Memory for rate limit
	now := time.Now()

	l.Mutex.Lock()
	defer l.Mutex.Unlock()

	entry, exists := l.Memory[key]

	if !exists {

		l.Memory[key] = &tokenBucketMemoryEntry{
			Tokens:     float64(burst - 1),
			LastRefill: now,
		}

		return true, nil
	}

	elapsed := now.Sub(entry.LastRefill).Seconds()

	if elapsed > 0 {
		entry.Tokens += elapsed * rate

		if entry.Tokens > float64(burst) {
			entry.Tokens = float64(burst)
		}

		entry.LastRefill = now
	}

	if entry.Tokens < 1 {
		return false, nil
	}

	entry.Tokens--

	return true, nil
}
