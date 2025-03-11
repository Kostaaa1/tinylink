package redisdb

import (
	"context"
	"log"

	service "github.com/Kostaaa1/tinylink/internal/store"
	"github.com/Kostaaa1/tinylink/pkg/config"
	"github.com/redis/go-redis/v9"
)

func NewRedisStore(cfg *config.Config) *service.Store {
	pingctx := context.Background()

	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "lagaosiprovidnokopas",
		DB:       0,
	})

	if err := client.Ping(pingctx).Err(); err != nil {
		log.Fatal(err)
	}

	return &service.Store{
		Tinylink: NewRedisTinylinkRepository(client),
		User:     NewRedisUserRepository(client),
	}
}
