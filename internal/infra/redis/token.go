package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/Kostaaa1/tinylink/internal/domain/auth"
	"github.com/redis/go-redis/v9"
)

type Tokenository struct {
	client *redis.Client
}

func NewTokenRepository(c *redis.Client) *Tokenository {
	return &Tokenository{client: c}
}

var (
	refreshTokensKey = "refresh_tokens"
)

func (r *Tokenository) Save(ctx context.Context, userID uint64, refreshToken string, ttl time.Duration) error {
	key := fmt.Sprintf("%s:%d", refreshTokensKey, userID)
	return r.client.Set(ctx, key, refreshToken, ttl).Err()
}

func (r *Tokenository) Revoke(ctx context.Context, userID uint64) error {
	return r.client.Del(ctx, fmt.Sprintf("%s:%d", refreshTokensKey, userID)).Err()
}

func (r *Tokenository) Valid(ctx context.Context, userID uint64, token string) error {
	refreshToken, err := r.client.Get(ctx, fmt.Sprintf("%s:%d", refreshTokensKey, userID)).Result()
	if err != nil {
		return err
	}
	if refreshToken != token {
		return auth.ErrTokenNotValid
	}
	return nil
}

type Session struct {
	GuestUUID string
	CSRF      string
	IssuedAt  string
	ExpiresAt string
	IP        string
	UserAgent string
}

// func (r *Tokenository) CreateSession(ctx context.Context, sess Session) error {
// 	key := fmt.Sprintf("session:%s", sess.GuestUUID)
// 	pipe := r.client.Pipeline()
// 	pipe.HSet(ctx, key, map[string]string{
// 		"uuid":       sess.GuestUUID,
// 		"csrf":       sess.CSRF,
// 		"issued_at":  sess.IssuedAt,
// 		"expires_at": sess.ExpiresAt,
// 		"ip":         sess.IP,
// 		"user_agent": sess.UserAgent,
// 	})
// 	// set expires at
// }

// func (r *Tokenository) GetSession(ctx context.Context, guestUUID string) error {
// }
