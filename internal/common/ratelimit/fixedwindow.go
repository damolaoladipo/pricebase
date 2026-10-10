package ratelimit

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type fixedWindowMemoryEntry struct {
	Count       int
	WindowStart time.Time
}

type FixedWindowLimiter struct {
	Redis *redis.Client

	Memory map[string]*fixedWindowMemoryEntry
	Mutex  sync.Mutex
}

func NewFixedWindowLimiter(redisClient *redis.Client) *FixedWindowLimiter {

	return &FixedWindowLimiter{
		Redis: redisClient,
		Memory: make(
			map[string]*fixedWindowMemoryEntry,
		),
	}
}

var fixedWindowScript = redis.NewScript(`
local count = redis.call("INCR", KEYS[1])

if count == 1 then
    redis.call("PEXPIRE", KEYS[1], ARGV[1])
end

return count
`)

func (l *FixedWindowLimiter) CheckRateLimit(ctx context.Context, key string, window time.Duration, max int) (bool, error) {

	// Check Redis for rate limit
	if l.Redis != nil {

		count, err := fixedWindowScript.Run(ctx, l.Redis, []string{key}, window.Milliseconds()).Int64()

		if err == nil {
			return count <= int64(max), nil
		}
	}

	// Check Memory for rate limit
	now := time.Now()

	l.Mutex.Lock()
	defer l.Mutex.Unlock()

	entry, exists := l.Memory[key]

	if !exists {
		l.Memory[key] = &fixedWindowMemoryEntry{Count: 1, WindowStart: now}
		return true, nil
	}

	if now.Sub(entry.WindowStart) >= window {

		entry.Count = 1
		entry.WindowStart = now

		return true, nil
	}

	entry.Count++

	return entry.Count <= max, nil
}
