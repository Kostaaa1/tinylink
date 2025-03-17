package redisdb

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Kostaaa1/tinylink/internal/data"
	"github.com/redis/go-redis/v9"
)

type RedisTokenStore struct {
	client *redis.Client
}

func (s *RedisTokenStore) RevokeAll(ctx context.Context, userID string) error {
	key := fmt.Sprintf("tokens:%s", userID)
	tokens, err := s.client.SMembers(ctx, key).Result()
	if err != nil {
		return err
	}

	_, err = s.client.TxPipelined(ctx, func(p redis.Pipeliner) error {
		for _, token := range tokens {
			if err := p.Del(ctx, fmt.Sprintf("session:%s", token)).Err(); err != nil {
				return err
			}
		}
		p.Del(ctx, key)
		return nil
	})

	return err
}

func (s *RedisTokenStore) Store(ctx context.Context, token *data.Token, sessionTTL time.Duration) error {
	fmt.Println("Store called")
	tokenKey := fmt.Sprintf("session:%s", token.PlainText)
	userTokensKey := fmt.Sprintf("tokens:%s", token.UserID)

	_, err := s.client.TxPipelined(ctx, func(p redis.Pipeliner) error {
		if err := p.HSet(ctx, tokenKey, map[string]interface{}{
			"user_id": token.UserID,
			"token":   token.PlainText,
			"expiry":  token.Expiry.Unix(),
			"scope":   token.Scope.String(),
		}).Err(); err != nil {
			return err
		}

		if err := p.Expire(ctx, tokenKey, token.TTL).Err(); err != nil {
			return err
		}

		if err := p.SAdd(ctx, userTokensKey, token.PlainText).Err(); err != nil {
			return err
		}

		return p.Expire(ctx, userTokensKey, sessionTTL).Err()
	})

	return err
}

func (s *RedisTokenStore) Get(ctx context.Context, tokenText string) (*data.Token, error) {
	fmt.Println("Get called")
	key := fmt.Sprintf("session:%s", tokenText)

	values, err := s.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	if len(values) == 0 {
		return nil, data.ErrRecordNotFound
	}

	unixTime, err := strconv.ParseInt(values["expiry"], 10, 64)
	if err != nil {
		return nil, err
	}
	expiry := time.Unix(unixTime, 0)

	return &data.Token{
		PlainText: tokenText,
		Expiry:    expiry,
		UserID:    values["user_id"],
		Scope:     data.GetScope(values["scope"]),
	}, nil
}
