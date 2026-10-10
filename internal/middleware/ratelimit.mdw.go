package middleware

import (
	"net/http"
	"time"

	"github.com/damolaoladipo/pricebase/internal/common/ratelimit"
	r "github.com/damolaoladipo/pricebase/internal/common/redis"
	"github.com/damolaoladipo/pricebase/internal/utils"
	"github.com/gin-gonic/gin"
)

var fixedWindowLimiter *ratelimit.FixedWindowLimiter
var slidingWindowLimiter *ratelimit.SlidingWindowLimiter
var tokenBucketLimiter *ratelimit.TokenBucketLimiter

func RateLimiters() {

	REDIS := r.REDIS.Client()

	fixedWindowLimiter = ratelimit.NewFixedWindowLimiter(REDIS)
	slidingWindowLimiter = ratelimit.NewSlidingWindowLimiter(REDIS)
	tokenBucketLimiter = ratelimit.NewTokenBucketLimiter(REDIS)
}

func limitHandler(c *gin.Context, message string) {

	c.AbortWithStatusJSON(http.StatusTooManyRequests,
		gin.H{
			"error":   true,
			"errors":  []string{"too many requests"},
			"message": message,
			"status":  429,
			"data":    nil,
		},
	)
}

func FixedWindowRateLimiter(config ratelimit.Config) gin.HandlerFunc {

	return func(c *gin.Context) {

		key := config.KeyGenerator(c)

		check, err := fixedWindowLimiter.CheckRateLimit(c.Request.Context(), config.Prefix+":"+key, config.Window, config.Max)

		if err != nil {
			c.Next()
			return
		}

		if !check {
			limitHandler(c, "rate limit exceeded")
			return
		}

		c.Next()
	}
}

func SlidingWindowRateLimiter(config ratelimit.Config) gin.HandlerFunc {

	return func(c *gin.Context) {

		key := config.KeyGenerator(c)

		check, err := slidingWindowLimiter.CheckRateLimit(c.Request.Context(), config.Prefix+":"+key, config.Window, config.Max)

		if err != nil {

			c.Next()

			return
		}

		if !check {

			limitHandler(c, "rate limit exceeded")

			return
		}

		c.Next()
	}
}

func TokenBucketRateLimiter(config ratelimit.TokenBucketConfig) gin.HandlerFunc {

	return func(c *gin.Context) {

		key := config.KeyGenerator(c)

		check, err := tokenBucketLimiter.CheckRateLimit(c.Request.Context(), config.Prefix+":"+key, config.Rate, config.Burst)

		if err != nil {

			c.Next()

			return
		}

		if !check {

			limitHandler(c, "rate limit exceeded")

			return
		}

		c.Next()
	}
}

func RateLimiter(env *utils.ENV) gin.HandlerFunc {

	window := ratelimit.ParseWindow(env.RL_GLOBAL_WINDOW_MS, 24*time.Hour)
	max := ratelimit.ParseMax(env.RL_GLOBAL_MAX, 30000)

	return FixedWindowRateLimiter(
		ratelimit.Config{
			Prefix:       "rate-limit:global",
			Window:       window, // 24 hours
			Max:          max,    // 30000 requests
			KeyGenerator: ratelimit.BuildAuthRateLimitKey,
		})
}

func AuthRateLimiter(env *utils.ENV) gin.HandlerFunc {

	window := ratelimit.ParseWindow(env.RL_AUTH_WINDOW_MS, 15*time.Minute)
	max := ratelimit.ParseMax(env.RL_AUTH_MAX, 100)

	return FixedWindowRateLimiter(
		ratelimit.Config{
			Prefix:       "rate-limit:auth",
			Window:       window, // 15 minutes
			Max:          max,    // 100 requests
			KeyGenerator: ratelimit.BuildAuthRateLimitKey,
		},
	)
}

func LoginRateLimiter(env *utils.ENV) gin.HandlerFunc {

	window := ratelimit.ParseWindow(env.RL_LOGIN_WINDOW_MS, 15*time.Minute)
	max := ratelimit.ParseMax(env.RL_LOGIN_MAX, 10)

	return FixedWindowRateLimiter(
		ratelimit.Config{
			Prefix:       "rate-limit:login",
			Window:       window, // 15 minutes
			Max:          max,    // 10 requests
			KeyGenerator: ratelimit.BuildLoginRateLimitKey,
		},
	)
}

func ProtectedRateLimiter(env *utils.ENV) gin.HandlerFunc {

	window := ratelimit.ParseWindow(env.RL_PROTECTED_WINDOW_MS, 15*time.Minute)
	max := ratelimit.ParseMax(env.RL_PROTECTED_MAX, 300)

	return FixedWindowRateLimiter(
		ratelimit.Config{
			Prefix:       "rate-limit:protected",
			Window:       window, // 15 minutes
			Max:          max,    // 300 requests
			KeyGenerator: ratelimit.ProtectedRateLimitKey,
		},
	)
}

func LimitRequests() gin.HandlerFunc {

	return FixedWindowRateLimiter(
		ratelimit.Config{
			Prefix:       "rate-limit:basic",
			Window:       24 * time.Hour, // 24 hours
			Max:          30000,          // 30000 requests
			KeyGenerator: ratelimit.ReadIp,
		},
	)
}

func FeedRateLimiter() gin.HandlerFunc {

	return SlidingWindowRateLimiter(
		ratelimit.Config{
			Prefix:       "rate-limit:feed",
			Window:       time.Minute, // 1 minute
			Max:          120,         // 120 requests
			KeyGenerator: ratelimit.ProtectedRateLimitKey,
		},
	)
}

func SearchRateLimiter() gin.HandlerFunc {

	return SlidingWindowRateLimiter(
		ratelimit.Config{
			Prefix:       "rate-limit:search",
			Window:       time.Minute, // 1 minute
			Max:          100,         // 100 requests
			KeyGenerator: ratelimit.ProtectedRateLimitKey,
		},
	)
}

func PlaybackRateLimiter() gin.HandlerFunc {

	return TokenBucketRateLimiter(
		ratelimit.TokenBucketConfig{
			Prefix:       "rate-limit:playback",
			Rate:         5,  // 5 requests per second
			Burst:        10, // 10 requests per burst
			KeyGenerator: ratelimit.ProtectedRateLimitKey,
		},
	)
}

// https://github.com/AliRizaAynaci/gorl
