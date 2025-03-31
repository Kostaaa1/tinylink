package redisdb

import "github.com/redis/go-redis/v9"

type Repositories struct {
	client *redis.Client
}
