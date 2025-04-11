package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/Kostaaa1/tinylink/internal/common/data"
	"github.com/redis/go-redis/v9"
)

type RedisTokenRepository struct {
	client *redis.Client
}

func NewRedisTokenRepository(c *redis.Client) *RedisTokenRepository {
	return &RedisTokenRepository{
		client: c,
	}
}

func (r *RedisTokenRepository) TxDelOldAndInsertNew(ctx context.Context, userID, oldToken, newToken string, ttl time.Duration) (string, error) {
	key := fmt.Sprintf("refresh:%s", oldToken)

	tx := r.client.TxPipeline()

	fetchedID, err := tx.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", data.ErrNotFound
		}
		return "", fmt.Errorf("failed to get token: %w", err)
	}

	if fetchedID != userID {
		return "", ErrTokenNotValid
	}

	newKey := fmt.Sprintf("refresh:%s", newToken)
	if err := tx.SetEx(ctx, newKey, fetchedID, ttl).Err(); err != nil {
		return "", fmt.Errorf("failed to insert token: %w", err)
	}

	if err := tx.Del(ctx, key).Err(); err != nil {
		return "", fmt.Errorf("failed to delete token: %w", err)
	}

	return fetchedID, nil
}

func (r *RedisTokenRepository) GetUserID(ctx context.Context, tokenID string) (string, error) {
	key := fmt.Sprintf("%s:%s", "refresh", tokenID)
	userID, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}
	return userID, nil
}

func (r *RedisTokenRepository) Store(ctx context.Context, tokenID, userID string, ttl time.Duration) error {
	key := fmt.Sprintf("%s:%s", "refresh", tokenID)
	return r.client.SetEx(ctx, key, userID, ttl).Err()
}

func (r *RedisTokenRepository) Revoke(ctx context.Context, tokenID string) error {
	key := fmt.Sprintf("%s:%s", "refresh", tokenID)
	return r.client.Del(ctx, key).Err()
}
