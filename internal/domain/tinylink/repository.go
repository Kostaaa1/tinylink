package tinylink

import (
	"context"
	"time"
)

type LinkLister interface {
	ListByUserID(ctx context.Context, userID uint64) ([]*Tinylink, error)
	ListByGuestUUID(ctx context.Context, guestUUID string) ([]*Tinylink, error)
}

type LinkWriter interface {
	Insert(ctx context.Context, tl *Tinylink) error
	Update(ctx context.Context, tl *Tinylink) error
	Delete(ctx context.Context, userID uint64, alias string) error
}

type DbRepository interface {
	LinkWriter
	LinkLister
	Redirect(ctx context.Context, userID *uint64, alias string) (*RedirectValue, error)
	Get(ctx context.Context, rowID uint64) (*Tinylink, error)
}

type CacheRepository interface {
	Redirect(ctx context.Context, alias string) (*RedirectValue, error)
	Cache(ctx context.Context, value RedirectValue, ttl time.Duration) error
	GenerateAlias(ctx context.Context) (string, error)
}
