package redisdb

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/Kostaaa1/tinylink/internal/data"

	"github.com/redis/go-redis/v9"
)

type RedisTinylinkStore struct {
	client *redis.Client
}

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func encodeBase62(num int64) string {
	if num == 0 {
		return "0"
	}
	result := ""
	for num > 0 {
		remainder := num % 62
		result = string(base62Chars[remainder]) + result
		num /= 62
	}
	return result
}

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = base62Chars[randInt(0, len(base62Chars))]
	}
	return string(b)
}

func randInt(min, max int) int {
	return min + int(rand.Int31n(int32(max-min)))
}

func (r *RedisTinylinkStore) GenerateAlias(ctx context.Context, n int) (string, error) {
	key := fmt.Sprintf("tinylink_count")
	value, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return "", err
	}

	alias := ""
	for value > 0 {
		alias = string(base62Chars[value%62]) + alias
		value /= 62
		fmt.Println(alias, value)
	}

	if len(alias) < n {
		alias = fmt.Sprintf("%s%s", randomString(n-len(alias)), alias)
	}

	return alias, nil
}

func (r *RedisTinylinkStore) Close() error {
	return r.client.Close()
}

func (r *RedisTinylinkStore) getTokenTTL(ctx context.Context, userID string) time.Duration {
	return 0
}

func (r *RedisTinylinkStore) Update(ctx context.Context, tl *data.Tinylink) error {
	return nil
}

func (r *RedisTinylinkStore) Insert(ctx context.Context, tl *data.Tinylink) error {
	return nil
}

func (r *RedisTinylinkStore) Get(ctx context.Context, userID, alias string) (*data.Tinylink, error) {
	return nil, nil
}

func (r *RedisTinylinkStore) List(ctx context.Context, userID string) ([]*data.Tinylink, error) {
	return nil, nil
}

func (r *RedisTinylinkStore) Delete(ctx context.Context, userID, alias string) error {
	return nil
}
