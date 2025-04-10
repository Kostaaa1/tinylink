package tinylink

import (
	"context"
	"database/sql"
)

type db interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

type Repository interface {
	List(ctx context.Context, userID string) ([]*Tinylink, error)
	Delete(ctx context.Context, userID, id string) error
	Get(ctx context.Context, alias string) (*Tinylink, error)
	Insert(ctx context.Context, tl *Tinylink) error
	Update(ctx context.Context, tl *Tinylink) error
}

type RedisRepository interface {
	Repository
	GenerateAlias(ctx context.Context) (string, error)
}

type DBRepository interface {
	Repository
	UpdateUsage(ctx context.Context, tl *Tinylink) error
	GetByUserID(ctx context.Context, userID, alias string) (*Tinylink, error)
}
