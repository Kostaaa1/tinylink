package redisdb_test

import (
	"context"
	"testing"
	"time"

	"github.com/Kostaaa1/tinylink/db/redisdb"
	"github.com/Kostaaa1/tinylink/internal/data"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestRedisTinylinkFlow(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "lagaosiprovidnokopas",
		DB:       0,
		PoolSize: 25,
	})
	err := client.Ping(ctx).Err()
	require.Nil(t, err)
	store := redisdb.NewRedisStoreFromClient(client)

	ttl := time.Hour
	token := data.GenerateToken("4", ttl, data.ScopeAuthentication)
	err = store.Token.Store(ctx, token)
	require.Nil(t, err)

	////////////////

	userID := "4"
	alias := "testalias123"
	targetURL := "https://www.youtube.com/watch?v=tNZnLkRBYA8"

	tl, err := data.NewTinylink("http://localhost:3000", targetURL, alias)
	require.Nil(t, err)

	tl.CreatedAt = time.Now().Add(time.Hour * 24)
	err = store.Tinylink.Save(ctx, tl, userID)
	require.Nil(t, err)

	err = store.Tinylink.Save(ctx, tl, userID)
	require.Equal(t, data.ErrAliasExists, err)

	links, err := store.Tinylink.List(ctx, userID)
	require.Nil(t, err)

	require.Equal(t, 1, len(links))
}
