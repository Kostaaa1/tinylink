package redisdb

import (
	"context"
	"fmt"
	"time"

	"github.com/Kostaaa1/tinylink/internal/data"

	"github.com/redis/go-redis/v9"
)

type RedisTinylinkStore struct {
	client *redis.Client
}

func (r *RedisTinylinkStore) Close() error {
	return r.client.Close()
}

func (r *RedisTinylinkStore) getTokenTTL(ctx context.Context, userID string) time.Duration {
	tokenKey := fmt.Sprintf("tokens:%s", userID)
	ttl := r.client.TTL(ctx, tokenKey).Val()
	return ttl
}

func (r *RedisTinylinkStore) Save(ctx context.Context, tl *data.Tinylink) error {
	tlKey := fmt.Sprintf("%s:%s", tl.UserID, tl.Alias)
	exists, err := r.client.Exists(ctx, tlKey).Result()

	if err != nil {
		return err
	}
	if exists > 0 {
		return data.ErrAliasExists
	}

	_, err = r.client.TxPipelined(ctx, func(p redis.Pipeliner) error {
		err := p.HSet(ctx, tlKey, map[string]interface{}{
			"url":          tl.URL,
			"alias":        tl.Alias,
			"created_at":   tl.CreatedAt.Unix(),
			"qr.data":      tl.QR.Data,
			"qr.width":     tl.QR.Width,
			"qr.height":    tl.QR.Height,
			"qr.mime_type": tl.QR.MimeType,
			"qr.size":      tl.QR.Size,
		}).Err()

		if err != nil {
			return err
		}

		tokenTTL := r.getTokenTTL(ctx, tl.UserID)
		if tokenTTL > 0 {
			return p.Expire(ctx, tlKey, tokenTTL).Err()
		}

		return err
	})

	return err
}

func (r *RedisTinylinkStore) Get(ctx context.Context, userID, alias string) (*data.Tinylink, error) {
	pattern := fmt.Sprintf("%s:%s", userID, alias)
	v, err := r.client.Get(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	fmt.Println(v)
	// return data.MapToTinylink(v)
	return nil, nil
}

func (r *RedisTinylinkStore) List(ctx context.Context, userID string) ([]*data.Tinylink, error) {
	pattern := fmt.Sprintf("%s:*", userID)

	var cursor uint64
	links := []*data.Tinylink{}

	for {
		keys, newCursor, err := r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return links, err
		}

		if len(keys) == 0 {
			break
		}

		pipe := r.client.Pipeline()
		var cmds = []*redis.MapStringStringCmd{}

		for _, key := range keys {
			keyType := r.client.Type(ctx, key).Val()
			if keyType == "hash" {
				cmds = append(cmds, pipe.HGetAll(ctx, key))
			}
		}

		_, err = pipe.Exec(ctx)
		if err != nil {
			return links, err
		}

		for _, cmd := range cmds {
			v, err := cmd.Result()
			if err != nil {
				continue
			}
			mapped, err := data.MapToTinylink(v)
			if err != nil {
				continue
			}
			links = append(links, mapped)
		}

		cursor = newCursor
		if cursor == 0 {
			return links, nil
		}
	}

	return links, nil
}

func (r *RedisTinylinkStore) Delete(ctx context.Context, userID, alias string) error {
	pattern := fmt.Sprintf("%s:%s", userID, alias)

	exists, err := r.client.Exists(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if exists > 0 {
		if err := r.client.Del(ctx, pattern).Err(); err != nil {
			return err
		}
		if err := r.client.Del(ctx, fmt.Sprintf("global:%s", alias)).Err(); err != nil {
			return err
		}
	}

	return nil
}
