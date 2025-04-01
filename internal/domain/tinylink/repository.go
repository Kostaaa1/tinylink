package tinylink

import (
	"context"
)

type Repository interface {
	List(ctx context.Context, userID string) ([]*Tinylink, error)
	Delete(ctx context.Context, userID, id string) error
	Insert(ctx context.Context, tl *Tinylink) error
	Update(ctx context.Context, tl *Tinylink) error
	Get(ctx context.Context, userID, alias string) (*Tinylink, error)
}

type RedisRepository interface {
	Repository
	GenerateAlias(ctx context.Context, n int) (string, error)
}

type DBRepository interface {
	Repository
	IncrementUsageCount(ctx context.Context, alias string) error
	GetPublic(ctx context.Context, alias string) (*Tinylink, error)
}
