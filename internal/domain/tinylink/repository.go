package tinylink

import (
	"context"
	"time"

	"github.com/Kostaaa1/tinylink/core/transactor"
)

type LinkLister interface {
	ListByUserID(ctx context.Context, userID uint64) ([]*Tinylink, error)
	ListByGuestUUID(ctx context.Context, guestUUID string) ([]*Tinylink, error)
}

type DbRepository interface {
	LinkWriter
	LinkLister
	Redirect(ctx context.Context, alias string, userID *uint64) (*RedirectValue, error)
	Get(ctx context.Context, rowID uint64) (*Tinylink, error)
	WithRepositoryTx(tx transactor.Tx) DbRepository
}

type LinkWriter interface {
	Insert(ctx context.Context, tl *Tinylink) error
	Update(ctx context.Context, tl *Tinylink) error
	Delete(ctx context.Context, userID uint64, alias string) error
}

type RedisRepository interface {
	Redirect(ctx context.Context, alias string) (*RedirectValue, error)
	Cache(ctx context.Context, value RedirectValue, ttl time.Duration) error
	GenerateAlias(ctx context.Context) (string, error)
}
