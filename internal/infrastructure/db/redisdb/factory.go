package redisdb

import (
	"context"
	"fmt"
	"time"

	"github.com/Kostaaa1/tinylink/pkg/config"
	"github.com/redis/go-redis/v9"
)

func StartRedis(conf config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.DB,
		PoolSize: conf.PoolSize,
	})

	ctx := context.Background()
	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		return nil, err
	}

	pubsub := client.PSubscribe(ctx, fmt.Sprintf("__keyevent@%d__:expired", 0))
	// on tinylink expire -
	// if it singular alias - push it to the list
	// if its session_id
	go func() {
		defer func() {
			fmt.Println("Closing redis gracefully...")
			client.Close()
			pubsub.Close()
		}()
		for {
			select {
			case <-ctx.Done():
				fmt.Println("Closing redis gracefully...")
				return
			case msg, ok := <-pubsub.Channel():
				if !ok {
					return
				}
				fmt.Println("received pubsub message: ", msg)
			}
		}
	}()

	return client, nil
}
