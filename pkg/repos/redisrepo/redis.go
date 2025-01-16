package redisrepo

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
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

func (r *RedisRepository) Create(ctx context.Context, tl *models.Tinylink, qp storage.QueryParams) (*models.Tinylink, error) {
	var buff bytes.Buffer
	enc := gob.NewEncoder(&buff)
	enc.Encode(tl)

	pattern := fmt.Sprintf("client:%s:tinylink:%s", qp.ClientID, tl.Alias)
	if err := r.client.Set(ctx, pattern, buff.Bytes(), 0).Err(); err != nil {
		return nil, err
	}

	if qp.CheckUnique {
		isValid, err := r.ValidAlias(ctx, tl.Alias)
		if err != nil {
			return nil, err
		}

		if !isValid {
			return nil, errors.New("provided alias is taken")
		}

		if err := r.client.Set(ctx, "unique", tl.Alias, 0).Err(); err != nil {
			return nil, fmt.Errorf("failed to create unique alias: %w", err)
		}
	}

	return r.Get(ctx, storage.QueryParams{ClientID: qp.ClientID, Alias: tl.Alias})
}

func (r *RedisRepository) Get(ctx context.Context, qp storage.QueryParams) (*models.Tinylink, error) {
	pattern := fmt.Sprintf("client:%s:tinylink:%s", qp.ClientID, qp.Alias)

	v, err := r.client.Get(ctx, pattern).Result()
	if err != nil {
		return nil, err
	}

	var dbuf = bytes.NewBuffer([]byte(v))
	dec := gob.NewDecoder(dbuf)

	var tl models.Tinylink
	if err := dec.Decode(&tl); err != nil {
		return nil, err
	}

	return &tl, nil
}

func (r *RedisRepository) ValidAlias(ctx context.Context, alias string) (bool, error) {
	n, err := r.client.Exists(ctx, fmt.Sprintf("unique:%s", alias)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

func (r *RedisRepository) GetAll(ctx context.Context, qp storage.QueryParams) ([]*models.Tinylink, error) {
	keys, err := r.client.Keys(ctx, fmt.Sprintf("client:%s:tinylink:*", qp.ClientID)).Result()
	fmt.Println(keys)
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return nil, errors.New("you have 0 tinylinks")
	}

	values, err := r.client.MGet(ctx, keys...).Result()
	if err != nil {
		return nil, err
	}

	var links []*models.Tinylink
	for _, val := range values {
		strVal, ok := val.(string)
		if !ok {
			return nil, errors.New("assertion to string from interface{} in GetAll failed")
		}
		dbuv := []byte(strVal)

		var dbuf = bytes.NewBuffer(dbuv)
		dec := gob.NewDecoder(dbuf)

		var tl models.Tinylink
		if err := dec.Decode(&tl); err != nil {
			return nil, err
		}
		links = append(links, &tl)
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
