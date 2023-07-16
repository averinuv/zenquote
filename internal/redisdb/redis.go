package redisdb

import (
	"context"
	"fmt"
	"time"

	"zenquote/internal/config"

	"github.com/redis/go-redis/v9"
)

type RedisStorage struct {
	rdb *redis.Client
}

func NewRedisStorage(cfg config.Config) *RedisStorage {
	addr := fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port)
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	return &RedisStorage{rdb: rdb}
}

func (r *RedisStorage) Store(ctx context.Context, key string, value string, ttl time.Duration) error {
	err := r.rdb.Set(ctx, key, value, ttl).Err()
	if err != nil {
		return fmt.Errorf("set key value failed: %w", err)
	}

	return nil
}

func (r *RedisStorage) Get(ctx context.Context, key string) (string, error) {
	val, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		return "", fmt.Errorf("get by key failed: %w", err)
	}

	return val, nil
}

func (r *RedisStorage) Delete(ctx context.Context, key string) error {
	if _, err := r.rdb.Del(ctx, key).Result(); err != nil {
		return fmt.Errorf("delete by key failed: %w", err)
	}

	return nil
}
