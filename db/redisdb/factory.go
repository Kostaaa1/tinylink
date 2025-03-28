package redisdb

import (
	"context"
	"log"
	"time"

	"github.com/Kostaaa1/tinylink/db"
	"github.com/Kostaaa1/tinylink/internal/data"
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatal(err)
	}

	return &db.RedisStore{
		Tinylink: &RedisTinylinkStore{client: client},
		Token:    &RedisTokenStore{client: client},
	}
}

func (s *RedisTinylinkStore) GetPublic(ctx context.Context, alias string) (*data.Tinylink, error) {
	return nil, nil
}
func (s *RedisTinylinkStore) IncrementUsageCount(ctx context.Context, alias string) error { return nil }

func NewRedisStoreFromClient(client *redis.Client) *db.RedisStore {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatal(err)
	}

	return &db.RedisStore{
		// Tinylink: &RedisTinylinkStore{client: client},
		Token: &RedisTokenStore{client: client},
	}
}
