package redisdb

import (
	"context"

	"github.com/Kostaaa1/tinylink/pkg/config"
	"github.com/redis/go-redis/v9"
)

func StartRedis(conf config.RedisConfig) (*redis.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5)
	defer cancel()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.DB,
		PoolSize: conf.PoolSize,
	})

	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return redisClient, nil
}

// type Repositories struct {
// 	Tinylink *TinylinkRepository
// 	Token    *TokenRepository
// }

// func NewRepositories(conf config.RedisConfig) *Repositories {
// 	client := redis.NewClient(&redis.Options{
// 		Addr:     conf.Addr,
// 		Password: conf.Password,
// 		DB:       conf.DB,
// 		PoolSize: conf.PoolSize,
// 	})

// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second*1)
// 	defer cancel()

// 	if err := client.Ping(ctx).Err(); err != nil {
// 		log.Fatal(err)
// 	}

// 	return &Repositories{
// 		Tinylink: &TinylinkRepository{client: client},
// 		Token:    &TokenRepository{client: client},
// 	}
// }
