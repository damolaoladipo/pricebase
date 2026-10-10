package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/redis/go-redis/v9"

	applogger "github.com/damolaoladipo/pricebase/internal/common/logger"
	"go.uber.org/zap"
)

type Platform string

const (
	CloudManaged Platform = "cloud-managed"
	Provisioned  Platform = "provisioned"
)

type Options struct {
	Host     string
	Port     int
	Password string
	User     string
	Platform Platform
}

type Data struct {
	Key   string
	Value any
}

type RedisWrapper struct {
	client *redis.Client
}

func NewRedis() *RedisWrapper {
	return &RedisWrapper{}
}

func (r *RedisWrapper) Connect(ctx context.Context, options Options) error {

	var opts *redis.Options

	// if the platform is cloud-managed, use the cloud-managed options
	if options.Platform == CloudManaged {
		opts = &redis.Options{
			Addr:      fmt.Sprintf("%s:%d", options.Host, options.Port),
			Username:  options.User,
			Password:  options.Password,
			TLSConfig: nil,
		}

		// if the platform is provisioned, use the provisioned options
	} else if options.Platform == Provisioned {
		opts = &redis.Options{
			Addr:     fmt.Sprintf("%s:%d", options.Host, options.Port),
			Username: options.User,
			Password: options.Password,
		}
	} else {
		return fmt.Errorf("unsupported Redis platform: %s", options.Platform)
	}

	r.client = redis.NewClient(opts)

	if err := r.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis connection failed: %w", err)
	}

	color.New(color.FgYellow, color.ReverseVideo).Println("Connected to REDIS")
	applogger.SystemLogger.Info("Connected to REDIS", zap.String("host", options.Host), zap.Int("port", options.Port),
		zap.String("platform", string(options.Platform)))

	return nil
}

func (r *RedisWrapper) KeepData(ctx context.Context, data Data, exp int) error {

	parsed, err := json.Marshal(data.Value)
	if err != nil {
		return fmt.Errorf("failed to serialize Redis value: %w", err)
	}

	return r.client.Set(ctx, data.Key, parsed, time.Duration(exp)*time.Second).Err()
}

func (r *RedisWrapper) KeepIfAbsent(ctx context.Context, data Data, exp int) (bool, error) {

	parsed, err := json.Marshal(data.Value)
	if err != nil {
		return false, fmt.Errorf("failed to serialize Redis value: %w", err)
	}

	result, err := r.client.SetNX(ctx, data.Key, parsed, time.Duration(exp)*time.Second).Result()
	if err != nil {
		return false, err
	}

	return result, nil
}

func (r *RedisWrapper) SKeep(ctx context.Context, data Data) error {

	parsed, err := json.Marshal(data.Value)
	if err != nil {
		return fmt.Errorf("failed to serialize Redis value: %w", err)
	}

	return r.client.SAdd(ctx, data.Key, parsed).Err()
}

func (r *RedisWrapper) FetchData(ctx context.Context, key string) (any, error) {

	data, err := r.client.Get(ctx, key).Result()

	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var result any

	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil, fmt.Errorf("failed to deserialize Redis value: %w", err)
	}

	return result, nil
}

func (r *RedisWrapper) FetchMembers(ctx context.Context, key string) ([]any, error) {

	members, err := r.client.SMembers(ctx, key).Result()

	if err != nil {
		return nil, err
	}

	result := make([]any, 0, len(members))

	for _, member := range members {
		var data any

		if err := json.Unmarshal([]byte(member), &data); err != nil {
			result = append(result, member)
			continue
		}

		result = append(result, data)
	}

	return result, nil
}

func (r *RedisWrapper) DeleteData(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisWrapper) SDelete(ctx context.Context, key string, value any) error {

	var member string

	if valueString, ok := value.(string); ok {
		member = valueString
	} else {
		parsed, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("failed to serialize Redis member: %w", err)
		}

		member = string(parsed)
	}

	return r.client.SRem(ctx, key, member).Err()
}

func (r *RedisWrapper) Client() *redis.Client {
	return r.client
}

// REDIS is the process-wide shared instance, mirroring the TS
// `export default new RedisWrapper()`. Import it wherever cache access is
// needed instead of creating or passing wrappers around.
var REDIS = NewRedis()

// DialTimeout:  5 * time.Second
// ReadTimeout:  3 * time.Second
// WriteTimeout: 3 * time.Second
// PoolSize:     20
// MinIdleConns: 5
// MaxIdleConns: 10
