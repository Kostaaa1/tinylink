package redisdb

import (
	"context"
	"log"
	"time"

	"github.com/Kostaaa1/tinylink/internal/store"
	"github.com/Kostaaa1/tinylink/pkg/config"
	"github.com/redis/go-redis/v9"
)

type Provider struct {
	client   *redis.Client
	tinylink store.TinylinkStore
}

func NewProvider(cfg *config.RedisConfig) *Provider {
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

	return &Provider{
		client:   client,
		tinylink: &RedisTinylinkStore{client: client},
	}
}

func (p *Provider) Tinylink() store.TinylinkStore {
	return p.tinylink
}

func (p *Provider) Close() error {
	return p.client.Close()
}
