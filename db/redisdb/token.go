package redisdb

import (
	"time"

	"github.com/Kostaaa1/tinylink/internal/data"
	"github.com/redis/go-redis/v9"
)

type RedisTokenStore struct {
	client *redis.Client
}

func (s *RedisTokenStore) New(userID int64, ttl time.Duration, scope data.Scope) (*data.Token, error) {
	return nil, nil
}

func (s *RedisTokenStore) Store(token *data.Token) error {
	return nil
}

func (s *RedisTokenStore) Get(token string) (*data.Token, error) {
	return nil, nil
}

func (s *RedisTokenStore) Revoke(token string) error {
	return nil
}

func (s *RedisTokenStore) Validate(token string) error {
	return nil
}

func (s *RedisTokenStore) ListActive() ([]*data.Token, error) {
	return nil, nil
}
