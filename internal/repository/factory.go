package tinylink

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func NewTinylinkRepository(storageType string, db interface{}) (domain.TinylinkRepository, error) {
	switch storageType {
	case "redis":
		return tinylink.NewRedisRepository(db.(*redis.Client), context.Background()), nil
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", storageType)
	}
}
