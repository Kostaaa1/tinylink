package redisdb

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/Kostaaa1/tinylink/internal/domain/tinylink"
	"github.com/redis/go-redis/v9"
)

type TinylinkRepository struct {
	client *redis.Client
}

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func base62Encode(num int64) string {
	alias := ""
	for num > 0 {
		alias = string(base62Chars[num]) + alias
		num /= 62
	}
	return alias
}

func randStr(n int) string {
	var str strings.Builder
	for i := 0; i < n; i++ {
		randInt := rand.Intn(len(base62Chars))
		str.WriteString(string(base62Chars[randInt]))
	}
	return str.String()
}

func (r *TinylinkRepository) GenerateAlias(ctx context.Context, n int) (string, error) {
	value, err := r.client.Incr(ctx, "tinylink_count").Result()
	if err != nil {
		return "", err
	}

	alias := base62Encode(value)

	length := len(alias)
	if length < n {
		padding := n - length
		alias = fmt.Sprintf("%s%s", randStr(padding), alias)
	}

	return alias, nil
}

func (r *TinylinkRepository) Close() error {
	return r.client.Close()
}

func (r *TinylinkRepository) getTokenTTL(ctx context.Context, userID string) time.Duration {
	return 0
}

func (r *TinylinkRepository) Update(ctx context.Context, tl *tinylink.Tinylink) error {
	return nil
}

func (r *TinylinkRepository) Insert(ctx context.Context, tl *tinylink.Tinylink) error {
	return nil
}

func (r *TinylinkRepository) Get(ctx context.Context, userID, alias string) (*tinylink.Tinylink, error) {
	return nil, nil
}

func (r *TinylinkRepository) List(ctx context.Context, userID string) ([]*tinylink.Tinylink, error) {
	return nil, nil
}

func (r *TinylinkRepository) Delete(ctx context.Context, userID, alias string) error {
	return nil
}
