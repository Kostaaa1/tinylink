package redisrepo

import (
	"context"
	"fmt"
	"strings"

	"github.com/Kostaaa1/tinylink/internal/data"
	"github.com/Kostaaa1/tinylink/internal/repository"
	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisRepo(ctx context.Context, opt *redis.Options) repository.Store {
	return &RedisRepository{
		client: redis.NewClient(opt),
		ctx:    ctx,
	}
}

func (r *RedisRepository) GetAll(key string) (map[string]data.Tinylink, error) {
	pattern := fmt.Sprintf("client:%s:tinylink:*", key)
	iter := r.client.Scan(r.ctx, 0, pattern, 0).Iterator()

	links := map[string]data.Tinylink{}

	for iter.Next(r.ctx) {
		redisKey := iter.Val()

		parts := strings.Split(redisKey, ":")
		tlHash := parts[3]

		var tl data.Tinylink
		if err := r.client.HGetAll(r.ctx, redisKey).Scan(&tl); err != nil {
			return nil, err
		}

		links[tlHash] = tl
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return links, nil
}

func (r *RedisRepository) Get(key, shortURL string) (data.Tinylink, error) {
	pattern := fmt.Sprintf("client:%s:tinylink", key)
	vals, err := r.client.HGet(r.ctx, pattern, shortURL).Result()
	if err != nil {
		return data.Tinylink{}, err
	}
	fmt.Println("GET VALUES: ", vals)
	return data.Tinylink{}, nil
}

func (r *RedisRepository) Ping() error {
	_, err := r.client.Ping(r.ctx).Result()
	return err
}

func (r *RedisRepository) Delete() {
}

func (r *RedisRepository) Set(id, key string, data interface{}) error {
	return r.client.HSet(r.ctx, id, key, data).Err()
}

func (r *RedisRepository) Check() bool { return false }
