package tinylink

import (
	"context"
	"fmt"

	tinylink "github.com/Kostaaa1/tinylink/internal/repository"
	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	client *redis.Client
	ctx    context.Context
}

func NewRedisRepository(client *redis.Client, ctx context.Context) *RedisRepository {
	return &RedisRepository{client: client, ctx: ctx}
}

func populateTinylink(v map[string]string) *domain.Tinylink {
	return &tinylink.Tinylink{
		Tinylink:    v["host"],
		Alias:       v["alias"],
		OriginalURL: v["original_url"],
		QR: repository.QR{
			Data:     []byte(v["qr:data"]),
			Width:    v["qr:width"],
			Height:   v["qr:height"],
			Size:     v["qr:size"],
			MimeType: v["qr:width"],
		},
	}
}

func (r *RedisRepository) ValidateOriginalURL(clientID, URL string) error {
	// O(1)
	pattern := fmt.Sprintf("client:%s:original_url:%s", clientID, URL)
	n, err := r.client.Exists(r.ctx, pattern).Result()
	if err != nil {
		return err
	}
	if n == 0 {
		if err := r.client.Set(r.ctx, pattern, nil, 0).Err(); err != nil {
			return err
		}
	}
	return nil
	// O(n)
	// pattern := fmt.Sprintf("client:%s:tinylink:*", clientID)
	// var cursor uint64
	// for {
	// 	keys, newCursor, err := r.client.Scan(r.ctx, cursor, pattern, 100).Result()
	// 	if err != nil {
	// 		return err
	// 	}
	// 	if len(keys) == 0 {
	// 		return nil
	// 	}
	// 	pipe := r.client.Pipeline()
	// 	cmds := make([]*redis.StringCmd, len(keys))
	// 	for i, key := range keys {
	// 		cmds[i] = pipe.HGet(r.ctx, key, "original_url")
	// 	}
	// 	_, err = pipe.Exec(ctx)
	// 	if err != nil {
	// 		return fmt.Errorf("failed to execute pipeline: %w", err)
	// 	}
	// 	for _, cmd := range cmds {
	// 		u, err := cmd.Result()
	// 		if err != nil {
	// 			return fmt.Errorf("failed to get cmd.Result()")
	// 		}
	// 		if URL == u {
	// 			return fmt.Errorf("you've already created tinylink for this URL: %s", URL)
	// 		}
	// 	}
	// 	cursor = newCursor
	// 	if cursor == 0 {
	// 		break
	// 	}
	// }
	// return nil
}

func (r *RedisRepository) ValidateAlias(alias string) error {
	n, err := r.client.Exists(r.ctx, fmt.Sprintf("unique:%s", alias)).Result()
	if err != nil {
		return err
	}
	if n > 0 {
		return fmt.Errorf("provided alias is taken: %s", alias)
	}
	if err := r.client.Set(r.ctx, fmt.Sprintf("unique:%s", alias), nil, 0).Err(); err != nil {
		return err
	}
	return nil
}

func (r *RedisRepository) Create(tl *domain.Tinylink, qp domain.QueryParams) error {
	pattern := fmt.Sprintf("client:%s:tinylink:%s", qp.ClientID, tl.Alias)
	if _, err := r.client.Pipelined(r.ctx, func(rdb redis.Pipeliner) error {
		rdb.HSet(r.ctx, pattern, "host", tl.Tinylink)
		rdb.HSet(r.ctx, pattern, "alias", tl.Alias)
		rdb.HSet(r.ctx, pattern, "original_url", tl.OriginalURL)
		rdb.HSet(r.ctx, pattern, "qr:data", tl.QR.Data)
		rdb.HSet(r.ctx, pattern, "qr:width", tl.QR.Width)
		rdb.HSet(r.ctx, pattern, "qr:height", tl.QR.Height)
		rdb.HSet(r.ctx, pattern, "qr:size", tl.QR.Size)
		rdb.HSet(r.ctx, pattern, "qr:mimetype", tl.QR.MimeType)
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (r *RedisRepository) Get(qp domain.QueryParams) (*domain.Tinylink, error) {
	pattern := fmt.Sprintf("client:%s:tinylink:%s", qp.ClientID, qp.Alias)
	v, err := r.client.HGetAll(r.ctx, pattern).Result()
	if err != nil {
		return nil, err
	}
	return populateTinylink(v), nil
}

func (r *RedisRepository) GetAll(qp domain.QueryParams) ([]*domain.Tinylink, error) {
	pattern := fmt.Sprintf("client:%s:tinylink:*", qp.ClientID)

	var cursor uint64
	links := []*domain.Tinylink{}

	for {
		keys, newCursor, err := r.client.Scan(r.ctx, cursor, pattern, 100).Result()
		if err != nil {
			return links, err
		}

		if len(keys) == 0 {
			break
		}

		pipe := r.client.Pipeline()
		cmds := make([]*redis.MapStringStringCmd, len(keys))

		for i, key := range keys {
			cmds[i] = pipe.HGetAll(r.ctx, key)
		}

		_, err = pipe.Exec(r.ctx)
		if err != nil {
			return links, err
		}

		for _, cmd := range cmds {
			v, err := cmd.Result()
			if err != nil {
				continue
			}
			links = append(links, populateTinylink(v))
		}

		cursor = newCursor
		if cursor == 0 {
			return links, nil
		}
	}

	return links, nil
}

func (r *RedisRepository) Delete(qp domain.QueryParams) error {
	pattern := fmt.Sprintf("client:%s:tinylink:%s", qp.ClientID, qp.Alias)
	if err := r.client.Del(r.ctx, pattern).Err(); err != nil {
		return err
	}
	uniqueKey := fmt.Sprintf("unique:%s", qp.Alias)
	if err := r.client.Del(r.ctx, uniqueKey).Err(); err != nil {
		return err
	}
	return nil
}

func (r *RedisRepository) Ping() error {
	_, err := r.client.Ping(r.ctx).Result()
	return err
}
