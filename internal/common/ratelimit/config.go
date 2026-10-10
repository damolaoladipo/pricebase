package ratelimit

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"
	"time"

	"github.com/damolaoladipo/pricebase/internal/models"
	"github.com/gin-gonic/gin"
)

type KeyGenerator func(key string) string

type Config struct {
	Prefix       string
	Window       time.Duration
	Max          int
	KeyGenerator func(*gin.Context) string
}

type TokenBucketConfig struct {
	Prefix string
	// Tokens added per second.
	Rate float64
	// Maximum number of tokens that can be stored.
	Burst        int
	KeyGenerator func(*gin.Context) string
}

func ParseWindow(value string, fallback time.Duration) time.Duration {

	if value == "" {
		return fallback
	}

	if parsed, err := time.ParseDuration(value); err == nil {
		return parsed
	}
	if parsed, err := strconv.Atoi(value); err == nil {
		return time.Duration(parsed) * time.Millisecond
	}

	return fallback
}

func ParseMax(value string, fallback int) int {

	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	if parsed <= 0 {
		return fallback
	}

	return parsed
}

func ParseFloat(value string, fallback float64) float64 {

	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseFloat(value, 64)

	if err != nil || parsed <= 0 {
		return fallback
	}

	return parsed
}

func ReadIp(c *gin.Context) string {

	if ip := c.ClientIP(); ip != "" {
		return ip
	}
	if addr := c.Request.RemoteAddr; addr != "" {
		return addr
	}
	return "unknown"
}

func ReadIdentifier(c *gin.Context) string {

	body := map[string]any{}

	if c.Request.Body != nil {

		raw, err := io.ReadAll(c.Request.Body)
		_ = c.Request.Body.Close()
		c.Request.Body = io.NopCloser(bytes.NewReader(raw))

		if err == nil && len(raw) > 0 {
			_ = json.Unmarshal(raw, &body)
		}
	}

	for _, key := range []string{"email", "username", "phonenumber", "countryNumber", "identifier"} {
		if value, ok := body[key].(string); ok && value != "" {
			return value
		}
	}

	return "anonymous"
}

// BuildAuthRateLimitKey scopes auth attempts by channel and IP.
func BuildAuthRateLimitKey(c *gin.Context) string {

	channel := c.GetString("channel")
	if channel == "" {
		channel = "unknown"
	}
	return "auth:" + channel + ":" + ReadIp(c)
}

// BuildLoginRateLimitKey scopes login attempts by channel, identifier and IP.
func BuildLoginRateLimitKey(c *gin.Context) string {

	channel := c.GetString("channel")
	if channel == "" {
		channel = "unknown"
	}
	return "login:" + channel + ":" + ReadIdentifier(c) + ":" + ReadIp(c)
}

// ProtectedRateLimitKey scopes authenticated routes by channel, user and
// business, falling back to channel and IP when no user is in context.
func ProtectedRateLimitKey(c *gin.Context) string {

	channel := c.GetString("channel")
	if channel == "" {
		channel = "unknown"
	}

	if value, exists := c.Get("user"); exists && value != nil {
		id, business := "", "none"

		switch user := value.(type) {
		case *models.User:
			if user != nil {
				id = user.ID
				if user.BusinessID != "" {
					business = user.BusinessID
				}
			}
		case models.User:
			id = user.ID
			if user.BusinessID != "" {
				business = user.BusinessID
			}
		case map[string]any:
			if v, ok := user["_id"].(string); ok {
				id = v
			}
			if v, ok := user["business"].(string); ok && v != "" {
				business = v
			}
		}

		if id != "" {
			return "protected:" + channel + ":" + id + ":" + business
		}
	}

	return "protected:" + channel + ":" + ReadIp(c)
}
