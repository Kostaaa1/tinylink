package redisdb

import (
	"context"
	"testing"
	"time"

	"github.com/Kostaaa1/tinylink/pkg/config"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func StartTest(t *testing.T) *redis.Client {
	t.Helper()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "lagaosiprovidnokopas",
		DB:       0,
		PoolSize: 25,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	require.Nil(t, redisClient.Ping(ctx).Err())
	redisClient.FlushAll(ctx)

	return redisClient
}

func Start(conf config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.DB,
		PoolSize: conf.PoolSize,
	})

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		return nil, err
	}

	// listening for keys that expire to make expired aliases available again. Should be used for generated aliases only. Missing flag that determines that.
	// ctx := context.Background()
	// pubsub := client.PSubscribe(ctx, fmt.Sprintf("__keyevent@%d__:expired", 0))
	// go func() {
	// 	defer func() {
	// 		fmt.Println("Closing redis gracefully...")
	// 		client.Close()
	// 		pubsub.Close()
	// 	}()
	// 	for {
	// 		select {
	// 		case <-ctx.Done():
	// 			fmt.Println("Closing redis gracefully...")
	// 			return
	// 		case msg, ok := <-pubsub.Channel():
	// 			if !ok {
	// 				return
	// 			}
	// 			fmt.Println("received pubsub message: ", msg)
	// 		}
	// 	}
	// }()

	return client, nil
}
