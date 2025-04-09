package tinylink

import (
	"context"
	"database/sql"
)

type db interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	GetContext(ctx context.Context, dest any, query string, args ...any) error
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
}

type Repository interface {
	List(ctx context.Context, userID string) ([]*Tinylink, error)
	Delete(ctx context.Context, userID, id string) error
	Insert(ctx context.Context, tl *Tinylink) error
	Update(ctx context.Context, tl *Tinylink) error
	Get(ctx context.Context, alias string) (*Tinylink, error)
}

type RedisRepository interface {
	Repository
	GenerateAlias(ctx context.Context, n int) (string, error)
}

type DBRepository interface {
	Repository
	IncrementUsageCount(ctx context.Context, alias string) error
	GetByUserID(ctx context.Context, userID, alias string) (*Tinylink, error)
}
