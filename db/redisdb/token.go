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

func (s *RedisTokenStore) RevokeAll(ctx context.Context, userID string, scope *data.Scope) error {
	tokenKey := fmt.Sprintf("tokens:%s", userID)

	tokens, err := s.client.SMembers(ctx, tokenKey).Result()
	if err != nil {
		return err
	}

	if len(tokens) == 0 {
		return nil
	}

	_, err = s.client.TxPipelined(ctx, func(p redis.Pipeliner) error {
		var found bool
		for _, token := range tokens {
			sessionKey := fmt.Sprintf("token:%s", token)

			if scope != nil {
				value, err := s.client.HGet(ctx, sessionKey, "scope").Result()
				if err != nil {
					return nil
				}

				if value == scope.String() {
					if err := p.Del(ctx, sessionKey).Err(); err != nil {
						return err
					}
					found = true
				}
			} else {
				if err := p.Del(ctx, sessionKey).Err(); err != nil {
					return err
				}
			}
		}

		if scope == nil {
			p.Del(ctx, tokenKey)
		} else if found {
			p.SRem(ctx, tokenKey, tokens)
		}

		return nil
	})

	return err
}

// tokens:user_id:tokenText - tokens:4:51d5kodsDa41 - set of tokens that belong to user, for revoking
// token:51d5kodsDa41 - holds token metadata
// token:51d5kodsDa41:token_data:
// token:51d5kodsDa41:tinylinks:
func (s *RedisTokenStore) Store(ctx context.Context, token *data.Token) error {
	sessionKey := fmt.Sprintf("token:%s", token.PlainText)

	_, err := s.client.TxPipelined(ctx, func(p redis.Pipeliner) error {
		if err := p.HSet(ctx, sessionKey, map[string]interface{}{
			"user_id": token.UserID,
			"token":   token.PlainText,
			"scope":   token.Scope.String(),
			"expiry":  token.Expiry.Unix(),
		}).Err(); err != nil {
			return err
		}

		if token.UserID != "" {
			tokenKey := fmt.Sprintf("tokens:%s", token.UserID)
			if err := p.SAdd(ctx, tokenKey, token.PlainText).Err(); err != nil {
				return err
			}
			if err := p.Expire(ctx, tokenKey, token.TTL).Err(); err != nil {
				return err
			}
		}

		return p.Expire(ctx, sessionKey, token.TTL).Err()
	})

	return err
}

func (s *RedisTokenStore) Get(ctx context.Context, tokenText string) (*data.Token, error) {
	key := fmt.Sprintf("token:%s", tokenText)

	values, err := s.client.HGetAll(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, data.ErrRecordNotFound
		}
		return nil, err
	}

	if len(values) == 0 {
		return nil, data.ErrRecordNotFound
	}

	ttl, err := s.client.TTL(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	v, _ := strconv.Atoi(values["expiry"])
	expiry := time.Unix(int64(v), 0)

	return &data.Token{
		PlainText: tokenText,
		UserID:    values["user_id"],
		Scope:     data.GetScope(values["scope"]),
		TTL:       ttl,
		Expiry:    expiry,
	}, nil
}
