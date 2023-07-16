package redisdb_test

import (
	"context"
	"strconv"
	"strings"
	"testing"
	"time"
	"zenquote/internal/config"
	"zenquote/internal/redisdb"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/require"
)

func TestRedisStorage(t *testing.T) {
	t.Parallel()

	// Start miniredis server
	server, err := miniredis.Run()
	require.NoError(t, err)

	defer server.Close()

	// Initialize RedisStorage with miniredis settings
	addr := strings.Split(server.Addr(), ":")
	port, _ := strconv.Atoi(addr[1])
	cfg := config.Config{
		Redis: config.Redis{
			Host: addr[0],
			Port: uint16(port),
		},
	}
	storage := redisdb.NewRedisStorage(cfg)

	key := "testkey"
	value := "testvalue"
	ttl := 10 * time.Second

	// Test Store function
	err = storage.Store(context.Background(), key, value, ttl)
	require.NoError(t, err)

	// Test Get function
	storedValue, err := storage.Get(context.Background(), key)
	require.NoError(t, err)
	require.Equal(t, value, storedValue)

	// Test Delete function
	err = storage.Delete(context.Background(), key)
	require.NoError(t, err)

	// Verify that value has been deleted
	_, err = storage.Get(context.Background(), key)
	require.Error(t, err)
	require.ErrorContains(t, err, "redis: nil")
}
