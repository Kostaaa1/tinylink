package redisdb

import (
	"context"
	"log"
	"time"

	"github.com/Kostaaa1/tinylink/internal/store"
	"github.com/Kostaaa1/tinylink/pkg/config"
	"github.com/redis/go-redis/v9"
)

func NewRedisStore(cfg *config.Config) *store.Store {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})

	pingctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	if err := client.Ping(pingctx).Err(); err != nil {
		log.Fatal(err)
	}

	return &store.Store{
		Tinylink: NewRedisTinylinkRepository(client),
		User:     NewRedisUserRepository(client),
	}
}
