package tinylink

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type TinylinkRedisRepository struct {
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

func (r *TinylinkRedisRepository) GenerateAlias(ctx context.Context) (string, error) {
	value, err := r.client.Incr(ctx, "tinylink_count").Result()
	if err != nil {
		return "", fmt.Errorf("failed to increment alias counter: %w", err)
	}
	alias := base62Encode(value)
	fmt.Println("incrementing and generating alias: ", value, alias)
	return alias, nil
	// n is fix length
	// length := len(alias)
	// if length < n {
	// 	padding := n - length
	// 	alias = fmt.Sprintf("%s%s", randStr(padding), alias)
	// }
}

func (r *TinylinkRedisRepository) Insert(ctx context.Context, tl *Tinylink) error {
	return nil
}

func (r *TinylinkRedisRepository) Close() error {
	return r.client.Close()
}

func (r *TinylinkRedisRepository) getTokenTTL(ctx context.Context, userID string) time.Duration {
	return 0
}

func (r *TinylinkRedisRepository) Update(ctx context.Context, tl *Tinylink) error {
	return nil
}

func (r *TinylinkRedisRepository) Get(ctx context.Context, alias string) (*Tinylink, error) {
	return nil, nil
}

func (r *TinylinkRedisRepository) List(ctx context.Context, userID string) ([]*Tinylink, error) {
	return nil, nil
}

func (r *TinylinkRedisRepository) Delete(ctx context.Context, userID, alias string) error {
	return nil
}
