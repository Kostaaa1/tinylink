package redis

import "github.com/redis/go-redis/v9"

func NewConnection(addr string, opts *redis.Options) *redis.Client {
	opts.Addr = addr
	return redis.NewClient(opts)
}
