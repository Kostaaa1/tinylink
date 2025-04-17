package tinylink

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type TinylinkRedisRepository struct {
	client *redis.Client
}

func randStr(n int) string {
	var str strings.Builder
	for i := 0; i < n; i++ {
		randInt := rand.Intn(len(base62Chars))
		str.WriteString(string(base62Chars[randInt]))
	}
	return str.String()
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

// first check available aliases that expired, if found use it, if not increment counter and generate alias
func (r *TinylinkRedisRepository) GenerateAlias(ctx context.Context) (string, error) {
	value, err := r.client.Incr(ctx, "tinylink_count").Result()
	if err != nil {
		return "", fmt.Errorf("failed to increment alias counter: %w", err)
	}
	alias := base62Encode(value)
	return alias, nil
}

func key(alias string) string {
	return fmt.Sprintf("tinylink:%s", alias)
}

func (r *TinylinkRedisRepository) Exists(ctx context.Context, alias string) (bool, error) {
	key := key(alias)
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func (r *TinylinkRedisRepository) Insert(ctx context.Context, tl *Tinylink) error {
	key := fmt.Sprintf("tinylink:%s", tl.Alias)

	pipe := r.client.Pipeline()

	pipe.HSet(ctx, key, map[string]interface{}{
		"id":  tl.ID,
		"url": tl.URL,
	})

	ttl := tl.ExpiresAt - time.Now().Unix()
	fmt.Println("SEtting ttl: ", ttl)
	if ttl > 0 {
		pipe.Expire(ctx, key, time.Duration(ttl)*time.Second)
	}

	_, err := pipe.Exec(ctx)
	return err
}

func (r *TinylinkRedisRepository) Redirect(ctx context.Context, alias string) (uint64, string, error) {
	key := fmt.Sprintf("tinylink:%s", alias)

	res, err := r.client.HMGet(ctx, key, "url", "id").Result()
	if err != nil {
		return 0, "", err
	}

	if res[0] == nil || res[1] == nil {
		return 0, "", err
	}

	url := res[0].(string)
	rowID, err := strconv.ParseUint(res[1].(string), 10, 64)
	if err != nil {
		return 0, "", err
	}

	return rowID, url, nil
}
