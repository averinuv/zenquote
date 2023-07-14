package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"zenquote/internal/config"
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
	return r.rdb.Set(ctx, key, value, ttl).Err()
}

func (r *RedisStorage) Get(ctx context.Context, key string) (string, error) {
	val, err := r.rdb.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return val, nil
}

func (r *RedisStorage) Delete(ctx context.Context, key string) error {
	_, err := r.rdb.Del(ctx, key).Result()
	if err != nil {
		return err
	}
	return nil
}
