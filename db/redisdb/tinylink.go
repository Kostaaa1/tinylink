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

// maybe add retries?
func (r *RedisTinylinkStore) Save(ctx context.Context, tl *data.Tinylink, userID string, ttl time.Duration) error {
	tlKey := fmt.Sprintf("%s:%s", userID, tl.Alias)
	urlKey := fmt.Sprintf("%s:url:%s", userID, tl.URL.RawPath)

	return r.client.Watch(ctx, func(tx *redis.Tx) error {
		tlExists, err := tx.Exists(ctx, tlKey).Result()
		if err != nil {
			return err
		}
		if tlExists > 0 {
			return data.ErrAliasExists
		}

		urlExists, err := tx.Exists(ctx, urlKey).Result()
		if err != nil {
			return err
		}
		if urlExists > 0 {
			return data.ErrURLExists
		}

		_, err = tx.TxPipelined(ctx, func(p redis.Pipeliner) error {
			p.HSet(ctx, tlKey, map[string]interface{}{
				"url":          tl.URL.String(),
				"alias":        tl.Alias,
				"created_at":   tl.CreatedAt.Unix(),
				"qr.data":      tl.QR.Data,
				"qr.width":     tl.QR.Width,
				"qr.height":    tl.QR.Height,
				"qr.mime_type": tl.QR.MimeType,
				"qr.size":      tl.QR.Size,
			})
			p.Set(ctx, urlKey, tl.URL.RawPath, ttl)
			p.Expire(ctx, tlKey, ttl)
			return nil
		})

		return err
	}, tlKey, urlKey)
}

func (r *RedisTinylinkStore) Get(ctx context.Context, userID, alias string) (*data.Tinylink, error) {
	pattern := fmt.Sprintf("%s:%s", userID, alias)
	v, err := r.client.HGetAll(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}
	return data.MapToTinylink(v)
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

	ok, err := r.Exists(ctx, pattern)
	if err != nil {
		return err
	}

	if ok {
		if err := r.client.Del(ctx, pattern).Err(); err != nil {
			return err
		}
		if err := r.client.Del(ctx, fmt.Sprintf("global:%s", alias)).Err(); err != nil {
			return err
		}
	}

	return nil
}

func (r *RedisTinylinkStore) Exists(ctx context.Context, id string) (bool, error) {
	n, err := r.client.Exists(ctx, id).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// func (r *RedisTinylinkStore) SetAlias(ctx context.Context, alias string) error {
// 	pattern := fmt.Sprintf("global:%s", alias)
// 	ok, err := r.Exists(ctx, pattern)
// 	if err != nil {
// 		return err
// 	}
// 	if !ok {
// 		if err := r.client.Set(ctx, pattern, nil, 0).Err(); err != nil {
// 			return err
// 		}
// 		return nil
// 	}
// 	return data.ErrAliasExists
// }

// func (r *RedisTinylinkStore) SetOriginalURL(ctx context.Context, clientID, URL string) error {
// 	pattern := fmt.Sprintf("%s:url:%s", clientID, URL)
// 	if err := r.client.Set(ctx, pattern, nil, 0).Err(); err != nil {
// 		return err
// 	}
// 	return errors.ErrURLExists
// }
