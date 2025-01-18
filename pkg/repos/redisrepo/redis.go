package redisrepo

import (
	"context"
	"fmt"

	"github.com/Kostaaa1/tinylink/internal/models"
	"github.com/Kostaaa1/tinylink/internal/repository/storage"
	"github.com/redis/go-redis/v9"
)

type RedisRepository struct {
	client *redis.Client
}

func NewRedisRepo(ctx context.Context, opt *redis.Options) storage.Storage {
	return &RedisRepository{
		client: redis.NewClient(opt),
	}
}

func populateTinylink(v map[string]string) *models.Tinylink {
	return &models.Tinylink{
		Tinylink:    v["host"],
		Alias:       v["alias"],
		OriginalURL: v["original_url"],
		QR: models.QR{
			Data:     []byte(v["qr:data"]),
			Width:    v["qr:width"],
			Height:   v["qr:height"],
			Size:     v["qr:size"],
			MimeType: v["qr:width"],
		},
	}
}

func (r *RedisRepository) ValidateOriginalURL(ctx context.Context, clientID, URL string) error {
	pattern := fmt.Sprintf("client:%s:tinylink:*", clientID)
	var cursor uint64

	for {
		keys, newCursor, err := r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return err
		}

		if len(keys) == 0 {
			return nil
		}

		pipe := r.client.Pipeline()

		cmds := make([]*redis.StringCmd, len(keys))
		for i, key := range keys {
			cmds[i] = pipe.HGet(ctx, key, "original_url")
		}

		_, err = pipe.Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to execute pipeline: %w", err)
		}

		for _, cmd := range cmds {
			u, err := cmd.Result()
			if err != nil {
				return fmt.Errorf("failed to get cmd.Result()")
			}
			if URL == u {
				return fmt.Errorf("you've already created tinylink for this URL: %s", URL)
			}
		}

		cursor = newCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}

func (r *RedisRepository) ValidateAlias(ctx context.Context, alias string) error {
	n, err := r.client.Exists(ctx, fmt.Sprintf("unique:%s", alias)).Result()
	if err != nil {
		return err
	}
	if n > 0 {
		return fmt.Errorf("provided alias is taken: %s", alias)
	}
	if err := r.client.Set(ctx, fmt.Sprintf("unique:%s", alias), nil, 0).Err(); err != nil {
		return err
	}
	return nil
}

func (r *RedisRepository) Create(ctx context.Context, tl *models.Tinylink, qp storage.QueryParams) error {
	pattern := fmt.Sprintf("client:%s:tinylink:%s", qp.ClientID, tl.Alias)
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
	return nil
}

func (r *RedisRepository) Get(ctx context.Context, qp storage.QueryParams) (*models.Tinylink, error) {
	pattern := fmt.Sprintf("client:%s:tinylink:%s", qp.ClientID, qp.Alias)
	v, err := r.client.HGetAll(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}
	return populateTinylink(v), nil
}

func (r *RedisRepository) GetAll(ctx context.Context, qp storage.QueryParams) ([]*models.Tinylink, error) {
	pattern := fmt.Sprintf("client:%s:tinylink:*", qp.ClientID)

	var cursor uint64
	links := []*models.Tinylink{}

	for {
		keys, newCursor, err := r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return links, err
		}

		if len(keys) == 0 {
			break
		}

		pipe := r.client.Pipeline()
		cmds := make([]*redis.MapStringStringCmd, len(keys))

		for i, key := range keys {
			cmds[i] = pipe.HGetAll(ctx, key)
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
			links = append(links, populateTinylink(v))
		}

		cursor = newCursor
		if cursor == 0 {
			return links, nil
		}
	}

	return links, nil
}

func (r *RedisRepository) Delete(ctx context.Context, qp storage.QueryParams) error {
	pattern := fmt.Sprintf("client:%s:tinylink:%s", qp.ClientID, qp.Alias)
	if err := r.client.Del(ctx, pattern).Err(); err != nil {
		return err
	}
	uniqueKey := fmt.Sprintf("unique:%s", qp.Alias)
	if err := r.client.Del(ctx, uniqueKey).Err(); err != nil {
		return err
	}
	return nil
}

func (r *RedisRepository) Check(ctx context.Context, pattern, key string) bool {
	exists, err := r.client.Exists(ctx, pattern).Result()
	if err != nil {
		return false
	}
	return exists > 0
}

func (r *RedisRepository) Ping(ctx context.Context) error {
	_, err := r.client.Ping(ctx).Result()
	return err
}
