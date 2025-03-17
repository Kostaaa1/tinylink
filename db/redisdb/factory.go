package redisdb

import (
	"context"
	"log"
	"time"

	"github.com/Kostaaa1/tinylink/db"
	"github.com/Kostaaa1/tinylink/pkg/config"
	"github.com/redis/go-redis/v9"
)

func NewRedisStore(cfg *config.RedisConfig) *db.RedisStore {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	pingctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	if err := client.Ping(pingctx).Err(); err != nil {
		log.Fatal(err)
	}

	return &db.RedisStore{
		Tinylink: &RedisTinylinkStore{client: client},
	}
}
