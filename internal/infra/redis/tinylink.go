package redis

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Kostaaa1/tinylink/internal/constants"
	"github.com/Kostaaa1/tinylink/internal/domain/tinylink"
	"github.com/redis/go-redis/v9"
)

type TinylinkRepository struct {
	client *redis.Client
}

func NewTinylinkRepository(redis *redis.Client) *TinylinkRepository {
	return &TinylinkRepository{client: redis}
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

func (r *TinylinkRepository) GenerateAlias(ctx context.Context) (string, error) {
	value, err := r.client.Incr(ctx, "tinylink_count").Result()
	if err != nil {
		return "", fmt.Errorf("failed to increment alias counter: %w", err)
	}
	alias := base62Encode(value)
	return alias, nil
}

func (r *TinylinkRepository) Redirect(ctx context.Context, alias string) (*tinylink.RedirectValue, error) {
	key := fmt.Sprintf("cached_alias:%s", alias)

	value, err := r.client.HGetAll(ctx, key).Result()
	if err != nil {
		switch err {
		case redis.Nil:
			return nil, constants.ErrNotFound
		}
		return nil, err
	}

	if len(value) == 0 {
		return nil, fmt.Errorf("no cache key found for alias: %s", alias)
	}

	rowIDStr, ok := value["row_id"]
	if !ok || rowIDStr == "" {
		return nil, fmt.Errorf("missing row_id for alias: %s", alias)
	}

	rowID, err := strconv.ParseUint(rowIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parseUint failed for row_id: %s", rowIDStr)
	}

	return &tinylink.RedirectValue{
		RowID: rowID,
		Alias: value["alias"],
		URL:   value["url"],
	}, nil
}

func (r *TinylinkRepository) Cache(ctx context.Context, val tinylink.RedirectValue, ttl time.Duration) error {
	key := fmt.Sprintf("cached_alias:%s", val.Alias)

	cacheVal := map[string]string{
		"row_id": strconv.Itoa(int(val.RowID)),
		"url":    val.URL,
	}

	pipe := r.client.Pipeline()
	pipe.HSet(ctx, key, cacheVal)
	// always reset ttl
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)

	return err
}

func (r *TinylinkRepository) Save(
	ctx context.Context,
	uuid string,
	val *tinylink.Tinylink,
) error {
	var aliasKey string
	if val.Private {
		aliasKey = fmt.Sprintf("alias:private:%s:%s", uuid, val.Alias)
	} else {
		aliasKey = fmt.Sprintf("alias:public:%s", val.Alias)
	}

	mapValue := map[string]string{
		"url":        val.URL,
		"alias":      val.Alias,
		"created_at": val.CreatedAt.Format(time.RFC3339),
	}
	pipe := r.client.Pipeline()
	pipe.HSet(ctx, fmt.Sprintf("guest:%s:alias:%s", uuid, val.Alias), mapValue)
	pipe.Set(ctx, aliasKey, "", 0)
	_, err := pipe.Exec(ctx)

	return err
}

func (r *TinylinkRepository) List(ctx context.Context, uuid string) ([]*tinylink.Tinylink, error) {
	userDataKey := fmt.Sprintf("guest:%s:alias:*", uuid)
	keys, err := r.client.Keys(ctx, userDataKey).Result()
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		return []*tinylink.Tinylink{}, nil
	}

	links := make([]*tinylink.Tinylink, len(keys))
	cmds := make([]*redis.MapStringStringCmd, len(keys))

	pipe := r.client.Pipeline()
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

		createdAt, err := time.Parse(time.RFC3339, data["created_at"])
		if err != nil {
			return nil, err
		}

		links[i] = &tinylink.Tinylink{
			URL:       data["url"],
			Alias:     data["alias"],
			Private:   data["private"] == "true",
			CreatedAt: createdAt,
		}
	}

	return links, nil
}

// func (r *TinylinkRepository) AliasValid(ctx context.Context, alias string, uuid string) (bool, error) {
// 	keys, _ := r.client.Keys(ctx, "alias:*").Result()
// 	fmt.Println("KEYS: ", keys)

// 	pubKey := fmt.Sprintf("alias:public:%s", alias)
// 	pubEx, err := r.client.Exists(ctx, pubKey).Result()
// 	if err != nil {
// 		return false, err
// 	}

// 	privKey := fmt.Sprintf("alias:private:%s:%s", uuid, alias)
// 	privEx, err := r.client.Exists(ctx, privKey).Result()
// 	if err != nil {
// 		return false, err
// 	}

// 	return pubEx == 0 && privEx == 0, nil
// }
