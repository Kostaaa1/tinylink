package redisdb

import (
	"context"
	"fmt"

	"github.com/Kostaaa1/tinylink/internal/domain/entities"
	service "github.com/Kostaaa1/tinylink/internal/store"
	"github.com/Kostaaa1/tinylink/pkg/errors"

	"github.com/redis/go-redis/v9"
)

type RedisTinylinkRepository struct {
	client *redis.Client
}

func NewRedisTinylinkRepository(client *redis.Client) service.TinylinkRepository {
	return &RedisTinylinkRepository{
		client: client,
	}
}

func (r *RedisTinylinkRepository) Save(ctx context.Context, tl *entities.Tinylink, qp entities.QueryParams) error {
	pattern := fmt.Sprintf("client:%s:%s", qp.SessionID, tl.Alias)
	urlKey := fmt.Sprintf("client:%s:original_url:%s", qp.SessionID, tl.OriginalURL)

	ok, err := r.Exists(ctx, urlKey)
	if err != nil {
		return err
	}
	if ok {
		return errors.ErrURLExists
	}

	if _, err := r.client.Pipelined(ctx, func(rdb redis.Pipeliner) error {
		rdb.HSet(ctx, pattern, "host", tl.Tinylink)
		rdb.HSet(ctx, pattern, "alias", tl.Alias)
		rdb.HSet(ctx, pattern, "original_url", tl.OriginalURL)
		rdb.HSet(ctx, pattern, "qr:data", tl.QR.Data)
		rdb.HSet(ctx, pattern, "qr:width", tl.QR.Width)
		rdb.HSet(ctx, pattern, "qr:height", tl.QR.Height)
		rdb.HSet(ctx, pattern, "qr:size", tl.QR.Size)
		rdb.HSet(ctx, pattern, "qr:mimetype", tl.QR.MimeType)
		return nil
	}); err != nil {
		return err
	}

	if err := r.client.Set(ctx, urlKey, nil, 0).Err(); err != nil {
		return err
	}

	return nil
}

func (r *RedisTinylinkRepository) Get(ctx context.Context, qp entities.QueryParams) (*entities.Tinylink, error) {
	pattern := fmt.Sprintf("client:%s:%s", qp.SessionID, qp.Alias)
	v, err := r.client.HGetAll(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}
	return entities.MapToTinylink(v), nil
}

func (r *RedisTinylinkRepository) List(ctx context.Context, qp entities.QueryParams) ([]*entities.Tinylink, error) {
	pattern := fmt.Sprintf("client:%s:*", qp.SessionID)

	var cursor uint64
	links := []*entities.Tinylink{}

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
			links = append(links, entities.MapToTinylink(v))
		}

		cursor = newCursor
		if cursor == 0 {
			return links, nil
		}
	}

	return links, nil
}

func (r *RedisTinylinkRepository) Delete(ctx context.Context, qp entities.QueryParams) error {
	pattern := fmt.Sprintf("client:%s:%s", qp.SessionID, qp.Alias)

	ok, err := r.Exists(ctx, pattern)
	if err != nil {
		return err
	}

	if ok {
		if err := r.client.Del(ctx, pattern).Err(); err != nil {
			return err
		}
		if err := r.client.Del(ctx, fmt.Sprintf("global:%s", qp.Alias)).Err(); err != nil {
			return err
		}
	}

	return nil
}

func (r *RedisTinylinkRepository) Exists(ctx context.Context, id string) (bool, error) {
	n, err := r.client.Exists(ctx, id).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *RedisTinylinkRepository) SetAlias(ctx context.Context, alias string) error {
	pattern := fmt.Sprintf("global:%s", alias)
	ok, err := r.Exists(ctx, pattern)
	if err != nil {
		return err
	}

	if !ok {
		if err := r.client.Set(ctx, pattern, nil, 0).Err(); err != nil {
			return err
		}
		return nil
	}

	return errors.ErrAliasExists
}

// func (r *RedisTinylinkRepository) SetOriginalURL(ctx context.Context, clientID, URL string) error {
// 	pattern := fmt.Sprintf("client:%s:original_url:%s", clientID, URL)
// 	if err := r.client.Set(ctx, pattern, nil, 0).Err(); err != nil {
// 		return err
// 	}
// 	return errors.ErrURLExists
// }
