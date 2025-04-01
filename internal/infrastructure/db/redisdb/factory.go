package redisdb

import (
	"context"
	"log"
	"time"

	"github.com/Kostaaa1/tinylink/pkg/config"
	"github.com/redis/go-redis/v9"
)

type Repositories struct {
	Tinylink *TinylinkRepository
	Token    *TokenRepository
}

func NewRepositories(cfg *config.RedisConfig) *Repositories {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatal(err)
	}

	return &Repositories{
		Tinylink: &TinylinkRepository{client: client},
		Token:    &TokenRepository{client: client},
	}
}
