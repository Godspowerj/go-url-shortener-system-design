package store

import (
	"context"
	"github.com/redis/go-redis/v9"
)

// RedisStore handles our caching logic for shortened URLs in memory.
type RedisStore struct {
	client *redis.Client
}

// NewRedisStore initializes and returns a new connection client to Redis.
func NewRedisStore(addr string) *RedisStore {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisStore{client: client}
}


// Set stores a key-value p	air in Redis with no expiration time.
func (r *RedisStore) Set(key string, value string) error {
	ctx := context.Background()
	err := r.client.Set(ctx, key, value, 0).Err()
	return err
}

// Get retrieves a value by key from Redis.
func (r *RedisStore) Get(key string) (string, bool) {
	ctx := context.Background()
	val, err := r.client.Get(ctx, key).Result()
	if err == nil {
		return val, true
	}
	return "", false
}
