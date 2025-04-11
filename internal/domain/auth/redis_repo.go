package auth

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Kostaaa1/tinylink/internal/common/data"
	"github.com/redis/go-redis/v9"
)

type RedisTokenRepository struct {
	client *redis.Client
}

func (r *RedisTokenRepository) Store(ctx context.Context, token *RefreshToken) error {
	key := fmt.Sprintf("%s:%s", SessionKey, token.ID)
	return r.client.SetEx(ctx, key, token, token.ExpiresAt).Err()
}

func (r *RedisTokenRepository) Valid(ctx context.Context, uuid, userID string) error {
	key := fmt.Sprintf("%s:%s", SessionKey, uuid)

	tokenData, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return data.ErrNotFound
	}

	if err != nil {
		return err
	}

	var token RefreshToken
	if err := json.Unmarshal([]byte(tokenData), &token); err != nil {
		return err
	}

	// durationSinceCreated := time.Since(time.Unix(token.CreatedAt, 0))
	// if durationSinceCreated > time {
	// 	return ErrRefreshTokenExpired
	// }

	if token.UserID != userID {
		return ErrTokenNotValid
	}

	return nil
}

func (r *RedisTokenRepository) Revoke(ctx context.Context, uuid string) error {
	return nil
}

func (r *RedisTokenRepository) RevokeAll(ctx context.Context, userID string) error {
	return nil
}
