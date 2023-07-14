package storage

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestRedisStorage(t *testing.T) {
	// Start miniredis server
	s, err := miniredis.Run()
	require.NoError(t, err)
	defer s.Close()

	// Initialize RedisStorage with miniredis settings
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	rs := &RedisStorage{rdb: rdb}

	key := "testkey"
	value := "testvalue"
	ttl := 10 * time.Second

	// Test Store function
	err = rs.Store(context.Background(), key, value, ttl)
	require.NoError(t, err)

	// Test Get function
	storedValue, err := rs.Get(context.Background(), key)
	require.NoError(t, err)
	require.Equal(t, value, storedValue)

	// Test Delete function
	err = rs.Delete(context.Background(), key)
	require.NoError(t, err)

	// Verify that value has been deleted
	_, err = rs.Get(context.Background(), key)
	require.Error(t, err)
	require.Equal(t, redis.Nil, err)
}
