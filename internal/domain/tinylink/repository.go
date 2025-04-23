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

type LinkPrivateRedirecter interface {
	// Used for redirects (redis - cached | sqlite - persisted). It returns row ID, original URL and error if fails. used in RedirectPrivate. Identifier can be (session ID - cookie value that's used for public redis tinylinks) and (user ID - used for persisted SQLite).
	RedirectURLByID(ctx context.Context, identifier, alias string) (uint64, string, error)
}

type LinkRedirecter interface {
	// Used for redirects (redis - cached | sqlite - persisted). Searches row by alias. It returns row ID, original URL and error if fails.
	RedirectURL(ctx context.Context, alias string) (uint64, string, error)
}

type LinkReader interface {
	// get data by rowID
	Get(ctx context.Context, rowID uint64) (*Tinylink, error)
}

type LinkWriter interface {
	// deletes the tinylink
	Delete(ctx context.Context, userID, id string) error
	// creates new tinylink. If user is authenticated, it will use userID from access token and it will store in persisten DB (sqlite). Otherwise, it will use sessionID from session cookie, and it will be stored under that session key in redis. If no userID and sessionID, respond with 401
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
	ListUserLinks(ctx context.Context, identifier string) ([]*Tinylink, error)
}

type AliasService interface {
	GenerateAlias(ctx context.Context) (string, error)
}

type GlobalAliasChecker interface {
	AliasExists(ctx context.Context, alias string) (bool, error)
}

type ScopedAliasChecker interface {
	AliasExistsWithID(ctx context.Context, identifier, alias string) (bool, error)
}

type FullAliasChecker interface {
	GlobalAliasChecker
	ScopedAliasChecker
}

type DBRepository interface {
	FullAliasChecker
	LinkRedirecter
	LinkPrivateRedirecter
	LinkReader
	LinkWriter
	LinkLister
}

type RedisRepository interface {
	GlobalAliasChecker
	LinkRedirecter
	LinkLister
	CacheStore
	AliasService
	// deletes all by identifier
	DeleteAll(ctx context.Context, identifier string) error
}
