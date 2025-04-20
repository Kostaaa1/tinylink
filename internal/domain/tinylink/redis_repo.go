package tinylink

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/Kostaaa1/tinylink/internal/domain/token"
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

// Used for ecnoding unique numbers (from redis counter) that will represent aliases for short urls. Its base62 because it includes 62 different characters: 0-9 (10), A-Z (26), a-z (26) [no duplicates].
func base62Encode(num int64) string {
	alias := ""
	for num > 0 {
		remainder := num % 62
		alias = string(base62Chars[remainder]) + alias
		num /= 62
	}
	return alias
}

func (r *TinylinkRedisRepository) GenerateAlias(ctx context.Context) (string, error) {
	value, err := r.client.Incr(ctx, "tinylink_count").Result()
	if err != nil {
		return "", fmt.Errorf("failed to increment alias counter: %w", err)
	}
	alias := base62Encode(value)
	return alias, nil
}

func (r *TinylinkRedisRepository) ListUserLinks(ctx context.Context, userID string) ([]*Tinylink, error) {
	return nil, nil
}

func (r *TinylinkRedisRepository) Exists(ctx context.Context, alias string) (bool, error) {
	exists, err := r.client.Exists(ctx, alias).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func (r *TinylinkRedisRepository) StoreBySessionID(ctx context.Context, sessionID string, tl map[string]interface{}) error {
	pipe := r.client.Pipeline()

	if _, err := pipe.HSet(ctx, sessionID, tl).Result(); err != nil {
		return fmt.Errorf("failed to HSET: %w", err)
	}

	pipe.Expire(ctx, sessionID, token.SessionTTL)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("pipeline failed: %w", err)
	}

	return nil
}

// cache for fast redirects
func (r *TinylinkRedisRepository) CacheURL(ctx context.Context, id uint64, alias, url string) error {
	pipe := r.client.Pipeline()

	if _, err := pipe.HSet(ctx, alias, map[string]interface{}{
		"id":  id,
		"url": url,
	}).Result(); err != nil {
		return fmt.Errorf("failed to HSET cacheURL: %w", err)
	}
	pipe.Expire(ctx, alias, cacheTTL)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute pipeline cacheURL: %w", err)
	}

	return nil
}

func (r *TinylinkRedisRepository) GetPersonalURL(ctx context.Context, userID, alias string) (uint64, string, error) {
	return 0, "", nil
}

func (r *TinylinkRedisRepository) GetURL(ctx context.Context, alias string) (uint64, string, error) {
	res, err := r.client.HMGet(ctx, alias, "url", "id").Result()
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
