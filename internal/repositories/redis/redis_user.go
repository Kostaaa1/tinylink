package redisdb

import (
	"context"

	service "github.com/Kostaaa1/tinylink/internal/store"
	"github.com/redis/go-redis/v9"
)

type RedisUserRepository struct {
	client *redis.Client
}

func NewRedisUserRepository(client *redis.Client) service.UserRepository {
	return RedisUserRepository{
		client: client,
	}
}

func (r RedisUserRepository) Save(ctx context.Context)    {}
func (r RedisUserRepository) GetByID(ctx context.Context) {}
