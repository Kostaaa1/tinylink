package tinylink

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type QueryParams struct {
	ClientID string
	Alias    string
}

type TinylinkRepository interface {
	List(qp QueryParams) ([]*Tinylink, error)
	Delete(qp QueryParams) error
	Save(tl *Tinylink, qp QueryParams) error
	Get(qp QueryParams) (*Tinylink, error)
	////////////////
	CheckAlias(alias string) error
	CheckOriginalURL(clientID, URL string) error
	// Ping(ctx context.Context) error
}

func NewTinylinkRepository(ctx context.Context, storageType string, db interface{}) (TinylinkRepository, error) {
	switch storageType {
	case "redis":
		return NewRedisTinylinkRepository(db.(*redis.Client), ctx), nil
	// case "sqlite":
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", storageType)
	}
}
