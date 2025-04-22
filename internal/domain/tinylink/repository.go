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

type LinkReader interface {
	GetURL(ctx context.Context, alias string) (uint64, string, error)
	GetPersonalURL(ctx context.Context, userID, alias string) (uint64, string, error)
}

type LinkWriter interface {
	// deletes the tinylink
	Delete(ctx context.Context, userID, id string) error
	// function that creates new tinylink. If user is authenticated, it will use userID from access token and it will store in persisten DB (sqlite). Otherwise, it will use sessionID from session cookie, and it will be stored under that session key in redis. If no userID and sessionID, respond with 401
	Create(ctx context.Context, tl *Tinylink) error
	// updates the tinylink. only for auth users
	Update(ctx context.Context, tl *Tinylink) error
}

type CacheStore interface {
	// caching the url and rowid for redirects
	CacheURL(ctx context.Context, id uint64, alias, url string) error
	// store tinylinks that are belonging to non-authenticated users. By default, session lasts 1 year.
	StoreBySessionID(ctx context.Context, sessionID string, tl map[string]interface{}) error
}

type LinkLister interface {
	// for redis, session ID needs to be used. For db persistence, use userID
	ListUserLinks(ctx context.Context, userID string) ([]*Tinylink, error)
}

type LinkChecker interface {
	// identifier - sessionID or userID
	Exists(ctx context.Context, identifier string, alias string) (bool, error)
}

type AliasService interface {
	GenerateAlias(ctx context.Context) (string, error)
}

type DBRepository interface {
	LinkReader
	LinkWriter
	LinkLister
	LinkChecker
}

type RedisRepository interface {
	LinkChecker
	LinkReader
	LinkLister
	CacheStore
	AliasService
}

// type Repository interface {
// 	Redirect(ctx context.Context, alias string) (uint64, string, error)
// }

// type RedisRepository interface {
// 	Repository
// 	GenerateAlias(ctx context.Context) (string, error)
// 	Exists(ctx context.Context, alias string) (bool, error)
// 	Insert(ctx context.Context, sessionID string, tl map[string]interface{})
// 	Cache(ctx context.Context, id, alias, url string)
// }

// type DBRepository interface {
// 	Repository
// 	RedirectPersonal(ctx context.Context, userID, alias string) (uint64, string, error)
// 	Insert(ctx context.Context, tl *Tinylink) error
// 	Get(ctx context.Context, alias string) (*Tinylink, error)
// 	GetByUserID(ctx context.Context, userID, alias string) (*Tinylink, error)
// 	Exists(ctx context.Context, userID string, alias string) (bool, error)
// 	List(ctx context.Context, userID string) ([]*Tinylink, error)
// 	Delete(ctx context.Context, userID, id string) error
// 	Update(ctx context.Context, tl *Tinylink) error
// }
