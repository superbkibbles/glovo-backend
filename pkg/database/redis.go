package database

import (
	"context"
	"time"

	"github.com/mendmzury/food-delivery/pkg/logger"
	"github.com/redis/go-redis/v9"
)

// RedisDB holds the Redis client
type RedisDB struct {
	Client *redis.Client
}

// NewRedisDB creates a new Redis connection
func NewRedisDB(uri, password string) (*RedisDB, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     uri,
		Password: password,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, err
	}

	logger.Info().Str("addr", uri).Msg("Connected to Redis")

	return &RedisDB{
		Client: client,
	}, nil
}

// Close closes the Redis connection
func (r *RedisDB) Close() error {
	logger.Info().Msg("Closing Redis connection")
	return r.Client.Close()
}

// Set stores a value with expiration
func (r *RedisDB) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return r.Client.Set(ctx, key, value, expiration).Err()
}

// Get retrieves a value
func (r *RedisDB) Get(ctx context.Context, key string) (string, error) {
	return r.Client.Get(ctx, key).Result()
}

// Delete removes a key
func (r *RedisDB) Delete(ctx context.Context, keys ...string) error {
	return r.Client.Del(ctx, keys...).Err()
}

// Exists checks if a key exists
func (r *RedisDB) Exists(ctx context.Context, key string) (bool, error) {
	result, err := r.Client.Exists(ctx, key).Result()
	return result > 0, err
}

// SetNX sets a value only if the key doesn't exist
func (r *RedisDB) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return r.Client.SetNX(ctx, key, value, expiration).Result()
}

// Incr increments a value
func (r *RedisDB) Incr(ctx context.Context, key string) (int64, error) {
	return r.Client.Incr(ctx, key).Result()
}

// Expire sets expiration on a key
func (r *RedisDB) Expire(ctx context.Context, key string, expiration time.Duration) error {
	return r.Client.Expire(ctx, key, expiration).Err()
}

// Keys returns all keys matching pattern
func (r *RedisDB) Keys(ctx context.Context, pattern string) ([]string, error) {
	return r.Client.Keys(ctx, pattern).Result()
}

const (
	KeyPrefixSession   = "session:"
	KeyPrefixRateLimit = "ratelimit:"
	KeyPrefixCache     = "cache:"
	KeyPrefixLock      = "lock:"
)
