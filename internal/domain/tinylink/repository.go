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
	Redirect(ctx context.Context, alias string) (uint64, string, error)
	Insert(ctx context.Context, tl *Tinylink) error
}

type RedisRepository interface {
	Repository
	GenerateAlias(ctx context.Context) (string, error)
	// Insert(ctx context.Context, tl *Tinylink, ttl time.Duration) error
	Exists(ctx context.Context, alias string) (bool, error)
}

type DBRepository interface {
	Repository
	// Insert(ctx context.Context, tl *Tinylink) error
	Get(ctx context.Context, alias string) (*Tinylink, error)
	List(ctx context.Context, userID string) ([]*Tinylink, error)
	Delete(ctx context.Context, userID, id string) error
	Update(ctx context.Context, tl *Tinylink) error
	UpdateUsage(ctx context.Context, rowID uint64) error
	RedirectPersonal(ctx context.Context, userID, alias string) (uint64, string, error)
}
