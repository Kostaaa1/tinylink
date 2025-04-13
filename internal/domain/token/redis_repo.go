package token

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisTokenRepository struct {
	client *redis.Client
}

func NewRedisTokenRepository(c *redis.Client) *RedisTokenRepository {
	return &RedisTokenRepository{client: c}
}

func (r *RedisTokenRepository) Save(ctx context.Context, userID, newToken string) error {
	fmt.Println("storing new token:", userID, newToken)
	newKey := key(userID)
	return r.client.SetEx(ctx, newKey, newToken, refreshTokenDuration).Err()
}

func (r *RedisTokenRepository) Delete(ctx context.Context, userID string) error {
	fmt.Println("deleting session for user: ", userID)
	newKey := key(userID)
	return r.client.Del(ctx, newKey).Err()
}

func (r *RedisTokenRepository) Valid(ctx context.Context, userID, token string) error {
	newKey := key(userID)
	rt, err := r.client.Get(ctx, newKey).Result()
	if err != nil {
		return err
	}
	if rt != token {
		return ErrTokenNotValid
	}
	return nil
}

func key(userID string) string {
	return fmt.Sprintf("%s:%s", "refresh", userID)
}
