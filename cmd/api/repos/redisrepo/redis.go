package redisrepo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Kostaaa1/tinylink/internal/data"
	"github.com/Kostaaa1/tinylink/internal/repository"
	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	client *redis.Client
}

func getTLPattern(clientID string) string {
	return fmt.Sprintf("client:%s:tinylink:*", clientID)
}

func NewRedisRepo(ctx context.Context, opt *redis.Options) repository.Storage {
	return &RedisRepository{
		client: redis.NewClient(opt),
	}
}

func getTinylinkPattern(userID string) string {
	return fmt.Sprintf("client:%s:tinylink", userID)
}

func (r *RedisRepository) GetAll(ctx context.Context, qp repository.StorageParams) ([]data.Tinylink, error) {
	pattern := fmt.Sprintf("client:%s:tinylink", qp.UserID)
	result, err := r.client.HGetAll(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	links := []data.Tinylink{}

	for key, value := range result {
		var tl data.Tinylink
		if err := json.Unmarshal([]byte(value), &tl); err != nil {
			return nil, err
		}
		tl.TinyURL = key
		links = append(links, tl)
	}

	return links, nil
}

func (r *RedisRepository) Get(ctx context.Context, qp repository.StorageParams) (data.Tinylink, error) {
	var tl data.Tinylink
	pattern := fmt.Sprintf("client:%s:tinylink", qp.UserID)

	val, err := r.client.HGet(ctx, pattern, qp.ID).Result()
	if err != nil {
		return tl, err
	}

	if err := json.Unmarshal([]byte(val), &tl); err != nil {
		return tl, err
	}

	return tl, nil
}

func (r *RedisRepository) Delete(ctx context.Context, qp repository.StorageParams) error {
	pattern := fmt.Sprintf("client:%s:tinylink", qp.UserID)
	if err := r.client.HDel(ctx, pattern, qp.ID).Err(); err != nil {
		return err
	}
	return nil
}

func (r *RedisRepository) Create(ctx context.Context, tl data.Tinylink, qp repository.StorageParams) error {
	b, err := json.Marshal(tl)
	if err != nil {
		return err
	}

	pattern := fmt.Sprintf("client:%s:tinylink", qp.UserID)
	if r.Check(ctx, pattern, tl.TinyURL) {
		return errors.New("url under this has already exists")
	}

	if _, err := r.client.HSet(ctx, pattern, tl.TinyURL, b).Result(); err != nil {
		return err
	}

	return nil
}

func (r *RedisRepository) Ping(ctx context.Context) error {
	_, err := r.client.Ping(ctx).Result()
	return err
}

func (r *RedisRepository) Check(ctx context.Context, pattern, key string) bool {
	exists, err := r.client.HExists(ctx, pattern, key).Result()
	if err != nil {
		return false
	}
	return exists
}
