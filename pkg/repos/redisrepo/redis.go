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
		Host:        v["host"],
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

func (r *RedisRepository) ValidOriginalURL(ctx context.Context, URL string, qp storage.QueryParams) (bool, error) {
	pattern := fmt.Sprintf("client:%s:tinylink:*", qp.ClientID)
	var cursor uint64

	for {
		keys, newCursor, err := r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return false, err
		}

		pipe := r.client.Pipeline()
		for _, key := range keys {
			u, err := pipe.HGet(ctx, key, "original_url").Result()
			if err != nil {
				return false, err
			}

			if URL == u {
				return true, nil
			}
		}
		_, err = pipe.Exec(ctx)
		if err != nil {
			return false, fmt.Errorf("failed to execute pipeline: %w", err)
		}

		cursor = newCursor
		if cursor == 0 {
			break
		}
	}
	return false, nil
}

func (r *RedisRepository) Create(ctx context.Context, tl *models.Tinylink, qp storage.QueryParams) (*models.Tinylink, error) {
	isValid, err := r.ValidOriginalURL(ctx, tl.OriginalURL, storage.QueryParams{ClientID: qp.ClientID})
	if err != nil {
		return nil, err
	}

	if !isValid {
		return nil, fmt.Errorf("you've already created tinylink for this URL: %s", tl.OriginalURL)
	}

	pattern := fmt.Sprintf("client:%s:tinylink:%s", qp.ClientID, tl.Alias)

	if _, err := r.client.Pipelined(ctx, func(rdb redis.Pipeliner) error {
		rdb.HSet(ctx, pattern, "host", tl.Host)
		rdb.HSet(ctx, pattern, "alias", tl.Alias)
		rdb.HSet(ctx, pattern, "original_url", tl.OriginalURL)
		rdb.HSet(ctx, pattern, "qr:data", tl.QR.Data)
		rdb.HSet(ctx, pattern, "qr:width", tl.QR.Width)
		rdb.HSet(ctx, pattern, "qr:height", tl.QR.Height)
		rdb.HSet(ctx, pattern, "qr:size", tl.QR.Size)
		rdb.HSet(ctx, pattern, "qr:mimetype", tl.QR.MimeType)
		return nil
	}); err != nil {
		return nil, err
	}

	// checking if alias is provided, then check if its valid

	qp.Alias = tl.Alias
	return r.Get(ctx, qp)
}

func (r *RedisRepository) Get(ctx context.Context, qp storage.QueryParams) (*models.Tinylink, error) {
	pattern := fmt.Sprintf("client:%s:tinylink:*", qp.ClientID)
	v, err := r.client.HGetAll(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}
	return populateTinylink(v), nil
}

func (r *RedisRepository) ValidAlias(ctx context.Context, alias string) (bool, error) {
	n, err := r.client.Exists(ctx, fmt.Sprintf("unique:%s", alias)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *RedisRepository) GetAll(ctx context.Context, qp storage.QueryParams) ([]*models.Tinylink, error) {
	pattern := fmt.Sprintf("client:%s:tinylink:*", qp.ClientID)

	var cursor uint64
	var links []*models.Tinylink

	for {
		keys, newCursor, err := r.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return links, nil
		}
		pipe := r.client.Pipeline()
		cmds := make([]*redis.MapStringStringCmd, len(keys))

		for i, key := range keys {
			cmds[i] = pipe.HGetAll(ctx, key)
		}

		_, err = pipe.Exec(ctx)
		if err != nil {
			return links, nil
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

func (r *RedisRepository) Ping(ctx context.Context) error {
	_, err := r.client.Ping(ctx).Result()
	return err
}

func (r *RedisRepository) Check(ctx context.Context, pattern, key string) bool {
	exists, err := r.client.Exists(ctx, pattern).Result()
	if err != nil {
		return false
	}
	return exists > 0
}
