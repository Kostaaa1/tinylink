package tinylink

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"

	"github.com/Kostaaa1/tinylink/internal/common/data"
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

func (r *TinylinkRedisRepository) ListUserLinks(ctx context.Context, sessionID string) ([]*Tinylink, error) {
	key := fmt.Sprintf("%s:*", sessionID)

	keys, err := r.client.Keys(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get ListUserLinks keys: %w", err)
	}

	if len(keys) == 0 {
		return nil, data.ErrNotFound
	}

	pipe := r.client.Pipeline()
	cmds := make([]*redis.MapStringStringCmd, len(keys))
	tinylinks := make([]*Tinylink, len(keys))

	for i, key := range keys {
		cmds[i] = pipe.HGetAll(ctx, key)
	}

	_, err = pipe.Exec(ctx)
	if err != nil {
		return nil, err
	}

	for i, cmd := range cmds {
		data, err := cmd.Result()
		if err != nil {
			return nil, err
		}

		tl, err := FromMap(data)
		if err != nil {
			return nil, err
		}

		tinylinks[i] = tl
	}

	return tinylinks, nil
}

func (r *TinylinkRedisRepository) DeleteAll(ctx context.Context, sessionID string) error {
	iter := r.client.Scan(ctx, 0, fmt.Sprintf("%s:*", sessionID), 0).Iterator()
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
		alias, err := r.client.HGet(ctx, iter.Val(), "alias").Result()
		if err == nil && alias != "" {
			keys = append(keys, alias)
		}
	}

	if err := iter.Err(); err != nil {
		return err
	}

	if len(keys) == 0 {
		return fmt.Errorf("no keys to delete for: %s", sessionID)
	}

	_, err := r.client.Del(ctx, keys...).Result()
	return err
}

func (r *TinylinkRedisRepository) AliasExists(ctx context.Context, alias string) (bool, error) {
	key := fmt.Sprintf("alias:%s", alias)
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

func (r *TinylinkRedisRepository) StoreBySessionID(ctx context.Context, sessionID string, tl map[string]interface{}) error {
	alias := tl["alias"].(string)
	key := fmt.Sprintf("%s:%s", sessionID, alias)
	reverseKey := fmt.Sprintf("alias:%s", alias)

	pipe := r.client.Pipeline()
	pipe.SetEx(ctx, reverseKey, sessionID, token.SessionTTL)
	pipe.HSet(ctx, key, map[string]interface{}{
		"alias":      alias,
		"url":        tl["url"],
		"created_at": tl["created_at"],
	})
	pipe.Expire(ctx, key, token.SessionTTL)

	_, err := pipe.Exec(ctx)
	return err
}

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

func (r *TinylinkRedisRepository) RedirectURL(ctx context.Context, alias string) (uint64, string, error) {
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
