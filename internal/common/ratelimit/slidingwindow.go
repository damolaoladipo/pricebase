// Sliding Window Counter
package ratelimit

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type slidingWindowMemoryEntry struct {
	CurrentStart time.Time
	CurrentCount int64

	PreviousStart time.Time
	PreviousCount int64
}

type SlidingWindowLimiter struct {
	Redis *redis.Client

	Memory map[string]*slidingWindowMemoryEntry
	Mutex  sync.Mutex
}

func NewSlidingWindowLimiter(redisClient *redis.Client) *SlidingWindowLimiter {

	return &SlidingWindowLimiter{
		Redis: redisClient,
		Memory: make(
			map[string]*slidingWindowMemoryEntry,
		),
	}
}

var slidingWindowScript = redis.NewScript(`
local current_key = KEYS[1]
local previous_key = KEYS[2]

local window_ms = tonumber(ARGV[1])
local now_ms = tonumber(ARGV[2])
local max_requests = tonumber(ARGV[3])

local current_count = redis.call("INCR", current_key)

if current_count == 1 then
    redis.call("PEXPIRE", current_key, window_ms * 2)
end

local previous_count = redis.call("GET", previous_key)

if not previous_count then
    previous_count = 0
else
    previous_count = tonumber(previous_count)
end

local current_elapsed = now_ms % window_ms
local previous_weight = (window_ms - current_elapsed) / window_ms

local estimated_count =
    (previous_count * previous_weight) +
    current_count

return estimated_count
`)

func (l *SlidingWindowLimiter) CheckRateLimit(ctx context.Context, key string, window time.Duration, max int) (bool, error) {

	// Check Redis for rate limit
	now := time.Now()

	windowStart := now.Truncate(window)

	currentKey := fmt.Sprintf("%s:%d", key, windowStart.UnixMilli())

	previousKey := fmt.Sprintf("%s:%d", key, windowStart.Add(-window).UnixMilli())

	if l.Redis != nil {

		result, err := slidingWindowScript.Run(
			ctx,
			l.Redis,
			[]string{currentKey, previousKey},
			window.Milliseconds(),
			now.UnixMilli(),
			max,
		).Float64()

		if err == nil {
			return result <= float64(max), nil
		}
	}

	// Check Memory for rate limit

	l.Mutex.Lock()
	defer l.Mutex.Unlock()

	windowStart = now.Truncate(window)

	entry, exists := l.Memory[key]

	if !exists {

		l.Memory[key] = &slidingWindowMemoryEntry{CurrentStart: windowStart, CurrentCount: 1}
		return true, nil
	}

	if windowStart.After(entry.CurrentStart) {

		entry.PreviousStart = entry.CurrentStart
		entry.PreviousCount = entry.CurrentCount

		entry.CurrentStart = windowStart
		entry.CurrentCount = 1

	} else {

		entry.CurrentCount++
	}

	elapsed := now.Sub(entry.CurrentStart)
	previousWeight := float64(window-elapsed) / float64(window)

	if previousWeight < 0 {
		previousWeight = 0
	}
	estimatedCount := float64(entry.PreviousCount)*previousWeight + float64(entry.CurrentCount)

	return estimatedCount <= float64(max), nil

}
